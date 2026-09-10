package numbers

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository owns cross-tenant Backoffice reads for phone numbers.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) List(ctx context.Context) ([]Number, error) {
	if r == nil || r.queries == nil {
		return nil, nil
	}
	rows, err := r.queries.ListBackofficePhoneNumbers(ctx)
	if err != nil {
		return nil, err
	}
	numbers := make([]Number, 0, len(rows))
	for _, row := range rows {
		numbers = append(numbers, Number{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Organization:   row.OrganizationName,
			Number:         row.Number,
			CountryCode:    row.CountryCode,
			Mode:           row.ProvisioningMode,
			Provider:       row.ProviderName,
			Voice:          row.VoiceEnabled,
			SMS:            row.SmsEnabled,
			Status:         row.Status,
			CreatedAt:      row.CreatedAt,
		})
	}
	return numbers, nil
}

func (r *Repository) Get(ctx context.Context, numberID uuid.UUID) (Detail, error) {
	row, err := r.queries.GetBackofficePhoneNumber(ctx, numberID)
	if err != nil {
		return Detail{}, err
	}
	return Detail{
		PhoneNumber: Number{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Organization:   row.OrganizationName,
			Number:         row.Number,
			CountryCode:    row.CountryCode,
			Mode:           row.ProvisioningMode,
			Provider:       row.ProviderName,
			Voice:          row.VoiceEnabled,
			SMS:            row.SmsEnabled,
			Status:         row.Status,
			CreatedAt:      row.CreatedAt,
		},
		CarrierConnectionID: row.CarrierConnectionID, CarrierConnection: row.CarrierConnectionName,
		ProviderID: row.ProviderID, ProviderResourceID: row.ProviderResourceID,
		ErrorCode: row.ErrorCode, ErrorMessage: row.ErrorMessage, UpdatedAt: row.UpdatedAt,
	}, nil
}
