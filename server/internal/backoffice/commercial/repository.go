package commercial

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for commercial account state.
type Repository struct {
	queries *sqlc.Queries
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeCommercialAccount(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	return Detail{Account: Account{OrganizationID: row.OrganizationID, Organization: row.OrganizationName, Plan: row.PlanName, SubscriptionStatus: row.SubscriptionStatus, BillingModel: row.BillingModel, RenewsAt: row.RenewsAt}, SubscriptionID: row.SubscriptionID, PlanID: row.PlanID, PriceID: row.PriceID, PricingType: row.PricingType, Currency: row.Currency, AmountMinor: row.AmountMinor, BillingInterval: row.BillingInterval, StartsAt: row.StartsAt, EndsAt: row.EndsAt, OrganizationCreatedAt: row.OrganizationCreatedAt}, nil
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Account, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeCommercialAccounts(ctx)
	if err != nil {
		return nil, err
	}
	accounts := make([]Account, 0, len(rows))
	for _, row := range rows {
		accounts = append(accounts, Account{
			OrganizationID:     row.OrganizationID,
			Organization:       row.OrganizationName,
			Plan:               row.PlanName,
			SubscriptionStatus: row.SubscriptionStatus,
			BillingModel:       row.BillingModel,
			RenewsAt:           row.RenewsAt,
		})
	}
	return accounts, nil
}
