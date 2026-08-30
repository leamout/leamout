package licensing

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

const signingAlgorithm = "Ed25519"

type Signer struct {
	keyID string
	key   ed25519.PrivateKey
}

func NewSigner(keyID string, key ed25519.PrivateKey) (*Signer, error) {
	if keyID == "" {
		return nil, fmt.Errorf("signing key ID is required")
	}
	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key")
	}
	return &Signer{keyID: keyID, key: append(ed25519.PrivateKey(nil), key...)}, nil
}

func (s *Signer) Sign(claims Claims) (Document, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return Document{}, fmt.Errorf("marshal license claims: %w", err)
	}
	return Document{
		Algorithm: signingAlgorithm,
		KeyID:     s.keyID,
		Payload:   base64.RawURLEncoding.EncodeToString(payload),
		Signature: base64.RawURLEncoding.EncodeToString(ed25519.Sign(s.key, payload)),
	}, nil
}
