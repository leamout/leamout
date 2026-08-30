package licensing

import (
	"errors"
	"fmt"
	"strings"

	"github.com/leamout/leamout/internal/commercial/entitlements"
)

var (
	ErrOrganizationRequired   = errors.New("organization_id is required")
	ErrInvalidDeploymentLimit = errors.New("max_deployments must be greater than zero")
	ErrInvalidExpiration      = errors.New("expires_at must be after issued_at")
	ErrInvalidClaims          = errors.New("invalid license claims")
)

// Validate checks invariants required before a license is persisted or issued.
func Validate(license License) error {
	if strings.TrimSpace(license.OrganizationID) == "" {
		return ErrOrganizationRequired
	}
	if license.MaxDeployments <= 0 {
		return ErrInvalidDeploymentLimit
	}
	if license.ExpiresAt != nil && !license.ExpiresAt.After(license.IssuedAt) {
		return ErrInvalidExpiration
	}
	return nil
}

// ValidateClaims checks the complete offline authority before signing and
// after verification. Validation on both sides prevents correctly signed but
// semantically unusable documents from reaching feature enforcement.
func ValidateClaims(claims Claims) error {
	if claims.Version != DocumentVersion {
		return fmt.Errorf("%w: unsupported version %d", ErrInvalidClaims, claims.Version)
	}
	if strings.TrimSpace(claims.LicenseID) == "" {
		return fmt.Errorf("%w: license_id is required", ErrInvalidClaims)
	}
	if strings.TrimSpace(claims.OrganizationID) == "" {
		return fmt.Errorf("%w: organization_id is required", ErrInvalidClaims)
	}
	if claims.MaxDeployments <= 0 {
		return fmt.Errorf("%w: max_deployments must be greater than zero", ErrInvalidClaims)
	}
	if claims.IssuedAt.IsZero() {
		return fmt.Errorf("%w: issued_at is required", ErrInvalidClaims)
	}
	if claims.ExpiresAt != nil && !claims.ExpiresAt.After(claims.IssuedAt) {
		return fmt.Errorf("%w: expires_at must be after issued_at", ErrInvalidClaims)
	}
	seen := make(map[string]struct{}, len(claims.Entitlements))
	for _, claim := range claims.Entitlements {
		entitlement := entitlements.Entitlement{Key: claim.Key, Kind: claim.Kind, Enabled: claim.Enabled, Limit: claim.Limit}
		if err := entitlements.Validate(entitlement); err != nil {
			return fmt.Errorf("%w: entitlement %q: %v", ErrInvalidClaims, claim.Key, err)
		}
		if _, exists := seen[claim.Key]; exists {
			return fmt.Errorf("%w: duplicate entitlement %q", ErrInvalidClaims, claim.Key)
		}
		seen[claim.Key] = struct{}{}
	}
	return nil
}
