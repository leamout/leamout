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

func TestParseDeploymentTokenPrefixRejectsOrganizationToken(t *testing.T) {
	if _, err := parseDeploymentTokenPrefix("lm_org_12345678_secret"); err == nil {
		t.Fatal("expected organization token to be rejected")
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
