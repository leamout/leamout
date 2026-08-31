package licensing

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestVerifyV1AuthenticatesKeyID(t *testing.T) {
	t.Parallel()

	privateKey := testPrivateKey(7)
	signer, err := NewSigner("key-a", privateKey)
	if err != nil {
		t.Fatalf("NewSigner() error = %v", err)
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	keyring, err := NewKeyring(map[string]ed25519.PublicKey{
		"key-a": publicKey,
		"key-b": publicKey,
	})
	if err != nil {
		t.Fatalf("NewKeyring() error = %v", err)
	}
	issuedAt := time.Date(2026, 8, 31, 20, 30, 0, 0, time.UTC)
	artifact, err := signer.SignV1(validTestClaims(issuedAt))
	if err != nil {
		t.Fatalf("SignV1() error = %v", err)
	}

	var envelope SignedLicenseV1
	if err := json.Unmarshal(artifact, &envelope); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	envelope.KeyID = "key-b"
	tampered, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if _, err := keyring.VerifyV1(tampered, "node-01", issuedAt.Add(time.Minute)); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("VerifyV1() error = %v, want %v", err, ErrInvalidSignature)
	}
}

func TestNormalizeClaimsV1RejectsFeatureLimitKeyCollision(t *testing.T) {
	t.Parallel()

	issuedAt := time.Date(2026, 8, 31, 20, 30, 0, 0, time.UTC)
	claims := validTestClaims(issuedAt)
	claims.Features["shared.key"] = true
	claims.Limits["shared.key"] = 1

	if _, err := normalizeClaimsV1(claims); !errors.Is(err, ErrDuplicateClaimKey) {
		t.Fatalf("normalizeClaimsV1() error = %v, want %v", err, ErrDuplicateClaimKey)
	}
}
