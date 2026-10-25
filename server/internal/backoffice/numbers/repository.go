package numbers

import (
	"context"

	"github.com/google/uuid"

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
	organizationID, err := optionalID(filter.OrganizationID)
	if err != nil {
		return List{}, err
	}
	trunkID, err := optionalID(filter.TrunkID)
	if err != nil {
		return List{}, err
	}
	rows, err := r.queries.ListBackofficePhoneNumbers(ctx, sqlc.ListBackofficePhoneNumbersParams{
		Search:         filter.Search,
		Status:         filter.Status,
		OrganizationID: organizationID,
		PageLimit:      pageSize + 1,
		PageOffset:     (filter.Page - 1) * pageSize,
		TrunkID:        trunkID,
		Country:        filter.Country,
		Voice:          filter.Voice,
	})
	if err != nil {
		return List{}, err
	}
	result := List{
		Filter:  filter,
		HasNext: len(rows) > pageSize,
	}
	if result.HasNext {
		rows = rows[:pageSize]
	}
	for _, row := range rows {
		result.Numbers = append(result.Numbers, Number{
			ID:                 row.ID,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			Number:             row.Number,
			CountryCode:        row.CountryCode,
			TrunkID:            row.TrunkID,
			TrunkName:          row.TrunkName,
			TrunkStatus:        row.TrunkStatus,
			AgentID:            row.AgentID,
			AgentName:          row.AgentName,
			AgentStatus:        row.AgentStatus,
			VoiceEnabled:       row.VoiceEnabled,
			Status:             row.Status,
			CreatedAt:          row.CreatedAt,
		})
	}
	return result, nil
}

func (r *Repository) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficePhoneNumber(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	result := Detail{
		Number: Number{
			ID:                 row.ID,
			OrganizationID:     row.OrganizationID,
			OrganizationName:   row.OrganizationName,
			OrganizationStatus: row.OrganizationStatus,
			Number:             row.Number,
			CountryCode:        row.CountryCode,
			TrunkID:            row.TrunkID,
			TrunkName:          row.TrunkName,
			TrunkStatus:        row.TrunkStatus,
			AgentID:            row.AgentID,
			AgentName:          row.AgentName,
			AgentStatus:        row.AgentStatus,
			VoiceEnabled:       row.VoiceEnabled,
			Status:             row.Status,
			CreatedAt:          row.CreatedAt,
		},
		UpdatedAt: row.UpdatedAt,
	}
	return result, nil
}

func optionalID(value string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
