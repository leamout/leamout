package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// Repository pages through locally known Stripe subscriptions for reconciliation.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{
		queries: queries,
	}
}

func (r *Repository) ListReconciliationTargets(ctx context.Context, after uuid.UUID) ([]ReconciliationTarget, error) {
	rows, err := r.queries.ListBillingReconciliationTargets(ctx, after)
	if err != nil {
		return nil, err
	}
	result := make([]ReconciliationTarget, 0, len(rows))
	for _, row := range rows {
		result = append(result, ReconciliationTarget{
			OrganizationID: row.OrganizationID,
			SubscriptionID: row.SubscriptionID,
			CustomerID:     row.CustomerID,
		})
	}
	return result, nil
}
