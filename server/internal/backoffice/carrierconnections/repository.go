package carrierconnections

import (
	"context"

	"github.com/google/uuid"
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
			OrganizationID:     row.OrganizationID,
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

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeCarrierConnection(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	ipRows, err := r.queries.ListBackofficeCarrierConnectionSourceIPs(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	ips := make([]SourceIP, 0, len(ipRows))
	for _, v := range ipRows {
		ips = append(ips, SourceIP{ID: v.ID, CIDR: v.Cidr, CreatedAt: v.CreatedAt})
	}
	resourceRows, err := r.queries.ListBackofficeCarrierConnectionResources(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	resources := make([]ProviderResource, 0, len(resourceRows))
	for _, v := range resourceRows {
		resources = append(resources, ProviderResource{Type: v.ResourceType, ProviderResourceID: v.ProviderResourceID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt})
	}
	return Detail{CarrierConnection: CarrierConnection{ID: row.ID, OrganizationID: row.OrganizationID, Organization: row.OrganizationName, Name: row.Name, Provider: row.ProviderName, Scope: row.Scope, Status: row.Status, InboundEnabled: row.InboundEnabled, MaxCPS: row.MaxCps, MaxConcurrentCalls: row.MaxConcurrentCalls, Trunks: row.TrunkCount}, ProviderID: row.ProviderID, OutboundAuth: row.OutboundAuthMethod, InboundAuth: row.InboundAuthMethod, MaxDailyMinutes: row.MaxDailyMinutes, Codecs: row.Codecs, SupportsVideo: row.SupportsVideo, SupportsFax: row.SupportsFax, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, SourceIPs: ips, Resources: resources}, nil
}
