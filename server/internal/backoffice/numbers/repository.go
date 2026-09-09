package numbers

import "github.com/leamout/leamout/internal/database/sqlc"

// Repository owns cross-tenant Backoffice reads for phone numbers.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}
