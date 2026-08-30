package licensing

import (
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
)

var ErrSigningKeyNotFound = errors.New("signing key not found")

// DocumentSigner permits production providers to sign through a KMS or HSM
// without exposing private-key bytes to the issuance service.
type DocumentSigner interface {
	Sign(Claims) (Document, error)
}

// SigningKeyProvider selects a signer by the immutable key ID stored on a
// license. Implementations may be backed by a KMS, HSM, or local development
// keyring.
type SigningKeyProvider interface {
	Signer(context.Context, string) (DocumentSigner, error)
}

// StaticKeyProvider is intended for tests and local development. Production
// deployments should provide an implementation whose keys are non-exportable.
type StaticKeyProvider struct {
	signers map[string]DocumentSigner
}

func NewStaticKeyProvider(keys map[string]ed25519.PrivateKey) (*StaticKeyProvider, error) {
	signers := make(map[string]DocumentSigner, len(keys))
	for keyID, key := range keys {
		keyCopy := append(ed25519.PrivateKey(nil), key...)
		signer, err := NewSigner(keyID, keyCopy)
		if err != nil {
			return nil, fmt.Errorf("configure signing key %q: %w", keyID, err)
		}
		signers[keyID] = signer
	}
	return &StaticKeyProvider{signers: signers}, nil
}

func (p *StaticKeyProvider) Signer(_ context.Context, keyID string) (DocumentSigner, error) {
	if p == nil {
		return nil, ErrSigningKeyNotFound
	}
	signer, ok := p.signers[keyID]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrSigningKeyNotFound, keyID)
	}
	return signer, nil
}
