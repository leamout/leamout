package numbers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/modules/audit"
)

const managedNumberSelectionTTL = 10 * time.Minute

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	redis   *redisintegration.Client
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db, queries: sqlc.New(db)}
}

func (r *Repository) SetRedis(redis *redisintegration.Client) {
	r.redis = redis
}

func (r *Repository) SaveManagedSelection(
	ctx context.Context,
	organizationID uuid.UUID,
	candidate ManagedNumberCandidate,
) (string, error) {
	if r == nil || r.redis == nil {
		return "", fmt.Errorf("managed number selection store is unavailable")
	}
	if organizationID == uuid.Nil {
		return "", fmt.Errorf("organization id is required")
	}

	selectionID := "sel_" + uuid.NewString()
	payload, err := json.Marshal(candidate)
	if err != nil {
		return "", fmt.Errorf("encode managed number selection: %w", err)
	}
	key := "telecom:numbers:selection:" + organizationID.String() + ":" + selectionID
	if err := r.redis.Set(ctx, key, payload, managedNumberSelectionTTL); err != nil {
		return "", fmt.Errorf("store managed number selection: %w", err)
	}
	return selectionID, nil
}

func (r *Repository) CreateBYOC(
	ctx context.Context,
	organizationID uuid.UUID,
	req BYOCCreateRequest,
) (sqlc.PhoneNumber, error) {
	return r.queries.CreateBYOCPhoneNumber(ctx, sqlc.CreateBYOCPhoneNumberParams{
		OrganizationID: organizationID,
		Number:         req.Number,
		CountryCode:    req.CountryCode,
		VoiceEnabled:   req.VoiceEnabled,
		SmsEnabled:     req.SMSEnabled,
	})
}

func (r *Repository) CreateManaged(
	ctx context.Context,
	organizationID uuid.UUID,
	req ManagedCreateRequest,
) (sqlc.PhoneNumber, error) {
	providerID := req.ProviderID
	providerResourceID := req.ProviderResourceID
	return r.queries.CreateManagedPhoneNumber(ctx, sqlc.CreateManagedPhoneNumberParams{
		OrganizationID:      organizationID,
		Number:              req.Number,
		CountryCode:         req.CountryCode,
		CarrierConnectionID: req.CarrierConnectionID,
		ProviderID:          &providerID,
		ProviderResourceID:  &providerResourceID,
		VoiceEnabled:        req.VoiceEnabled,
		SmsEnabled:          req.SMSEnabled,
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

func (r *Repository) GetForRelease(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.PhoneNumber, error) {
	return r.queries.GetPhoneNumberForRelease(ctx, sqlc.GetPhoneNumberForReleaseParams{
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
		VoiceEnabled:   req.VoiceEnabled,
		SmsEnabled:     req.SMSEnabled,
	})
}

func (r *Repository) ReleaseBYOC(
	ctx context.Context,
	organizationID uuid.UUID,
	id uuid.UUID,
) (sqlc.PhoneNumber, error) {
	return r.queries.ReleaseBYOCPhoneNumber(ctx, sqlc.ReleaseBYOCPhoneNumberParams{
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
		CarrierConnectionID: &connectionID,
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
