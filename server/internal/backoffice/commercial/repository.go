package commercial

import (
	"context"
	"fmt"

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
	return Detail{
		Account: Account{
			OrganizationID: row.OrganizationID,
			Organization:   row.OrganizationName,
			BillingModel:   row.BillingModel,
			WalletCount:    row.WalletCount,
			Currencies:     fmt.Sprint(row.Currencies),
		},
		OrganizationCreatedAt: row.OrganizationCreatedAt,
	}, nil
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
			OrganizationID: row.OrganizationID,
			Organization:   row.OrganizationName,
			BillingModel:   row.BillingModel,
			WalletCount:    row.WalletCount,
			Currencies:     fmt.Sprint(row.Currencies),
		})
	}
	return accounts, nil
}
