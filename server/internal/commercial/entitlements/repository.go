package entitlements

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Source provides the three entitlement layers used during resolution.
type Source interface {
	ListPlan(context.Context, uuid.UUID) ([]Entitlement, error)
	ListOrganization(context.Context, uuid.UUID) ([]Entitlement, error)
	ListLicense(context.Context, uuid.UUID, uuid.UUID) ([]Entitlement, error)
}

// Repository loads organization-scoped entitlement data from PostgreSQL.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	if db == nil {
		panic("entitlements: database is required")
	}
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) ListPlan(ctx context.Context, planID uuid.UUID) ([]Entitlement, error) {
	rows, err := r.queries.ListPlanEntitlements(ctx, &planID)
	if err != nil {
		return nil, fmt.Errorf("list plan entitlements: %w", err)
	}
	return fromRows(rows)
}

func (r *Repository) ListOrganization(ctx context.Context, organizationID uuid.UUID) ([]Entitlement, error) {
	rows, err := r.queries.ListOrganizationEntitlements(ctx, &organizationID)
	if err != nil {
		return nil, fmt.Errorf("list organization entitlements: %w", err)
	}
	return fromRows(rows)
}

func (r *Repository) ListLicense(ctx context.Context, organizationID, licenseID uuid.UUID) ([]Entitlement, error) {
	rows, err := r.queries.ListLicenseEntitlements(ctx, sqlc.ListLicenseEntitlementsParams{
		LicenseID: &licenseID, OrganizationID: organizationID,
	})
	if err != nil {
		return nil, fmt.Errorf("list license entitlements: %w", err)
	}
	return fromRows(rows)
}

func fromRows(rows []sqlc.Entitlement) ([]Entitlement, error) {
	result := make([]Entitlement, 0, len(rows))
	for _, row := range rows {
		entitlement := Entitlement{
			Key: row.EntitlementKey, Kind: Kind(row.Kind), Enabled: row.Enabled, Limit: row.LimitValue,
			StartsAt: pgconv.TimestamptzToTimePtr(row.StartsAt), ExpiresAt: pgconv.TimestamptzToTimePtr(row.ExpiresAt),
		}
		if err := Validate(entitlement); err != nil {
			return nil, fmt.Errorf("invalid persisted entitlement %q: %w", row.EntitlementKey, err)
		}
		result = append(result, entitlement)
	}
	return result, nil
}
