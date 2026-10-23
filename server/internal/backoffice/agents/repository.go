package agents

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
	var organizationID *uuid.UUID
	if filter.OrganizationID != "" {
		id, err := uuid.Parse(filter.OrganizationID)
		if err != nil {
			return List{}, err
		}
		organizationID = &id
	}
	rows, err := r.queries.ListBackofficeAgents(ctx, sqlc.ListBackofficeAgentsParams{
		Search:         filter.Search,
		Status:         filter.Status,
		Engine:         filter.Engine,
		OrganizationID: organizationID,
		PageLimit:      pageSize + 1,
		PageOffset:     (filter.Page - 1) * pageSize,
	})
	if err != nil {
		return List{}, err
	}
	result := List{Filter: filter, HasNext: len(rows) > pageSize}
	if result.HasNext {
		rows = rows[:pageSize]
	}
	for _, row := range rows {
		result.Agents = append(result.Agents, Agent{
			ID:                 row.ID,
			Name:               row.Name,
			Status:             row.Status,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			Engine:             row.Engine,
			Language:           row.Language,
			CreatedAt:          formatTime(row.CreatedAt),
		})
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficeAgent(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	return Detail{
		Agent: Agent{
			ID:                 row.ID,
			Name:               row.Name,
			Status:             row.Status,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			Engine:             row.Engine,
			Language:           row.Language,
			CreatedAt:          formatTime(row.CreatedAt),
		},
		UpdatedAt:             formatTime(row.UpdatedAt),
		Instructions:          row.Instructions,
		Voice:                 row.Voice,
		InterruptionPolicy:    row.InterruptionPolicy,
		RecordingPolicy:       row.RecordingPolicy,
		ConfigurationRevision: row.ConfigurationRevision,
	}, nil
}

func formatTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return "—"
	}
	return value.Time.UTC().Format("2006-01-02 15:04 UTC")
}
