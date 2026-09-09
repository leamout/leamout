package trunks

import (
	"context"

	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for SIP trunks.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Trunk, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeTrunks(ctx)
	if err != nil {
		return nil, err
	}
	trunks := make([]Trunk, 0, len(rows))
	for _, row := range rows {
		trunks = append(trunks, Trunk{
			ID:           row.ID,
			Organization: row.OrganizationName,
			Name:         row.Name,
			Mode:         row.ProvisioningMode,
			Provider:     row.ProviderName,
			Direction:    row.Direction,
			Status:       row.Status,
			Endpoints:    row.EndpointCount,
		})
	}
	return trunks, nil
}
