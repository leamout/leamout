package providercdrs

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for provider CDRs.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]ProviderCDR, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	return r.queries.ListBackofficeProviderCDRs(ctx)
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	return r.queries.GetBackofficeProviderCDR(ctx, id)
}
