package billing

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// Repository pages through locally known Polar subscriptions for reconciliation.
type Repository struct {
	queries  *sqlc.Queries
	database *pgxpool.Pool
}

func NewRepository(queries *sqlc.Queries, databases ...*pgxpool.Pool) *Repository {
	r := &Repository{
		queries: queries,
	}
	if len(databases) > 0 {
		r.database = databases[0]
	}
	return r
}

func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{
		queries:  r.queries.WithTx(tx),
		database: r.database,
	}
}

func (r *Repository) ListReconciliationTargets(ctx context.Context, after uuid.UUID) ([]ReconciliationTarget, error) {
	rows, err := r.queries.ListBillingOrganizationReconciliationTargets(ctx, after)
	if err != nil {
		return nil, err
	}
	result := make([]ReconciliationTarget, 0, len(rows))
	for _, row := range rows {
		result = append(result, ReconciliationTarget{
			OrganizationID: row.OrganizationID,
			SubscriptionID: stringValue(row.SubscriptionID),
			CustomerID:     stringValue(row.CustomerID),
		})
	}
	return result, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
