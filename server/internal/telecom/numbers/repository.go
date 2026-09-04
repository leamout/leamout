package numbers

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/modules/audit"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, queries: sqlc.New(db)}
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

func (r *Repository) SetCarrierConnection(ctx context.Context, organizationID, id, connectionID uuid.UUID, event audit.Event) (sqlc.PhoneNumber, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.PhoneNumber{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	number, err := r.queries.WithTx(tx).SetBYOCPhoneNumberCarrierConnection(ctx, sqlc.SetBYOCPhoneNumberCarrierConnectionParams{
		ID:                  id,
		OrganizationID:      organizationID,
		CarrierConnectionID: connectionID,
	})
	if err != nil {
		return sqlc.PhoneNumber{}, err
	}
	if err := audit.Insert(ctx, tx, event); err != nil {
		return sqlc.PhoneNumber{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return sqlc.PhoneNumber{}, err
	}
	return number, nil
}
