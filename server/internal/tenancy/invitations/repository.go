package invitations

import "github.com/jackc/pgx/v5/pgxpool"

// Repository provides the transaction boundary for invitation and email persistence.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}
