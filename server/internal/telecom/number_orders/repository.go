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

func (r *Repository) MarkPurchasing(ctx context.Context, id uuid.UUID) (sqlc.NumberOrder, error) {
	return r.queries.MarkNumberOrderPurchasing(ctx, id)
}

func (r *Repository) MarkPurchased(ctx context.Context, id uuid.UUID, providerOrderID string) (sqlc.NumberOrder, error) {
	return r.queries.MarkNumberOrderPurchased(ctx, sqlc.MarkNumberOrderPurchasedParams{
		ID:              id,
		ProviderOrderID: &providerOrderID,
	})
}

func (r *Repository) MarkPersisting(ctx context.Context, id uuid.UUID, providerResourceID *string) (sqlc.NumberOrder, error) {
	return r.queries.MarkNumberOrderPersisting(ctx, sqlc.MarkNumberOrderPersistingParams{
		ID:                 id,
		ProviderResourceID: providerResourceID,
	})
}

func (r *Repository) MarkConfiguring(ctx context.Context, id uuid.UUID, phoneNumberID *uuid.UUID) (sqlc.NumberOrder, error) {
	return r.queries.MarkNumberOrderConfiguring(ctx, sqlc.MarkNumberOrderConfiguringParams{
		ID:            id,
		PhoneNumberID: phoneNumberID,
	})
}

func (r *Repository) MarkCompleted(ctx context.Context, id uuid.UUID) (sqlc.NumberOrder, error) {
	return r.queries.MarkNumberOrderCompleted(ctx, id)
}

func (r *Repository) MarkFailed(ctx context.Context, id uuid.UUID, expected Status, failure Failure) (sqlc.NumberOrder, error) {
	var code *string
	if failure.Code != "" {
		code = &failure.Code
	}
	return r.queries.MarkNumberOrderFailed(ctx, sqlc.MarkNumberOrderFailedParams{
		ID:             id,
		FailedStage:    string(failure.Stage),
		ErrorCode:      code,
		ErrorMessage:   failure.Message,
		ExpectedStatus: string(expected),
	})
}

func (r *Repository) ListRecoverable(ctx context.Context, limit int32) ([]sqlc.NumberOrder, error) {
	return r.queries.ListRecoverableNumberOrders(ctx, limit)
}
