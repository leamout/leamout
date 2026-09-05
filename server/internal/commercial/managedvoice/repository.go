package managedvoice

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ db *pgxpool.Pool }

func NewRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

func (r *Repository) ManagedSpendMicros(ctx context.Context, organizationID uuid.UUID, day time.Time) (int64, error) {
	var amount int64
	err := r.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount_micros),0) FROM wholesale_charges WHERE organization_id=$1 AND occurred_at >= $2 AND occurred_at < $2 + interval '1 day'`, organizationID, day).Scan(&amount)
	return amount, err
}
