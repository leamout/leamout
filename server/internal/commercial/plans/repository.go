package plans

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) ListActive(ctx context.Context) ([]Plan, error) {
	rows, err := r.queries.ListActivePlans(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Plan, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromRow(row))
	}
	return result, nil
}

func (r *Repository) ListEntitlements(ctx context.Context, planID uuid.UUID) ([]Entitlement, error) {
	rows, err := r.queries.ListPlanEntitlements(ctx, planID)
	if err != nil {
		return nil, err
	}
	result := make([]Entitlement, 0, len(rows))
	for _, row := range rows {
		result = append(result, Entitlement{
			Capability: row.Capability,
			Enabled:    row.Enabled,
		})
	}
	return result, nil
}

func (r *Repository) SetEntitlement(
	ctx context.Context,
	planID uuid.UUID,
	capability string,
	enabled bool,
) error {
	return r.queries.UpsertPlanEntitlement(
		ctx,
		sqlc.UpsertPlanEntitlementParams{
			PlanID:     planID,
			Capability: capability,
			Enabled:    enabled,
		},
	)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Plan, error) {
	row, err := r.queries.GetPlanByID(ctx, id)
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (Plan, error) {
	row, err := r.queries.GetPlanByCode(ctx, code)
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) Upsert(ctx context.Context, id uuid.UUID, req UpsertRequest) (Plan, error) {
	row, err := r.queries.UpsertPlan(ctx, sqlc.UpsertPlanParams{
		ID:              id,
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		Currency:        req.Currency,
		AmountMinor:     req.AmountMinor,
		BillingInterval: req.BillingInterval,
		Status:          req.Status,
	})
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row), nil
}

func fromRow(row sqlc.Plan) Plan {
	return Plan{
		ID:              row.ID,
		Code:            row.Code,
		Name:            row.Name,
		Description:     row.Description,
		Currency:        row.Currency,
		AmountMinor:     row.AmountMinor,
		BillingInterval: row.BillingInterval,
		Status:          row.Status,
		CreatedAt:       pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}
