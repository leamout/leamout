package numbers

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(q *sqlc.Queries) *Repository { return &Repository{q} }
func (r *Repository) Create(c context.Context, org uuid.UUID, req CreateRequest) (sqlc.PhoneNumber, error) {
	return r.queries.CreatePhoneNumber(c, sqlc.CreatePhoneNumberParams{OrganizationID: org, Number: req.Number, CountryCode: req.CountryCode, VoiceEnabled: req.VoiceEnabled, SmsEnabled: req.SMSEnabled})
}
func (r *Repository) List(c context.Context, org uuid.UUID) ([]sqlc.PhoneNumber, error) {
	return r.queries.ListPhoneNumbersByOrganizationID(c, org)
}
func (r *Repository) Get(c context.Context, org, id uuid.UUID) (sqlc.PhoneNumber, error) {
	return r.queries.GetPhoneNumberByID(c, sqlc.GetPhoneNumberByIDParams{ID: id, OrganizationID: org})
}
func (r *Repository) Update(c context.Context, org, id uuid.UUID, req UpdateRequest) (sqlc.PhoneNumber, error) {
	return r.queries.UpdatePhoneNumber(c, sqlc.UpdatePhoneNumberParams{ID: id, OrganizationID: org, CountryCode: req.CountryCode, VoiceEnabled: req.VoiceEnabled, SmsEnabled: req.SMSEnabled})
}
func (r *Repository) Disable(c context.Context, org, id uuid.UUID) error {
	return r.queries.DisablePhoneNumber(c, sqlc.DisablePhoneNumberParams{ID: id, OrganizationID: org})
}
