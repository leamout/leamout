package licensing

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

const LicenseClaimsVersionV1 = 1

var (
	ErrClaimsSubscriptionIDRequired = errors.New("subscription id is required in signed license claims")
	ErrClaimsIssuedAtRequired       = errors.New("issued_at is required in signed license claims")
	ErrClaimsExpiresAtRequired      = errors.New("expires_at is required in signed license claims")
	ErrInvalidClaimKey              = errors.New("license claim key must be non-empty and contain no whitespace")
	ErrInvalidClaimLimit            = errors.New("license claim limit must be non-negative")
	ErrDuplicateClaimKey            = errors.New("duplicate license claim key")
	ErrMalformedClaims              = errors.New("malformed signed license claims")
)

// LicenseClaimsV1 is the versioned commercial authorization carried by a
// self-hosted deployment artifact. ExpiresAt is the artifact validity boundary
// and may be shorter than the durable database license expiration.
type LicenseClaimsV1 struct {
	LicenseID      uuid.UUID
	OrganizationID uuid.UUID
	SubscriptionID uuid.UUID
	DeploymentID   string
	IssuedAt       time.Time
	ExpiresAt      time.Time
	Features       map[string]bool
	Limits         map[string]int64
}

type featureClaimV1 struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
}

type limitClaimV1 struct {
	Key   string `json:"key"`
	Value int64  `json:"value"`
}

type licenseClaimsWireV1 struct {
	Version        int              `json:"version"`
	LicenseID      string           `json:"license_id"`
	OrganizationID string           `json:"organization_id"`
	SubscriptionID string           `json:"subscription_id"`
	DeploymentID   string           `json:"deployment_id"`
	IssuedAt       int64            `json:"issued_at"`
	ExpiresAt      int64            `json:"expires_at"`
	Features       []featureClaimV1 `json:"features"`
	Limits         []limitClaimV1   `json:"limits"`
}

func normalizeClaimsV1(claims LicenseClaimsV1) (LicenseClaimsV1, error) {
	if err := validateID(claims.LicenseID, ErrLicenseIDRequired); err != nil {
		return LicenseClaimsV1{}, err
	}
	if err := validateID(claims.OrganizationID, ErrOrganizationIDRequired); err != nil {
		return LicenseClaimsV1{}, err
	}
	if claims.SubscriptionID == uuid.Nil {
		return LicenseClaimsV1{}, ErrClaimsSubscriptionIDRequired
	}
	deployment, err := normalizeDeployment(ActivateDeploymentInput{DeploymentID: claims.DeploymentID})
	if err != nil {
		return LicenseClaimsV1{}, err
	}
	if claims.IssuedAt.IsZero() {
		return LicenseClaimsV1{}, ErrClaimsIssuedAtRequired
	}
	if claims.ExpiresAt.IsZero() {
		return LicenseClaimsV1{}, ErrClaimsExpiresAtRequired
	}

	claims.DeploymentID = deployment.DeploymentID
	claims.IssuedAt = claims.IssuedAt.UTC().Truncate(time.Second)
	claims.ExpiresAt = claims.ExpiresAt.UTC().Truncate(time.Second)
	if !claims.ExpiresAt.After(claims.IssuedAt) {
		return LicenseClaimsV1{}, ErrInvalidExpiration
	}

	features := make(map[string]bool, len(claims.Features))
	for key, enabled := range claims.Features {
		normalized, err := normalizeClaimKey(key)
		if err != nil {
			return LicenseClaimsV1{}, err
		}
		if _, exists := features[normalized]; exists {
			return LicenseClaimsV1{}, ErrDuplicateClaimKey
		}
		features[normalized] = enabled
	}

	limits := make(map[string]int64, len(claims.Limits))
	for key, value := range claims.Limits {
		normalized, err := normalizeClaimKey(key)
		if err != nil {
			return LicenseClaimsV1{}, err
		}
		if value < 0 {
			return LicenseClaimsV1{}, ErrInvalidClaimLimit
		}
		if _, exists := features[normalized]; exists {
			return LicenseClaimsV1{}, ErrDuplicateClaimKey
		}
		if _, exists := limits[normalized]; exists {
			return LicenseClaimsV1{}, ErrDuplicateClaimKey
		}
		limits[normalized] = value
	}

	claims.Features = features
	claims.Limits = limits
	return claims, nil
}

func normalizeClaimKey(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", ErrInvalidClaimKey
	}
	return value, nil
}

func marshalClaimsV1(claims LicenseClaimsV1) ([]byte, LicenseClaimsV1, error) {
	normalized, err := normalizeClaimsV1(claims)
	if err != nil {
		return nil, LicenseClaimsV1{}, err
	}

	featureKeys := make([]string, 0, len(normalized.Features))
	for key := range normalized.Features {
		featureKeys = append(featureKeys, key)
	}
	sort.Strings(featureKeys)
	features := make([]featureClaimV1, 0, len(featureKeys))
	for _, key := range featureKeys {
		features = append(features, featureClaimV1{Key: key, Enabled: normalized.Features[key]})
	}

	limitKeys := make([]string, 0, len(normalized.Limits))
	for key := range normalized.Limits {
		limitKeys = append(limitKeys, key)
	}
	sort.Strings(limitKeys)
	limits := make([]limitClaimV1, 0, len(limitKeys))
	for _, key := range limitKeys {
		limits = append(limits, limitClaimV1{Key: key, Value: normalized.Limits[key]})
	}

	payload, err := json.Marshal(licenseClaimsWireV1{
		Version:        LicenseClaimsVersionV1,
		LicenseID:      normalized.LicenseID.String(),
		OrganizationID: normalized.OrganizationID.String(),
		SubscriptionID: normalized.SubscriptionID.String(),
		DeploymentID:   normalized.DeploymentID,
		IssuedAt:       normalized.IssuedAt.Unix(),
		ExpiresAt:      normalized.ExpiresAt.Unix(),
		Features:       features,
		Limits:         limits,
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
	subscriptionID, err := uuid.Parse(wire.SubscriptionID)
	if err != nil {
		return LicenseClaimsV1{}, ErrMalformedClaims
	}

	features := make(map[string]bool, len(wire.Features))
	for _, claim := range wire.Features {
		key, err := normalizeClaimKey(claim.Key)
		if err != nil {
			return LicenseClaimsV1{}, err
		}
		if _, exists := features[key]; exists {
			return LicenseClaimsV1{}, ErrDuplicateClaimKey
		}
		features[key] = claim.Enabled
	}
	limits := make(map[string]int64, len(wire.Limits))
	for _, claim := range wire.Limits {
		key, err := normalizeClaimKey(claim.Key)
		if err != nil {
			return LicenseClaimsV1{}, err
		}
		if claim.Value < 0 {
			return LicenseClaimsV1{}, ErrInvalidClaimLimit
		}
		if _, exists := features[key]; exists {
			return LicenseClaimsV1{}, ErrDuplicateClaimKey
		}
		if _, exists := limits[key]; exists {
			return LicenseClaimsV1{}, ErrDuplicateClaimKey
		}
		limits[key] = claim.Value
	}

	claims, err := normalizeClaimsV1(LicenseClaimsV1{
		LicenseID:      licenseID,
		OrganizationID: organizationID,
		SubscriptionID: subscriptionID,
		DeploymentID:   wire.DeploymentID,
		IssuedAt:       time.Unix(wire.IssuedAt, 0).UTC(),
		ExpiresAt:      time.Unix(wire.ExpiresAt, 0).UTC(),
		Features:       features,
		Limits:         limits,
	})
	if err != nil {
		return LicenseClaimsV1{}, err
	}
	return claims, nil
}
