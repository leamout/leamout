package licensing

import (
	"crypto/ed25519"
	"fmt"
)

// Keyring owns the trusted signing keys used by the commercial licensing
// authority. It is not constructed in self-hosted runtimes.
type Keyring struct {
	signers map[string]*Signer
}

func NewKeyring(keys map[string]ed25519.PrivateKey) (*Keyring, error) {
	signers := make(map[string]*Signer, len(keys))
	for keyID, key := range keys {
		signer, err := NewSigner(keyID, key)
		if err != nil {
			return nil, fmt.Errorf("configure signing key %q: %w", keyID, err)
		}
		signers[keyID] = signer
	}
	return &Keyring{signers: signers}, nil
}

func (k *Keyring) Signer(keyID string) (*Signer, error) {
	if k == nil {
		return nil, fmt.Errorf("signing key %q is unavailable", keyID)
	}
	signer, ok := k.signers[keyID]
	if !ok {
		return nil, fmt.Errorf("signing key %q is unavailable", keyID)
	}
	return signer, nil
}
