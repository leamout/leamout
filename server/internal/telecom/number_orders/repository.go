package number_orders

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

func (r *Repository) Create(ctx context.Context, input CreateInput) (sqlc.NumberOrder, error) {
	return r.queries.CreateNumberOrder(ctx, sqlc.CreateNumberOrderParams{
		OrganizationID:      input.OrganizationID,
		ProviderID:          input.ProviderID,
		ProviderInventoryID: input.ProviderInventoryID,
		ProviderProductID:   input.ProviderProductID,
		Number:              input.Number,
		CountryCode:         input.CountryCode,
	})
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.NumberOrder, error) {
	return r.queries.GetNumberOrderByID(ctx, sqlc.GetNumberOrderByIDParams{
		ID:             id,
		OrganizationID: organizationID,
	})
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]sqlc.NumberOrder, error) {
	return r.queries.ListNumberOrdersByOrganizationID(ctx, organizationID)
}
