package managedvoice

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, queries: sqlc.New(db)}
}

func (r *Repository) ManagedSpendMicros(ctx context.Context, organizationID uuid.UUID, day time.Time) (int64, error) {
	return r.queries.SumWholesaleChargesForDay(ctx, sqlc.SumWholesaleChargesForDayParams{
		OrganizationID: organizationID,
		Day:            pgtype.Timestamptz{Time: day, Valid: true},
	})
}
