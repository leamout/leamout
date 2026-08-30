package licensing

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Keyring verifies documents issued by currently trusted signing keys. Keeping
// multiple public keys permits signing-key rotation without private keys on a
// customer-owned deployment.
type Keyring struct {
	keys map[string]ed25519.PublicKey
}

func NewKeyring(keys map[string]ed25519.PublicKey) (*Keyring, error) {
	copyKeys := make(map[string]ed25519.PublicKey, len(keys))
	for keyID, key := range keys {
		if keyID == "" {
			return nil, fmt.Errorf("key_id is required")
		}
		if len(key) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("invalid Ed25519 public key length for %q", keyID)
		}
		copyKeys[keyID] = append(ed25519.PublicKey(nil), key...)
	}
	return &Keyring{keys: copyKeys}, nil
}

// Verify authenticates a document, validates its claims, and applies current
// lifecycle checks. The caller supplies time to keep policy deterministic and
// testable; clock-skew allowance belongs at the application boundary.
func (k *Keyring) Verify(document Document, at time.Time) (Claims, error) {
	if document.Algorithm != "Ed25519" || document.KeyID == "" {
		return Claims{}, ErrInvalidDocument
	}
	key, ok := k.keys[document.KeyID]
	if !ok {
		return Claims{}, fmt.Errorf("%w: unknown signing key %q", ErrInvalidSignature, document.KeyID)
	}
	payload, err := base64.RawURLEncoding.DecodeString(document.Payload)
	if err != nil {
		return Claims{}, fmt.Errorf("%w: decode payload", ErrInvalidDocument)
	}
	signature, err := base64.RawURLEncoding.DecodeString(document.Signature)
	if err != nil || !ed25519.Verify(key, payload, signature) {
		return Claims{}, ErrInvalidSignature
	}
	var claims Claims
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil {
		return Claims{}, fmt.Errorf("%w: decode claims", ErrInvalidDocument)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Claims{}, fmt.Errorf("%w: trailing claims content", ErrInvalidDocument)
	}
	if err := ValidateClaims(claims); err != nil {
		return Claims{}, err
	}
	if at.Before(claims.IssuedAt) {
		return Claims{}, ErrLicenseNotActive
	}
	if claims.ExpiresAt != nil && !at.Before(*claims.ExpiresAt) {
		return Claims{}, ErrLicenseExpired
	}
	return claims, nil
}

// ParseDocument decodes the transport representation and rejects unknown
// fields so misspelled or future envelope fields fail closed.
func ParseDocument(encoded []byte) (Document, error) {
	var document Document
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return Document{}, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Document{}, fmt.Errorf("%w: trailing content", ErrInvalidDocument)
	}
	return document, nil
}
