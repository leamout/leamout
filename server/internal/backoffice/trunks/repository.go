package trunks

import (
	"context"

	"github.com/google/uuid"
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
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Organization:   row.OrganizationName,
			Name:           row.Name,
			Mode:           row.ProvisioningMode,
			Provider:       row.ProviderName,
			Direction:      row.Direction,
			Status:         row.Status,
			Endpoints:      row.EndpointCount,
		})
	}
	return trunks, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeTrunk(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	endpointRows, err := r.queries.ListBackofficeTrunkEndpoints(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	endpoints := make([]Endpoint, 0, len(endpointRows))
	for _, e := range endpointRows {
		endpoints = append(
			endpoints,
			Endpoint{
				ID:           e.ID,
				Host:         e.Host,
				Port:         e.Port,
				Transport:    e.Transport,
				Direction:    e.Direction,
				Priority:     e.Priority,
				Weight:       e.Weight,
				Enabled:      e.Enabled,
				Health:       e.HealthStatus,
				Failures:     e.ConsecutiveFailures,
				LastResponse: e.LastResponseCode,
				LastLatency:  e.LastLatencyMs,
				LastError:    e.LastError,
				LastChecked:  e.LastCheckedAt,
			},
		)
	}
	return Detail{
		Trunk: Trunk{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Organization:   row.OrganizationName,
			Name:           row.Name,
			Mode:           row.ProvisioningMode,
			Provider:       row.ProviderName,
			Direction:      row.Direction,
			Status:         row.Status,
			Endpoints:      row.EndpointCount,
		},
		ManagedDefault:      row.ManagedDefault,
		CarrierConnectionID: row.CarrierConnectionID,
		CarrierConnection:   row.CarrierConnectionName,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
		EnabledEndpoints:    row.EnabledEndpointCount,
		EndpointList:        endpoints,
	}, nil
}
