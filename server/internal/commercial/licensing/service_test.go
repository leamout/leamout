package licensing

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func signed(t *testing.T, claims Claims, private ed25519.PrivateKey) []byte {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(Document{
		KeyID:     "test",
		Payload:   base64.StdEncoding.EncodeToString(payload),
		Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(private, payload)),
	})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLicenseLifecycleAndAtomicRenewal(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	claims := Claims{
		Version:            1,
		LicenseID:          "lic-1",
		CustomerID:         "customer",
		DeploymentID:       "deployment",
		ValidFrom:          now,
		ExpiresAt:          now.Add(time.Hour),
		GraceUntil:         now.Add(2 * time.Hour),
		MaxConcurrentCalls: 50,
		MaxNodes:           3,
		Features: []string{
			"sso",
			"audit_logging",
		},
	}
	path := filepath.Join(t.TempDir(), "license.json")
	config := Config{
		Mode:         SelfHosted,
		DeploymentID: "deployment",
		File:         path,
		PublicKeys: map[string]string{
			"test": base64.StdEncoding.EncodeToString(public),
		},
	}
	service, err := NewService(config)
	if err != nil {
		t.Fatal(err)
	}
	service.now = func() time.Time {
		return now
	}
	if got := service.Status().State; got != "missing" {
		t.Fatalf("state = %s", got)
	}
	if _, err := service.RequireRuntime(); err == nil {
		t.Fatal("missing license authorized runtime")
	}
	if err := os.WriteFile(path, signed(t, claims, private), 0600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		offset  time.Duration
		state   string
		allowed bool
	}{
		{
			-time.Nanosecond,
			"not_yet_valid",
			false,
		},
		{
			0,
			"active",
			true,
		},
		{
			time.Hour,
			"grace",
			true,
		},
		{
			2 * time.Hour,
			"expired",
			false,
		},
	} {
		t.Run(tc.state, func(t *testing.T) {
			service.now = func() time.Time {
				return now.Add(tc.offset)
			}
			if got := service.Status().State; got != tc.state {
				t.Fatalf("state=%s, want=%s", got, tc.state)
			}
			auditErr := service.RequireDataFeature("audit_logging")
			if (auditErr == nil) != (tc.state != "not_yet_valid") {
				t.Fatalf("audit access error=%v", auditErr)
			}
			if service.RequireDataFeature("premium_templates") == nil {
				t.Fatal("unlicensed data feature allowed")
			}
			_, err := service.RequireRuntime()
			if (err == nil) != tc.allowed {
				t.Fatalf("runtime error=%v", err)
			}
		})
	}
	claims.ExpiresAt = now.Add(3 * time.Hour)
	claims.GraceUntil = claims.ExpiresAt
	replacement := path + ".new"
	if err := os.WriteFile(replacement, signed(t, claims, private), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, path); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequireRuntime(); err != nil {
		t.Fatalf("renewal: %v", err)
	}
	if enabled, err := service.Enabled(context.Background(), "sso"); err != nil || !enabled {
		t.Fatalf("SSO=%v, %v", enabled, err)
	}
	if enabled, err := service.Enabled(context.Background(), "scim"); err != nil || enabled {
		t.Fatalf("SCIM=%v, %v", enabled, err)
	}
	// A trust-key mutation by the caller must not change the running service.
	config.PublicKeys["test"] = "invalid"
	if _, err := service.RequireRuntime(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("tampered"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RequireRuntime(); err == nil {
		t.Fatal("tampered license authorized runtime")
	}
}

func TestVerifyRejectsUntrustedClaims(t *testing.T) {
	public, private, _ := ed25519.GenerateKey(rand.Reader)
	_, other, _ := ed25519.GenerateKey(rand.Reader)
	now := time.Now().UTC()
	valid := Claims{
		Version:            1,
		LicenseID:          "id",
		CustomerID:         "customer",
		DeploymentID:       "deployment",
		ValidFrom:          now,
		ExpiresAt:          now.Add(time.Hour),
		GraceUntil:         now.Add(time.Hour),
		MaxConcurrentCalls: 1,
		MaxNodes:           1,
	}
	keys := map[string]string{
		"test": base64.StdEncoding.EncodeToString(public),
	}
	for _, tc := range []struct {
		name       string
		edit       func(*Claims)
		key        ed25519.PrivateKey
		deployment string
	}{
		{
			"wrong signer",
			func(*Claims) {
			},
			other,
			"deployment",
		},
		{
			"wrong deployment",
			func(*Claims) {
			},
			private,
			"other",
		},
		{
			"unknown version",
			func(c *Claims) {
				c.Version = 2
			},
			private,
			"deployment",
		},
		{
			"zero capacity",
			func(c *Claims) {
				c.MaxConcurrentCalls = 0
			},
			private,
			"deployment",
		},
		{
			"unbounded grace",
			func(c *Claims) {
				c.GraceUntil = c.ExpiresAt.AddDate(0, 0, 31)
			},
			private,
			"deployment",
		},
		{
			"unknown feature",
			func(c *Claims) {
				c.Features = []string{
					"anything",
				}
			},
			private,
			"deployment",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := valid
			tc.edit(&c)
			if _, err := Verify(signed(t, c, tc.key), keys, tc.deployment); err == nil {
				t.Fatal("invalid license accepted")
			}
		})
	}
	data := signed(t, valid, private)
	if _, err := Verify(append(data, []byte("{}")...), keys, "deployment"); err == nil {
		t.Fatal("trailing data accepted")
	}
}

func TestDeploymentConfig(t *testing.T) {
	for _, config := range []Config{
		{
			Mode: "unknown",
		},
		{
			Mode: SelfHosted,
		},
		{
			Mode:         SelfHosted,
			DeploymentID: "deployment",
			File:         "license",
			PublicKeys: map[string]string{
				"bad": "bad",
			},
		},
	} {
		if _, err := NewService(config); err == nil {
			t.Fatalf("invalid config accepted: %+v", config)
		}
	}
	cloud, err := NewService(Config{})
	if err != nil || cloud.SelfHosted() || cloud.Status().State != "cloud" {
		t.Fatalf("Cloud default: %v", err)
	}
}

func TestPinnedKeyringCannotBeReplacedByEnvironment(t *testing.T) {
	public, _, _ := ed25519.GenerateKey(rand.Reader)
	original := BuildPublicKeys
	t.Cleanup(func() {
		BuildPublicKeys = original
	})
	BuildPublicKeys = "issuer:" + base64.StdEncoding.EncodeToString(public)
	t.Setenv("LICENSE_PUBLIC_KEYS", "attacker:fake")
	keys, err := (Config{}).keyring()
	if err != nil || keys["issuer"] == "" || keys["attacker"] != "" {
		t.Fatalf("keyring=%v err=%v", keys, err)
	}
	BuildPublicKeys = "issuer:bad,issuer:bad"
	if err := (Config{}).Validate(); err == nil {
		t.Fatal("duplicate build key accepted")
	}
}
