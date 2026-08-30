package licensing

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/leamout/leamout/internal/commercial/entitlements"
)

const DocumentVersion = 1

var (
	ErrInvalidDocument  = errors.New("invalid license document")
	ErrInvalidSignature = errors.New("invalid license signature")
	ErrLicenseExpired   = errors.New("license has expired")
	ErrLicenseNotActive = errors.New("license is not active")
)

// EntitlementClaim is the stable wire representation embedded in an offline
// license. Pointer values preserve the distinction between zero and absent.
type EntitlementClaim struct {
	Key     string            `json:"key"`
	Kind    entitlements.Kind `json:"kind"`
	Enabled *bool             `json:"enabled,omitempty"`
	Limit   *int64            `json:"limit,omitempty"`
}

// Claims are the signed, provider-independent authority consumed by a
// customer-owned deployment. Times are encoded as UTC RFC 3339 JSON values.
type Claims struct {
	Version        int                `json:"version"`
	LicenseID      string             `json:"license_id"`
	OrganizationID string             `json:"organization_id"`
	SubscriptionID string             `json:"subscription_id,omitempty"`
	MaxDeployments int                `json:"max_deployments"`
	IssuedAt       time.Time          `json:"issued_at"`
	ExpiresAt      *time.Time         `json:"expires_at,omitempty"`
	Entitlements   []EntitlementClaim `json:"entitlements"`
}

// Document is the portable signed license envelope. Payload and Signature use
// unpadded URL-safe base64 so the document can be moved through environment
// variables and configuration files without further escaping.
type Document struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

// Signer issues Ed25519 license documents with one private signing key.
type Signer struct {
	keyID string
	key   ed25519.PrivateKey
}

func NewSigner(keyID string, key ed25519.PrivateKey) (*Signer, error) {
	if keyID == "" {
		return nil, fmt.Errorf("key_id is required")
	}
	if len(key) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid Ed25519 private key length")
	}
	return &Signer{keyID: keyID, key: key}, nil
}

// Sign validates and signs claims without mutating the caller's value.
func (s *Signer) Sign(claims Claims) (Document, error) {
	claims.IssuedAt = claims.IssuedAt.UTC()
	if claims.ExpiresAt != nil {
		expiresAt := claims.ExpiresAt.UTC()
		claims.ExpiresAt = &expiresAt
	}
	if err := ValidateClaims(claims); err != nil {
		return Document{}, err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return Document{}, fmt.Errorf("marshal license claims: %w", err)
	}
	return Document{
		Algorithm: "Ed25519",
		KeyID:     s.keyID,
		Payload:   base64.RawURLEncoding.EncodeToString(payload),
		Signature: base64.RawURLEncoding.EncodeToString(ed25519.Sign(s.key, payload)),
	}, nil
}

// Marshal serializes a document for transport or storage.
func (d Document) Marshal() ([]byte, error) {
	return json.Marshal(d)
}
