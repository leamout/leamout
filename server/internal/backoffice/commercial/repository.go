package commercial

import "github.com/leamout/leamout/internal/database/sqlc"

// Repository owns cross-tenant Backoffice reads for commercial state.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}
