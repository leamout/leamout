package sipdomains

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for SIP domains.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]SIPDomain, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	return r.queries.ListBackofficeSIPDomains(ctx)
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	return r.queries.GetBackofficeSIPDomain(ctx, id)
}
