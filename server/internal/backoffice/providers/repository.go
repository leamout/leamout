package providers

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for managed carrier providers.
type Repository struct {
	queries *sqlc.Queries
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeProvider(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	operationRows, err := r.queries.ListBackofficeProviderOperations(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	operations := make([]Operation, 0, len(operationRows))
	for _, v := range operationRows {
		operations = append(operations, Operation{ID: v.ID, Organization: v.OrganizationName, Number: v.Number, Type: v.OperationType, State: v.State, Attempts: v.Attempts, ProviderOperationID: v.ProviderOperationID, LastError: v.LastError, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt})
	}
	return Detail{Provider: Provider{ID: row.ID, Slug: row.Slug, Name: row.Name, Adapter: row.Adapter, Status: row.Status, Connections: row.ConnectionCount, PendingOperations: row.PendingOperationCount, FailedOperations: row.FailedOperationCount}, PhoneNumbers: row.PhoneNumberCount, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Operations: operations}, nil
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
