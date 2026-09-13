package licensing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
)

const LicenseClaimsVersionV1 = 1

var (
	ErrClaimsIssuedAtRequired  = errors.New("issued_at is required in signed license claims")
	ErrClaimsExpiresAtRequired = errors.New("expires_at is required in signed license claims")
	ErrMalformedClaims         = errors.New("malformed signed license claims")
)

// LicenseClaimsV1 is the versioned commercial authorization carried by a
// self-hosted deployment artifact. ExpiresAt is the artifact validity boundary
// and may be shorter than the durable database license expiration.
type LicenseClaimsV1 struct {
	LicenseID           uuid.UUID
	OrganizationID      uuid.UUID
	DeploymentID        string
	DeploymentPublicKey string
	IssuedAt            time.Time
	ExpiresAt           time.Time
}

type licenseClaimsWireV1 struct {
	Version             int    `json:"version"`
	LicenseID           string `json:"license_id"`
	OrganizationID      string `json:"organization_id"`
	DeploymentID        string `json:"deployment_id"`
	DeploymentPublicKey string `json:"deployment_public_key"`
	IssuedAt            int64  `json:"issued_at"`
	ExpiresAt           int64  `json:"expires_at"`
}

func normalizeClaimsV1(claims LicenseClaimsV1) (LicenseClaimsV1, error) {
	if err := validateID(claims.LicenseID, ErrLicenseIDRequired); err != nil {
		return LicenseClaimsV1{}, err
	}
	if err := validateID(claims.OrganizationID, ErrOrganizationIDRequired); err != nil {
		return LicenseClaimsV1{}, err
	}
	claims.DeploymentID = strings.TrimSpace(claims.DeploymentID)
	if claims.DeploymentID == "" {
		return LicenseClaimsV1{}, ErrDeploymentIDRequired
	}
	if strings.IndexFunc(claims.DeploymentID, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
		return LicenseClaimsV1{}, ErrInvalidDeploymentID
	}
	publicKey, err := normalizeDeploymentPublicKey(claims.DeploymentPublicKey)
	if err != nil {
		return LicenseClaimsV1{}, err
	}
	claims.DeploymentPublicKey = publicKey
	if claims.IssuedAt.IsZero() {
		return LicenseClaimsV1{}, ErrClaimsIssuedAtRequired
	}
	if claims.ExpiresAt.IsZero() {
		return LicenseClaimsV1{}, ErrClaimsExpiresAtRequired
	}

	claims.IssuedAt = claims.IssuedAt.UTC().Truncate(time.Second)
	claims.ExpiresAt = claims.ExpiresAt.UTC().Truncate(time.Second)
	if !claims.ExpiresAt.After(claims.IssuedAt) {
		return LicenseClaimsV1{}, ErrInvalidExpiration
	}
	return claims, nil
}

func marshalClaimsV1(claims LicenseClaimsV1) ([]byte, LicenseClaimsV1, error) {
	normalized, err := normalizeClaimsV1(claims)
	if err != nil {
		return nil, LicenseClaimsV1{}, err
	}

	payload, err := json.Marshal(licenseClaimsWireV1{
		Version:             LicenseClaimsVersionV1,
		LicenseID:           normalized.LicenseID.String(),
		OrganizationID:      normalized.OrganizationID.String(),
		DeploymentID:        normalized.DeploymentID,
		DeploymentPublicKey: normalized.DeploymentPublicKey,
		IssuedAt:            normalized.IssuedAt.Unix(),
		ExpiresAt:           normalized.ExpiresAt.Unix(),
	})
	if err != nil {
		return nil, LicenseClaimsV1{}, err
	}
	return payload, normalized, nil
}

func unmarshalClaimsV1(payload []byte) (LicenseClaimsV1, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	var wire licenseClaimsWireV1
	if err := decoder.Decode(&wire); err != nil {
		return LicenseClaimsV1{}, ErrMalformedClaims
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return LicenseClaimsV1{}, ErrMalformedClaims
	}
	if wire.Version != LicenseClaimsVersionV1 {
		return LicenseClaimsV1{}, ErrUnsupportedLicenseVersion
	}

	licenseID, err := uuid.Parse(wire.LicenseID)
	if err != nil {
		return LicenseClaimsV1{}, ErrMalformedClaims
	}
	organizationID, err := uuid.Parse(wire.OrganizationID)
	if err != nil {
		return LicenseClaimsV1{}, ErrMalformedClaims
	}

	claims, err := normalizeClaimsV1(LicenseClaimsV1{
		LicenseID:           licenseID,
		OrganizationID:      organizationID,
		DeploymentID:        wire.DeploymentID,
		DeploymentPublicKey: wire.DeploymentPublicKey,
		IssuedAt:            time.Unix(wire.IssuedAt, 0).UTC(),
		ExpiresAt:           time.Unix(wire.ExpiresAt, 0).UTC(),
	})
	if err != nil {
		return LicenseClaimsV1{}, err
	}
	return claims, nil
}
