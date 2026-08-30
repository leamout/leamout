package licensing

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/leamout/leamout/internal/commercial/entitlements"
)

func TestSignerAndKeyringRoundTrip(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := NewSigner("2026-08", privateKey)
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := NewKeyring(map[string]ed25519.PublicKey{"2026-08": publicKey})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.FixedZone("test", 3600))
	expiresAt := now.Add(24 * time.Hour)
	enabled := true
	limit := int64(20)
	document, err := signer.Sign(Claims{
		Version:        DocumentVersion,
		LicenseID:      "license-1",
		OrganizationID: "organization-1",
		MaxDeployments: 2,
		IssuedAt:       now,
		ExpiresAt:      &expiresAt,
		Entitlements: []EntitlementClaim{
			{Key: "voice.recording", Kind: entitlements.KindFeature, Enabled: &enabled},
			{Key: "voice.concurrent_calls", Kind: entitlements.KindLimit, Limit: &limit},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := keyring.Verify(document, now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if claims.OrganizationID != "organization-1" || claims.IssuedAt.Location() != time.UTC {
		t.Fatalf("unexpected verified claims: %+v", claims)
	}

	encoded, err := document.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseDocument(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if parsed != document {
		t.Fatalf("parsed document does not match issued document")
	}
}

func TestKeyringRejectsTamperingUnknownKeysAndExpiration(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, _ := NewSigner("active", privateKey)
	keyring, _ := NewKeyring(map[string]ed25519.PublicKey{"active": publicKey})
	now := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(time.Hour)
	document, err := signer.Sign(Claims{
		Version: DocumentVersion, LicenseID: "license-1", OrganizationID: "organization-1",
		MaxDeployments: 1, IssuedAt: now, ExpiresAt: &expiresAt,
	})
	if err != nil {
		t.Fatal(err)
	}

	tampered := document
	if tampered.Signature[0] == 'A' {
		tampered.Signature = "B" + tampered.Signature[1:]
	} else {
		tampered.Signature = "A" + tampered.Signature[1:]
	}
	if _, err := keyring.Verify(tampered, now); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected invalid signature, got %v", err)
	}

	unknown := document
	unknown.KeyID = "retired"
	if _, err := keyring.Verify(unknown, now); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected unknown key failure, got %v", err)
	}

	if _, err := keyring.Verify(document, now.Add(-time.Nanosecond)); !errors.Is(err, ErrLicenseNotActive) {
		t.Fatalf("expected not active failure, got %v", err)
	}

	if _, err := keyring.Verify(document, expiresAt); !errors.Is(err, ErrLicenseExpired) {
		t.Fatalf("expected expiration failure, got %v", err)
	}
}

func TestValidateClaimsRejectsDuplicateEntitlements(t *testing.T) {
	enabled := true
	claims := Claims{
		Version: DocumentVersion, LicenseID: "license-1", OrganizationID: "organization-1",
		MaxDeployments: 1, IssuedAt: time.Now(),
		Entitlements: []EntitlementClaim{
			{Key: "voice.recording", Kind: entitlements.KindFeature, Enabled: &enabled},
			{Key: "voice.recording", Kind: entitlements.KindFeature, Enabled: &enabled},
		},
	}
	if err := ValidateClaims(claims); !errors.Is(err, ErrInvalidClaims) {
		t.Fatalf("expected invalid claims, got %v", err)
	}
}
