package licensing

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
	"unicode"
)

const (
	SignedLicenseVersionV1 = 1
	SignatureAlgorithmV1   = "Ed25519"
)

var signatureDomainV1 = []byte("leamout-license-v1\x00")

// SignedLicenseV1 is the transport envelope consumed by self-hosted runtimes.
// Claims are encoded separately so the envelope can evolve independently.
type SignedLicenseV1 struct {
	Version   int    `json:"version"`
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}

// Signer signs deployment-bound license claims with one trusted Ed25519 key.
type Signer struct {
	keyID      string
	privateKey ed25519.PrivateKey
}

func NewSigner(keyID string, privateKey ed25519.PrivateKey) (*Signer, error) {
	keyID, err := normalizeSigningKeyID(keyID)
	if err != nil {
		return nil, err
	}
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, ErrSigningKeyUnavailable
	}
	keyCopy := append(ed25519.PrivateKey(nil), privateKey...)
	return &Signer{keyID: keyID, privateKey: keyCopy}, nil
}

func (s *Signer) KeyID() string {
	return s.keyID
}

func (s *Signer) SignV1(claims LicenseClaimsV1) ([]byte, error) {
	payload, _, err := marshalClaimsV1(claims)
	if err != nil {
		return nil, err
	}
	signature := ed25519.Sign(s.privateKey, signatureMessageV1(s.keyID, payload))
	return json.Marshal(SignedLicenseV1{
		Version:   SignedLicenseVersionV1,
		Algorithm: SignatureAlgorithmV1,
		KeyID:     s.keyID,
		Payload:   base64.RawURLEncoding.EncodeToString(payload),
		Signature: base64.RawURLEncoding.EncodeToString(signature),
	})
}

// Keyring is the public verification material shipped to self-hosted runtimes.
// Keeping multiple key IDs allows artifacts signed before a rotation to remain
// verifiable until their normal expiry.
type Keyring struct {
	keys map[string]ed25519.PublicKey
}

func NewKeyring(keys map[string]ed25519.PublicKey) (*Keyring, error) {
	result := &Keyring{keys: make(map[string]ed25519.PublicKey, len(keys))}
	for keyID, publicKey := range keys {
		normalized, err := normalizeSigningKeyID(keyID)
		if err != nil {
			return nil, err
		}
		if len(publicKey) != ed25519.PublicKeySize {
			return nil, ErrSigningKeyUnavailable
		}
		result.keys[normalized] = append(ed25519.PublicKey(nil), publicKey...)
	}
	return result, nil
}

func (k *Keyring) VerifyV1(artifact []byte, expectedDeploymentID string, at time.Time) (LicenseClaimsV1, error) {
	envelope, payload, signature, err := decodeArtifactV1(artifact)
	if err != nil {
		return LicenseClaimsV1{}, err
	}
	publicKey, ok := k.keys[envelope.KeyID]
	if !ok {
		return LicenseClaimsV1{}, ErrSigningKeyUnavailable
	}
	if !ed25519.Verify(publicKey, signatureMessageV1(envelope.KeyID, payload), signature) {
		return LicenseClaimsV1{}, ErrInvalidSignature
	}
	claims, err := unmarshalClaimsV1(payload)
	if err != nil {
		return LicenseClaimsV1{}, err
	}

	expected, err := normalizeDeployment(ActivateDeploymentInput{DeploymentID: expectedDeploymentID})
	if err != nil {
		return LicenseClaimsV1{}, err
	}
	if claims.DeploymentID != expected.DeploymentID {
		return LicenseClaimsV1{}, ErrDeploymentMismatch
	}
	at = at.UTC()
	if at.Before(claims.IssuedAt) {
		return LicenseClaimsV1{}, ErrArtifactNotYetValid
	}
	if !at.Before(claims.ExpiresAt) {
		return LicenseClaimsV1{}, ErrArtifactExpired
	}
	return claims, nil
}

func decodeArtifactV1(artifact []byte) (SignedLicenseV1, []byte, []byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(artifact))
	decoder.DisallowUnknownFields()
	var envelope SignedLicenseV1
	if err := decoder.Decode(&envelope); err != nil {
		return SignedLicenseV1{}, nil, nil, ErrMalformedArtifact
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return SignedLicenseV1{}, nil, nil, ErrMalformedArtifact
	}
	if envelope.Version != SignedLicenseVersionV1 {
		return SignedLicenseV1{}, nil, nil, ErrUnsupportedLicenseVersion
	}
	if envelope.Algorithm != SignatureAlgorithmV1 {
		return SignedLicenseV1{}, nil, nil, ErrUnsupportedAlgorithm
	}
	keyID, err := normalizeSigningKeyID(envelope.KeyID)
	if err != nil {
		return SignedLicenseV1{}, nil, nil, err
	}
	envelope.KeyID = keyID
	payload, err := base64.RawURLEncoding.DecodeString(envelope.Payload)
	if err != nil || len(payload) == 0 {
		return SignedLicenseV1{}, nil, nil, ErrMalformedArtifact
	}
	signature, err := base64.RawURLEncoding.DecodeString(envelope.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return SignedLicenseV1{}, nil, nil, ErrMalformedArtifact
	}
	return envelope, payload, signature, nil
}

func signatureMessageV1(keyID string, payload []byte) []byte {
	message := make([]byte, 0, len(signatureDomainV1)+len(keyID)+1+len(payload))
	message = append(message, signatureDomainV1...)
	message = append(message, keyID...)
	message = append(message, 0)
	message = append(message, payload...)
	return message
}

func normalizeSigningKeyID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrSigningKeyRequired
	}
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", ErrInvalidSigningKey
	}
	return value, nil
}
