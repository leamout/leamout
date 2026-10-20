package organizations

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/leamout/leamout/server/internal/database/sqlc"
)

// Repository owns explicit cross-organization operator read projections.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context, filter Filter) (List, error) {
	rows, err := r.queries.ListBackofficeOrganizations(ctx, sqlc.ListBackofficeOrganizationsParams{
		Search:     filter.Search,
		Status:     filter.Status,
		PageLimit:  pageSize + 1,
		PageOffset: (filter.Page - 1) * pageSize,
	})
	if err != nil {
		return List{}, err
	}
	result := List{Filter: filter, HasNext: len(rows) > pageSize}
	if result.HasNext {
		rows = rows[:pageSize]
	}
	for _, row := range rows {
		result.Organizations = append(result.Organizations, Organization{
			ID:        row.ID,
			Name:      row.Name,
			Status:    row.Status,
			CreatedAt: formatTime(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeOrganization(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	return Detail{
		Organization: Organization{
			ID:        row.ID,
			Name:      row.Name,
			Status:    row.Status,
			CreatedAt: formatTime(row.CreatedAt),
		},
		UpdatedAt:          formatTime(row.UpdatedAt),
		DeletedAt:          formatTime(row.DeletedAt),
		ActiveMembers:      row.ActiveMembers,
		Agents:             row.Agents,
		Trunks:             row.Trunks,
		PlanName:           row.PlanName,
		SubscriptionStatus: row.SubscriptionStatus,
		PeriodEnd:          formatTime(row.CurrentPeriodEnd),
	}, nil
}

func formatTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return "—"
	}
	return value.Time.UTC().Format("2006-01-02 15:04 UTC")
}
