// Package secrets provides compatibility wrappers for persisted secret
// encryption. New code should import internal/security/encryption directly.
package secrets

import "github.com/leamout/leamout/internal/security/encryption"

// Cipher is an alias for the authenticated encryption implementation.
type Cipher = encryption.Cipher

// New constructs a cipher from the configured base64url-encoded key.
func New(encodedKey string) (*Cipher, error) {
	return encryption.New(encodedKey)
}
