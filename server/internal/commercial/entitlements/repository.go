package entitlements

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository persists durable entitlement records through SQLC.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) CreatePlan(ctx context.Context, planID uuid.UUID, input CreateInput) (Entitlement, error) {
	row, err := r.queries.CreatePlanEntitlement(ctx, sqlc.CreatePlanEntitlementParams{
		EntitlementKey: input.Key,
		Kind:           string(input.Kind),
		Enabled:        input.Enabled,
		LimitValue:     input.LimitValue,
		StartsAt:       pgconv.NullableTimestamptz(input.StartsAt),
		ExpiresAt:      pgconv.NullableTimestamptz(input.ExpiresAt),
		PlanID:         planID,
	})
	if err != nil {
		return Entitlement{}, mapWriteError(err)
	}
	return entitlementFromRow(row), nil
}

func (r *Repository) CreateOrganization(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Entitlement, error) {
	row, err := r.queries.CreateOrganizationEntitlement(ctx, sqlc.CreateOrganizationEntitlementParams{
		EntitlementKey: input.Key,
		Kind:           string(input.Kind),
		Enabled:        input.Enabled,
		LimitValue:     input.LimitValue,
		StartsAt:       pgconv.NullableTimestamptz(input.StartsAt),
		ExpiresAt:      pgconv.NullableTimestamptz(input.ExpiresAt),
		OrganizationID: organizationID,
	})
	if err != nil {
		return Entitlement{}, mapWriteError(err)
	}
	return entitlementFromRow(row), nil
}

func (r *Repository) CreateLicense(ctx context.Context, organizationID, licenseID uuid.UUID, input CreateInput) (Entitlement, error) {
	row, err := r.queries.CreateLicenseEntitlement(ctx, sqlc.CreateLicenseEntitlementParams{
		EntitlementKey: input.Key,
		Kind:           string(input.Kind),
		Enabled:        input.Enabled,
		LimitValue:     input.LimitValue,
		StartsAt:       pgconv.NullableTimestamptz(input.StartsAt),
		ExpiresAt:      pgconv.NullableTimestamptz(input.ExpiresAt),
		LicenseID:      licenseID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return Entitlement{}, mapWriteError(err)
	}
	return entitlementFromRow(row), nil
}

func (r *Repository) ListPlan(ctx context.Context, planID uuid.UUID) ([]Entitlement, error) {
	rows, err := r.queries.ListPlanEntitlements(ctx, &planID)
	if err != nil {
		return nil, err
	}
	return entitlementsFromRows(rows), nil
}

func (r *Repository) ListOrganization(ctx context.Context, organizationID uuid.UUID) ([]Entitlement, error) {
	rows, err := r.queries.ListOrganizationEntitlements(ctx, &organizationID)
	if err != nil {
		return nil, err
	}
	return entitlementsFromRows(rows), nil
}

func (r *Repository) ListLicense(ctx context.Context, organizationID, licenseID uuid.UUID) ([]Entitlement, error) {
	rows, err := r.queries.ListLicenseEntitlements(ctx, sqlc.ListLicenseEntitlementsParams{
		LicenseID:      &licenseID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return nil, err
	}
	return entitlementsFromRows(rows), nil
}

func (r *Repository) DeletePlan(ctx context.Context, planID, id uuid.UUID) error {
	return r.queries.DeletePlanEntitlement(ctx, sqlc.DeletePlanEntitlementParams{ID: id, PlanID: &planID})
}

func (r *Repository) DeleteOrganization(ctx context.Context, organizationID, id uuid.UUID) error {
	return r.queries.DeleteOrganizationEntitlement(ctx, sqlc.DeleteOrganizationEntitlementParams{ID: id, OrganizationID: &organizationID})
}

func (r *Repository) DeleteLicense(ctx context.Context, organizationID, licenseID, id uuid.UUID) error {
	return r.queries.DeleteLicenseEntitlement(ctx, sqlc.DeleteLicenseEntitlementParams{
		ID:             id,
		LicenseID:      &licenseID,
		OrganizationID: organizationID,
	})
}

func entitlementFromRow(row sqlc.Entitlement) Entitlement {
	return Entitlement{
		ID:             row.ID,
		PlanID:         row.PlanID,
		OrganizationID: row.OrganizationID,
		LicenseID:      row.LicenseID,
		Key:            row.EntitlementKey,
		Kind:           Kind(row.Kind),
		Enabled:        row.Enabled,
		LimitValue:     row.LimitValue,
		StartsAt:       pgconv.TimestamptzToTimePtr(row.StartsAt),
		ExpiresAt:      pgconv.TimestamptzToTimePtr(row.ExpiresAt),
		CreatedAt:      pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:      pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func entitlementsFromRows(rows []sqlc.Entitlement) []Entitlement {
	result := make([]Entitlement, 0, len(rows))
	for _, row := range rows {
		result = append(result, entitlementFromRow(row))
	}
	return result
}

func mapWriteError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrScopeUnavailable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrEntitlementConflict
		case "23514":
			return ErrInvalidEntitlement
		}
	}
	return err
}
