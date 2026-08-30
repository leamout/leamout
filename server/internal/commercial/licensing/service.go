package licensing

import (
	"context"
	"fmt"
	"time"

	"github.com/leamout/leamout/internal/database/sqlc"
)

type Service struct {
	repo    *Repository
	keyring *Keyring
}

func NewService(repo *Repository, keyring *Keyring) *Service {
	return &Service{repo: repo, keyring: keyring}
}

// Issue signs the entitlement snapshot already attached to an active license.
// Plan and organization entitlements are resolved when the license snapshot is
// created; they are intentionally not re-resolved here.
func (s *Service) Issue(ctx context.Context, req IssueRequest) (Document, error) {
	if err := validateIssueRequest(req); err != nil {
		return Document{}, err
	}

	license, err := s.repo.Get(ctx, req.OrganizationID, req.LicenseID)
	if err != nil {
		return Document{}, fmt.Errorf("get license: %w", err)
	}
	if err := validateActiveLicense(license, req.At); err != nil {
		return Document{}, err
	}

	entitlements, err := s.repo.Entitlements(ctx, req.OrganizationID, req.LicenseID)
	if err != nil {
		return Document{}, fmt.Errorf("list license entitlements: %w", err)
	}
	claims, err := buildClaims(license, entitlements, req.At)
	if err != nil {
		return Document{}, err
	}
	signer, err := s.keyring.Signer(*license.SigningKeyID)
	if err != nil {
		return Document{}, err
	}
	return signer.Sign(claims)
}

func buildClaims(license sqlc.License, entitlements []sqlc.Entitlement, at time.Time) (Claims, error) {
	claims := Claims{
		Version: DocumentVersion, OrganizationID: license.OrganizationID, LicenseID: license.ID,
		IssuedAt: license.IssuedAt.Time.UTC(), MaxDeployments: license.MaxDeployments,
		Features: make(map[string]bool), Limits: make(map[string]int64),
	}
	if license.ExpiresAt.Valid {
		expiresAt := license.ExpiresAt.Time.UTC()
		claims.ExpiresAt = &expiresAt
	}
	for _, entitlement := range entitlements {
		if !activeEntitlement(entitlement, at) {
			continue
		}
		switch entitlement.Kind {
		case "feature":
			if entitlement.Enabled == nil || entitlement.LimitValue != nil {
				return Claims{}, fmt.Errorf("invalid feature entitlement %q", entitlement.EntitlementKey)
			}
			claims.Features[entitlement.EntitlementKey] = *entitlement.Enabled
		case "limit":
			if entitlement.LimitValue == nil || entitlement.Enabled != nil || *entitlement.LimitValue < 0 {
				return Claims{}, fmt.Errorf("invalid limit entitlement %q", entitlement.EntitlementKey)
			}
			claims.Limits[entitlement.EntitlementKey] = *entitlement.LimitValue
		default:
			return Claims{}, fmt.Errorf("invalid entitlement kind %q", entitlement.Kind)
		}
	}
	return claims, nil
}

func activeEntitlement(entitlement sqlc.Entitlement, at time.Time) bool {
	return (!entitlement.StartsAt.Valid || !at.Before(entitlement.StartsAt.Time)) &&
		(!entitlement.ExpiresAt.Valid || at.Before(entitlement.ExpiresAt.Time))
}
