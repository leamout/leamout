package carrierconnections

import (
	"context"

	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for generic carrier connections.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]CarrierConnection, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeCarrierConnections(ctx)
	if err != nil {
		return nil, err
	}
	connections := make([]CarrierConnection, 0, len(rows))
	for _, row := range rows {
		connections = append(connections, CarrierConnection{
			ID:                 row.ID,
			Organization:       row.OrganizationName,
			Name:               row.Name,
			Provider:           row.ProviderName,
			Scope:              row.Scope,
			Status:             row.Status,
			InboundEnabled:     row.InboundEnabled,
			MaxCPS:             row.MaxCps,
			MaxConcurrentCalls: row.MaxConcurrentCalls,
			Trunks:             row.TrunkCount,
		})
	}
	return connections, nil
}
