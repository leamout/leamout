package catalog

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

// Repository owns durable product and plan persistence.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) CreateProduct(ctx context.Context, input CreateProductInput) (Product, error) {
	var active any
	if input.Active != nil {
		active = *input.Active
	}
	row, err := r.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
		Active:      active,
	})
	if err != nil {
		return Product{}, mapWriteError(err)
	}
	return productFromRow(row), nil
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

func (r *Repository) UpdateProduct(ctx context.Context, id uuid.UUID, input UpdateProductInput) (Product, error) {
	row, err := r.queries.UpdateProduct(ctx, sqlc.UpdateProductParams{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
		Active:      input.Active,
		ID:          id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Product{}, ErrProductNotFound
		}
		return Product{}, mapWriteError(err)
	}
	return productFromRow(row), nil
}

func (r *Repository) CreatePlan(ctx context.Context, input CreatePlanInput) (Plan, error) {
	row, err := r.queries.CreatePlan(ctx, sqlc.CreatePlanParams{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
		Active:      input.Active,
		ProductID:   input.ProductID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Plan{}, ErrProductInactive
		}
		return Plan{}, mapWriteError(err)
	}
	return planFromRow(row), nil
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

func (r *Repository) UpdatePlan(ctx context.Context, id uuid.UUID, input UpdatePlanInput) (Plan, error) {
	row, err := r.queries.UpdatePlan(ctx, sqlc.UpdatePlanParams{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
		Active:      input.Active,
		ID:          id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Plan{}, ErrPlanNotFound
		}
		return Plan{}, mapWriteError(err)
	}
	return Plan{
		ID:          row.ID_2,
		ProductID:   row.ProductID,
		Code:        row.Code_2,
		Name:        row.Name_2,
		Description: row.Description_2,
		Active:      row.Active_2,
		CreatedAt:   pgconv.TimestamptzToTime(row.CreatedAt_2),
		UpdatedAt:   pgconv.TimestamptzToTime(row.UpdatedAt_2),
	}, nil
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

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrCodeConflict
	}
	return err
}
