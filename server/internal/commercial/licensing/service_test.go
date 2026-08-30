package licensing

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leamout/leamout/internal/database/sqlc"
)

func TestBuildClaimsUsesLicenseEntitlementSnapshot(t *testing.T) {
	now := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	expiresAt := now.Add(24 * time.Hour)
	trueValue, falseValue := true, false
	limit := int64(100)
	license := sqlc.License{
		ID: uuid.New(), OrganizationID: uuid.New(), Status: string(StatusActive), MaxDeployments: 2,
		IssuedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	}
	entitlements := []sqlc.Entitlement{
		{EntitlementKey: "voice.enabled", Kind: "feature", Enabled: &trueValue},
		{EntitlementKey: "recording.enabled", Kind: "feature", Enabled: &falseValue},
		{EntitlementKey: "max.concurrent.calls", Kind: "limit", LimitValue: &limit},
		{EntitlementKey: "expired.feature", Kind: "feature", Enabled: &trueValue, ExpiresAt: pgtype.Timestamptz{Time: now, Valid: true}},
	}

	claims, err := buildClaims(license, entitlements, now)
	if err != nil {
		t.Fatal(err)
	}
	if !claims.Features["voice.enabled"] || claims.Features["recording.enabled"] {
		t.Fatalf("unexpected feature snapshot: %v", claims.Features)
	}
	if claims.Limits["max.concurrent.calls"] != 100 {
		t.Fatalf("unexpected limit snapshot: %v", claims.Limits)
	}
	if _, ok := claims.Features["expired.feature"]; ok {
		t.Fatal("expired entitlement was included")
	}
}

func TestValidateActiveLicense(t *testing.T) {
	now := time.Now().UTC()
	keyID := "current"
	valid := sqlc.License{
		Status: string(StatusActive), MaxDeployments: 1, SigningKeyID: &keyID,
		IssuedAt: pgtype.Timestamptz{Time: now.Add(-time.Hour), Valid: true},
	}
	if err := validateActiveLicense(valid, now); err != nil {
		t.Fatal(err)
	}

	pending := valid
	pending.Status = string(StatusPending)
	if err := validateActiveLicense(pending, now); !errors.Is(err, ErrLicenseNotActive) {
		t.Fatalf("expected inactive error, got %v", err)
	}
	expired := valid
	expired.ExpiresAt = pgtype.Timestamptz{Time: now, Valid: true}
	if err := validateActiveLicense(expired, now); !errors.Is(err, ErrLicenseExpired) {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestKeyringSelectsConfiguredSignerAndProducesVerifiableDocument(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := NewKeyring(map[string]ed25519.PrivateKey{"current": privateKey})
	if err != nil {
		t.Fatal(err)
	}
	signer, err := keyring.Signer("current")
	if err != nil {
		t.Fatal(err)
	}
	claims := Claims{
		Version: DocumentVersion, OrganizationID: uuid.New(), LicenseID: uuid.New(),
		IssuedAt: time.Now().UTC(), MaxDeployments: 1,
		Features: map[string]bool{"voice.enabled": true}, Limits: map[string]int64{},
	}
	document, err := signer.Sign(claims)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := base64.RawURLEncoding.DecodeString(document.Payload)
	if err != nil {
		t.Fatal(err)
	}
	signature, err := base64.RawURLEncoding.DecodeString(document.Signature)
	if err != nil {
		t.Fatal(err)
	}
	if !ed25519.Verify(publicKey, payload, signature) {
		t.Fatal("document signature did not verify")
	}
	var decoded Claims
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.LicenseID != claims.LicenseID {
		t.Fatalf("decoded license ID = %s, want %s", decoded.LicenseID, claims.LicenseID)
	}
	if _, err := keyring.Signer("missing"); err == nil {
		t.Fatal("expected missing signing key error")
	}
}
