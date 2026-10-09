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
	rows, err := r.queries.ListActiveSubscriptionPlans(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]Plan, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromRow(row))
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Plan, error) {
	row, err := r.queries.GetSubscriptionPlanByID(ctx, id)
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) GetByCode(ctx context.Context, code string) (Plan, error) {
	row, err := r.queries.GetSubscriptionPlanByCode(ctx, code)
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) Upsert(ctx context.Context, id uuid.UUID, req UpsertRequest) (Plan, error) {
	row, err := r.queries.UpsertSubscriptionPlan(ctx, sqlc.UpsertSubscriptionPlanParams{
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

func fromRow(row sqlc.SubscriptionPlan) Plan {
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
