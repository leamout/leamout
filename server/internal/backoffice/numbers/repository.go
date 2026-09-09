package numbers

import (
	"context"

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
			ID:           row.ID,
			Organization: row.OrganizationName,
			Number:       row.Number,
			CountryCode:  row.CountryCode,
			Mode:         row.ProvisioningMode,
			Provider:     row.ProviderName,
			Voice:        row.VoiceEnabled,
			SMS:          row.SmsEnabled,
			Status:       row.Status,
			CreatedAt:    row.CreatedAt,
		})
	}
	return numbers, nil
}
