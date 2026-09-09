package calls

import (
	"context"

	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for calls.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Call, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeCalls(ctx)
	if err != nil {
		return nil, err
	}
	calls := make([]Call, 0, len(rows))
	for _, row := range rows {
		calls = append(calls, Call{
			ID:           row.ID,
			Organization: row.OrganizationName,
			From:         row.FromUri,
			To:           row.ToUri,
			Direction:    row.Direction,
			Duration:     formatDuration(row.DurationSeconds),
			Status:       row.State,
			CreatedAt:    row.CreatedAt,
		})
	}
	return calls, nil
}
