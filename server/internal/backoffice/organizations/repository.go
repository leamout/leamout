package organizations

import (
	"context"

	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for organizations.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Organization, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficeOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	organizations := make([]Organization, 0, len(rows))
	for _, row := range rows {
		organizations = append(organizations, Organization{
			ID:        row.ID,
			Name:      row.Name,
			Plan:      row.PlanName,
			Members:   row.MemberCount,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
		})
	}
	return organizations, nil
}
