package commercial

import (
	"context"

	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for commercial account state.
type Repository struct {
	queries *sqlc.Queries
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
			BillingProvider:    row.BillingProvider,
			RenewsAt:           row.RenewsAt,
		})
	}
	return accounts, nil
}
