package providers

import (
	"context"

	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for managed carrier providers.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Provider, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeProviders(ctx)
	if err != nil {
		return nil, err
	}
	providers := make([]Provider, 0, len(rows))
	for _, row := range rows {
		providers = append(providers, Provider{
			ID:                row.ID,
			Slug:              row.Slug,
			Name:              row.Name,
			Adapter:           row.Adapter,
			Status:            row.Status,
			Connections:       row.ConnectionCount,
			PendingOperations: row.PendingOperationCount,
			FailedOperations:  row.FailedOperationCount,
		})
	}
	return providers, nil
}
