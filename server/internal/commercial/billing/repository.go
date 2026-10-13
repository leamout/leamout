package billing

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// Repository pages through locally known Stripe subscriptions for reconciliation.
type Repository struct{ db sqlc.DBTX }

func NewRepository(db sqlc.DBTX) *Repository { return &Repository{db: db} }

type ReconciliationTarget struct {
	OrganizationID uuid.UUID
	SubscriptionID string
	CustomerID     string
}

func (r *Repository) ListReconciliationTargets(ctx context.Context, after uuid.UUID) ([]ReconciliationTarget, error) {
	rows, err := r.db.Query(ctx, `SELECT organization_id, provider_subscription_id, provider_customer_id FROM subscriptions
 WHERE provider = 'stripe' AND provider_subscription_id IS NOT NULL AND provider_customer_id IS NOT NULL
 AND organization_id > $1 ORDER BY organization_id LIMIT 100`, after)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]ReconciliationTarget, 0, 100)
	for rows.Next() {
		var value ReconciliationTarget
		if err := rows.Scan(&value.OrganizationID, &value.SubscriptionID, &value.CustomerID); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}
