package licensing

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	licenseID uuid.UUID,
) (sqlc.License, error) {
	return r.queries.GetLicense(ctx, sqlc.GetLicenseParams{
		OrganizationID: organizationID,
		ID:             licenseID,
	})
}

func (r *Repository) Entitlements(
	ctx context.Context,
	organizationID uuid.UUID,
	licenseID uuid.UUID,
) ([]sqlc.Entitlement, error) {
	return r.queries.ListLicenseEntitlements(ctx, sqlc.ListLicenseEntitlementsParams{
		OrganizationID: organizationID,
		LicenseID:      &licenseID,
	})
}
