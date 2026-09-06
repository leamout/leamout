package licensing

import (
	"strings"
	"testing"
)

func TestGenerateDeploymentToken(t *testing.T) {
	token, prefix, hash, err := generateDeploymentToken()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(token, deploymentTokenPrefix) {
		t.Fatalf("token = %q", token)
	}
	if !strings.HasPrefix(token, prefix+"_") {
		t.Fatalf("token %q does not contain prefix %q", token, prefix)
	}
	if parsed, err := parseDeploymentTokenPrefix(token); err != nil || parsed != prefix {
		t.Fatalf("parsed prefix = %q, err = %v", parsed, err)
	}
	if hash == "" || hash != hashDeploymentToken(token) {
		t.Fatalf("hash = %q", hash)
	}
}

func TestGenerateDeploymentTokenProducesDistinctCredentials(t *testing.T) {
	first, _, firstHash, err := generateDeploymentToken()
	if err != nil {
		t.Fatal(err)
	}
	second, _, secondHash, err := generateDeploymentToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || firstHash == secondHash {
		t.Fatal("expected independently generated deployment credentials")
	}
}

func TestParseDeploymentTokenPrefixRejectsOrganizationToken(t *testing.T) {
	if _, err := parseDeploymentTokenPrefix("lm_org_12345678_secret"); err == nil {
		t.Fatal("expected organization token to be rejected")
	}
}

func TestParseDeploymentTokenPrefixRejectsMalformedCredential(t *testing.T) {
	for _, token := range []string{
		"",
		"lm_dep_",
		"lm_dep_12345678",
		"lm_dep_12345678_",
	} {
		if _, err := parseDeploymentTokenPrefix(token); err == nil {
			t.Fatalf("expected %q to be rejected", token)
		}
	}
}

func TestHasScope(t *testing.T) {
	scopes := []string{ManagedCarrierTransitScope}
	if !hasScope(scopes, ManagedCarrierTransitScope) {
		t.Fatal("expected managed carrier transit scope")
	}
	if hasScope(scopes, "managed_carrier:admin") {
		t.Fatal("unexpected unrelated scope")
	}
}
