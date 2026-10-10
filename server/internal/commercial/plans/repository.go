package plans

import (
	"context"
	"encoding/json"

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
		value, err := fromRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Plan, error) {
	row, err := r.queries.GetPlanByID(ctx, id)
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row)
}

func (r *Repository) GetByCode(ctx context.Context, code string) (Plan, error) {
	row, err := r.queries.GetPlanByCode(ctx, code)
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row)
}

func (r *Repository) Upsert(ctx context.Context, id uuid.UUID, req UpsertRequest) (Plan, error) {
	entitlements, err := json.Marshal(req.Entitlements)
	if err != nil {
		return Plan{}, err
	}
	limits, err := json.Marshal(req.Limits)
	if err != nil {
		return Plan{}, err
	}
	row, err := r.queries.UpsertPlan(ctx, sqlc.UpsertPlanParams{
		ID:              id,
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		PricingType:     req.PricingType,
		Currency:        req.Currency,
		AmountMinor:     req.AmountMinor,
		BillingInterval: req.BillingInterval,
		Entitlements:    entitlements,
		Limits:          limits,
		Status:          req.Status,
	})
	if err != nil {
		return Plan{}, err
	}
	return fromRow(row)
}

func fromRow(row sqlc.Plan) (Plan, error) {
	entitlements := map[string]bool{}
	if len(row.Entitlements) > 0 {
		if err := json.Unmarshal(row.Entitlements, &entitlements); err != nil {
			return Plan{}, err
		}
	}
	limits := map[string]int64{}
	if len(row.Limits) > 0 {
		if err := json.Unmarshal(row.Limits, &limits); err != nil {
			return Plan{}, err
		}
	}
	return Plan{
		ID:              row.ID,
		Code:            row.Code,
		Name:            row.Name,
		Description:     row.Description,
		PricingType:     row.PricingType,
		Currency:        row.Currency,
		AmountMinor:     row.AmountMinor,
		BillingInterval: row.BillingInterval,
		Entitlements:    entitlements,
		Limits:          limits,
		Status:          row.Status,
		CreatedAt:       pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgconv.TimestamptzToTime(row.UpdatedAt),
	}, nil
}

