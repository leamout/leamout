package numbers

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	req CreateRequest,
) (sqlc.PhoneNumber, error) {
	return r.queries.CreatePhoneNumber(ctx, sqlc.CreatePhoneNumberParams{
		OrganizationID: organizationID,
		Number:         req.Number,
		CountryCode:    req.CountryCode,
		VoiceEnabled:   req.VoiceEnabled,
		SmsEnabled:     req.SMSEnabled,
	})
}

func (r *Repository) List(
	ctx context.Context,
	organizationID uuid.UUID,
) ([]sqlc.PhoneNumber, error) {
	return r.queries.ListPhoneNumbersByOrganizationID(ctx, organizationID)
}

func (r *Repository) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.PhoneNumber, error) {
	return r.queries.GetPhoneNumberByID(ctx, sqlc.GetPhoneNumberByIDParams{
		ID:             id,
		OrganizationID: organizationID,
	})
}

func (r *Repository) Update(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
	req UpdateRequest,
) (sqlc.PhoneNumber, error) {
	return r.queries.UpdatePhoneNumber(ctx, sqlc.UpdatePhoneNumberParams{
		ID:             id,
		OrganizationID: organizationID,
		CountryCode:    req.CountryCode,
		VoiceEnabled:   req.VoiceEnabled,
		SmsEnabled:     req.SMSEnabled,
	})
}

func (r *Repository) Disable(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) error {
	return r.queries.DisablePhoneNumber(ctx, sqlc.DisablePhoneNumberParams{
		ID:             id,
		OrganizationID: organizationID,
	})
}
