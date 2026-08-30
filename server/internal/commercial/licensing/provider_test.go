package licensing

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
)

func TestStaticKeyProviderSelectsByKeyID(t *testing.T) {
	_, first, _ := ed25519.GenerateKey(rand.Reader)
	_, second, _ := ed25519.GenerateKey(rand.Reader)
	provider, err := NewStaticKeyProvider(map[string]ed25519.PrivateKey{"first": first, "second": second})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Signer(context.Background(), "second"); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Signer(context.Background(), "missing"); !errors.Is(err, ErrSigningKeyNotFound) {
		t.Fatalf("expected missing key error, got %v", err)
	}
}
