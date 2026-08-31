package licensing

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSignerRoundTripV1(t *testing.T) {
	t.Parallel()

	privateKey := testPrivateKey(1)
	signer, err := NewSigner("key-2026-01", privateKey)
	if err != nil {
		t.Fatalf("NewSigner() error = %v", err)
	}
	keyring, err := NewKeyring(map[string]ed25519.PublicKey{
		signer.KeyID(): privateKey.Public().(ed25519.PublicKey),
	})
	if err != nil {
		t.Fatalf("NewKeyring() error = %v", err)
	}

	issuedAt := time.Date(2026, 8, 31, 20, 30, 0, 0, time.UTC)
	claims := LicenseClaimsV1{
		LicenseID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrganizationID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		SubscriptionID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		DeploymentID:   "node-01",
		IssuedAt:       issuedAt,
		ExpiresAt:      issuedAt.Add(24 * time.Hour),
		Features: map[string]bool{
			"recording.enabled": true,
			"byoc.enabled":      true,
		},
		Limits: map[string]int64{
			"max.concurrent_calls": 500,
			"max.deployments":      3,
		},
	}

	artifact, err := signer.SignV1(claims)
	if err != nil {
		t.Fatalf("SignV1() error = %v", err)
	}
	verified, err := keyring.VerifyV1(artifact, "node-01", issuedAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("VerifyV1() error = %v", err)
	}
	if verified.LicenseID != claims.LicenseID || verified.OrganizationID != claims.OrganizationID || verified.SubscriptionID != claims.SubscriptionID {
		t.Fatalf("verified identity = %#v, want %#v", verified, claims)
	}
	if !verified.Features["recording.enabled"] || verified.Limits["max.concurrent_calls"] != 500 {
		t.Fatalf("verified entitlements = %#v / %#v", verified.Features, verified.Limits)
	}
}

func TestSignerV1IsDeterministicForNormalizedClaims(t *testing.T) {
	t.Parallel()

	signer, err := NewSigner("key-1", testPrivateKey(2))
	if err != nil {
		t.Fatalf("NewSigner() error = %v", err)
	}
	issuedAt := time.Date(2026, 8, 31, 20, 30, 0, 999, time.UTC)
	claims := LicenseClaimsV1{
		LicenseID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrganizationID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		SubscriptionID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		DeploymentID:   " node-01 ",
		IssuedAt:       issuedAt,
		ExpiresAt:      issuedAt.Add(time.Hour),
		Features:       map[string]bool{"z.feature": true, "a.feature": false},
		Limits:         map[string]int64{"z.limit": 9, "a.limit": 1},
	}

	first, err := signer.SignV1(claims)
	if err != nil {
		t.Fatalf("first SignV1() error = %v", err)
	}
	second, err := signer.SignV1(claims)
	if err != nil {
		t.Fatalf("second SignV1() error = %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("SignV1() produced non-deterministic artifacts\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestVerifyV1RejectsTamperingAndWrongDeployment(t *testing.T) {
	t.Parallel()

	privateKey := testPrivateKey(3)
	signer, err := NewSigner("key-1", privateKey)
	if err != nil {
		t.Fatalf("NewSigner() error = %v", err)
	}
	keyring, err := NewKeyring(map[string]ed25519.PublicKey{"key-1": privateKey.Public().(ed25519.PublicKey)})
	if err != nil {
		t.Fatalf("NewKeyring() error = %v", err)
	}
	issuedAt := time.Date(2026, 8, 31, 20, 30, 0, 0, time.UTC)
	artifact, err := signer.SignV1(validTestClaims(issuedAt))
	if err != nil {
		t.Fatalf("SignV1() error = %v", err)
	}

	if _, err := keyring.VerifyV1(artifact, "node-02", issuedAt.Add(time.Minute)); !errors.Is(err, ErrDeploymentMismatch) {
		t.Fatalf("VerifyV1() error = %v, want %v", err, ErrDeploymentMismatch)
	}

	var envelope SignedLicenseV1
	if err := json.Unmarshal(artifact, &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	envelope.Payload = envelope.Payload[:len(envelope.Payload)-1] + "A"
	tampered, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if _, err := keyring.VerifyV1(tampered, "node-01", issuedAt.Add(time.Minute)); !errors.Is(err, ErrInvalidSignature) && !errors.Is(err, ErrMalformedArtifact) {
		t.Fatalf("VerifyV1() error = %v, want tamper rejection", err)
	}
}

func TestVerifyV1EnforcesValidityWindowAndKeyRotation(t *testing.T) {
	t.Parallel()

	oldPrivate := testPrivateKey(4)
	newPrivate := testPrivateKey(5)
	oldSigner, err := NewSigner("key-old", oldPrivate)
	if err != nil {
		t.Fatalf("NewSigner(old) error = %v", err)
	}
	issuedAt := time.Date(2026, 8, 31, 20, 30, 0, 0, time.UTC)
	artifact, err := oldSigner.SignV1(validTestClaims(issuedAt))
	if err != nil {
		t.Fatalf("SignV1() error = %v", err)
	}

	keyring, err := NewKeyring(map[string]ed25519.PublicKey{
		"key-old": oldPrivate.Public().(ed25519.PublicKey),
		"key-new": newPrivate.Public().(ed25519.PublicKey),
	})
	if err != nil {
		t.Fatalf("NewKeyring() error = %v", err)
	}
	if _, err := keyring.VerifyV1(artifact, "node-01", issuedAt.Add(-time.Second)); !errors.Is(err, ErrArtifactNotYetValid) {
		t.Fatalf("VerifyV1(before issuance) error = %v, want %v", err, ErrArtifactNotYetValid)
	}
	if _, err := keyring.VerifyV1(artifact, "node-01", issuedAt.Add(2*time.Hour)); !errors.Is(err, ErrArtifactExpired) {
		t.Fatalf("VerifyV1(at expiry) error = %v, want %v", err, ErrArtifactExpired)
	}

	newOnly, err := NewKeyring(map[string]ed25519.PublicKey{"key-new": newPrivate.Public().(ed25519.PublicKey)})
	if err != nil {
		t.Fatalf("NewKeyring(new only) error = %v", err)
	}
	if _, err := newOnly.VerifyV1(artifact, "node-01", issuedAt.Add(time.Minute)); !errors.Is(err, ErrSigningKeyUnavailable) {
		t.Fatalf("VerifyV1() error = %v, want %v", err, ErrSigningKeyUnavailable)
	}
}

func validTestClaims(issuedAt time.Time) LicenseClaimsV1 {
	return LicenseClaimsV1{
		LicenseID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		OrganizationID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		SubscriptionID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		DeploymentID:   "node-01",
		IssuedAt:       issuedAt,
		ExpiresAt:      issuedAt.Add(2 * time.Hour),
		Features:       map[string]bool{"recording.enabled": true},
		Limits:         map[string]int64{"max.deployments": 3},
	}
}

func testPrivateKey(marker byte) ed25519.PrivateKey {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = marker + byte(index)
	}
	return ed25519.NewKeyFromSeed(seed)
}
