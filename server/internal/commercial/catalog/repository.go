package catalog

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository reads durable product, plan, and price catalog state.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) GetProduct(ctx context.Context, id uuid.UUID) (Product, error) {
	row, err := r.queries.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrProductNotFound
		}
		return Product{}, err
	}
	return productFromRow(row), nil
}

func (r *Repository) GetProductByCode(ctx context.Context, code string) (Product, error) {
	row, err := r.queries.GetProductByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrProductNotFound
		}
		return Product{}, err
	}
	return productFromRow(row), nil
}

func (r *Repository) ListProducts(ctx context.Context, activeOnly bool) ([]Product, error) {
	var (
		rows []sqlc.Product
		err  error
	)
	if activeOnly {
		rows, err = r.queries.ListActiveProducts(ctx)
	} else {
		rows, err = r.queries.ListProducts(ctx)
	}
	if err != nil {
		return nil, err
	}
	products := make([]Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, productFromRow(row))
	}
	return products, nil
}

func (r *Repository) GetPlan(ctx context.Context, id uuid.UUID) (Plan, error) {
	row, err := r.queries.GetPlanByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Plan{}, ErrPlanNotFound
		}
		return Plan{}, err
	}
	return planFromRow(row), nil
}

func (r *Repository) GetPlanByCode(ctx context.Context, code string) (Plan, error) {
	row, err := r.queries.GetPlanByCode(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Plan{}, ErrPlanNotFound
		}
		return Plan{}, err
	}
	return planFromRow(row), nil
}

func (r *Repository) ListPlans(ctx context.Context, productID uuid.UUID, activeOnly bool) ([]Plan, error) {
	var (
		rows []sqlc.Plan
		err  error
	)
	if activeOnly {
		rows, err = r.queries.ListActivePlansByProduct(ctx, productID)
	} else {
		rows, err = r.queries.ListPlansByProduct(ctx, productID)
	}
	if err != nil {
		return nil, err
	}
	plans := make([]Plan, 0, len(rows))
	for _, row := range rows {
		plans = append(plans, planFromRow(row))
	}
	return plans, nil
}

func (r *Repository) GetPrice(ctx context.Context, id uuid.UUID) (Price, error) {
	row, err := r.queries.GetPriceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Price{}, ErrPriceNotFound
		}
		return Price{}, err
	}
	return priceFromRow(row), nil
}

func (r *Repository) ListPrices(ctx context.Context, planID uuid.UUID, activeOnly bool, at time.Time) ([]Price, error) {
	var (
		rows []sqlc.Price
		err  error
	)
	if activeOnly {
		rows, err = r.queries.ListActivePricesByPlan(ctx, sqlc.ListActivePricesByPlanParams{
			PlanID: planID,
			At:     pgconv.NullableTimestamptz(&at),
		})
	} else {
		rows, err = r.queries.ListPricesByPlan(ctx, planID)
	}
	if err != nil {
		return nil, err
	}
	prices := make([]Price, 0, len(rows))
	for _, row := range rows {
		prices = append(prices, priceFromRow(row))
	}
	return prices, nil
}

func productFromRow(row sqlc.Product) Product {
	return Product{
		ID:          row.ID,
		Code:        row.Code,
		Name:        row.Name,
		Description: row.Description,
		Active:      row.Active,
		CreatedAt:   pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:   pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func planFromRow(row sqlc.Plan) Plan {
	return Plan{
		ID:          row.ID,
		ProductID:   row.ProductID,
		Code:        row.Code,
		Name:        row.Name,
		Description: row.Description,
		Active:      row.Active,
		CreatedAt:   pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:   pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func priceFromRow(row sqlc.Price) Price {
	return Price{
		ID:              row.ID,
		PlanID:          row.PlanID,
		Currency:        row.Currency,
		AmountMinor:     row.AmountMinor,
		BillingInterval: BillingInterval(row.BillingInterval),
		Active:          row.Active,
		EffectiveFrom:   pgconv.TimestamptzToTime(row.EffectiveFrom),
		EffectiveUntil:  pgconv.TimestamptzToTimePtr(row.EffectiveUntil),
		CreatedAt:       pgconv.TimestamptzToTime(row.CreatedAt),
	}
}
