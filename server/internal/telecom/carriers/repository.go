package carriers

import (
	"context"
	"errors"
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/security/encryption"
)

type Repository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
	cipher  *encryption.Cipher
}

func NewRepository(db *pgxpool.Pool, cipher *encryption.Cipher) *Repository {
	return &Repository{db: db, queries: sqlc.New(db), cipher: cipher}
}

func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{db: r.db, queries: r.queries.WithTx(tx), cipher: r.cipher}
}

func (r *Repository) Create(ctx context.Context, arg sqlc.CreateCarrierConnectionParams) (sqlc.CarrierConnection, error) {
	return r.queries.CreateCarrierConnection(ctx, arg)
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]sqlc.CarrierConnection, error) {
	return r.queries.ListCarrierConnections(ctx, &organizationID)
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (sqlc.CarrierConnection, error) {
	row, err := r.queries.GetCarrierConnectionByID(ctx, sqlc.GetCarrierConnectionByIDParams{
		ID:             id,
		OrganizationID: &organizationID,
	})
	if err != nil {
		return sqlc.CarrierConnection{}, err
	}
	return carrierConnectionFromGet(row), nil
}

func (r *Repository) Update(ctx context.Context, arg sqlc.UpdateCarrierConnectionParams) (sqlc.CarrierConnection, error) {
	return r.queries.UpdateCarrierConnection(ctx, arg)
}

func (r *Repository) UpdateOutboundAuth(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
	request OutboundAuthRequest,
) (sqlc.CarrierConnection, error) {
	if r.cipher == nil {
		return sqlc.CarrierConnection{}, errors.New("carrier credential cipher is required")
	}
	secretCiphertext, err := r.cipher.EncryptString(request.Secret)
	if err != nil {
		return sqlc.CarrierConnection{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.CarrierConnection{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	updated, err := queries.UpdateCarrierConnectionOutboundAuth(ctx, sqlc.UpdateCarrierConnectionOutboundAuthParams{
		OrganizationID:      &organizationID,
		ID:                  connectionID,
		OutboundAuthMethod:  request.Method,
		AuthUsername:        request.Username,
		AuthRealm:           request.Realm,
		AuthSecretCiphertext: &secretCiphertext,
	})
	if err != nil {
		return sqlc.CarrierConnection{}, err
	}

	if err := syncDigestCredential(ctx, queries, updated, request.Secret); err != nil {
		return sqlc.CarrierConnection{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return sqlc.CarrierConnection{}, err
	}
	return updated, nil
}

func (r *Repository) DeleteOutboundAuth(ctx context.Context, organizationID, connectionID uuid.UUID) (sqlc.CarrierConnection, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return sqlc.CarrierConnection{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	updated, err := queries.ClearCarrierConnectionOutboundAuth(ctx, sqlc.ClearCarrierConnectionOutboundAuthParams{
		ID:             connectionID,
		OrganizationID: &organizationID,
	})
	if err != nil {
		return sqlc.CarrierConnection{}, err
	}
	if err := queries.DeleteCarrierDigestCredentialsByConnection(ctx, connectionID); err != nil {
		return sqlc.CarrierConnection{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return sqlc.CarrierConnection{}, err
	}
	return updated, nil
}

func (r *Repository) CreateSourceIP(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
	cidr netip.Prefix,
) (sqlc.CarrierConnectionSourceIp, error) {
	return r.queries.CreateCarrierConnectionSourceIP(ctx, sqlc.CreateCarrierConnectionSourceIPParams{
		CarrierConnectionID: connectionID,
		OrganizationID:      &organizationID,
		Cidr:                cidr,
	})
}

func (r *Repository) ListSourceIPs(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
) ([]sqlc.CarrierConnectionSourceIp, error) {
	return r.queries.ListCarrierConnectionSourceIPs(ctx, sqlc.ListCarrierConnectionSourceIPsParams{
		CarrierConnectionID: connectionID,
		OrganizationID:      &organizationID,
	})
}

func (r *Repository) GetSourceIP(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
	id uuid.UUID,
) (sqlc.CarrierConnectionSourceIp, error) {
	return r.queries.GetCarrierConnectionSourceIPByID(ctx, sqlc.GetCarrierConnectionSourceIPByIDParams{
		ID:                  id,
		CarrierConnectionID: connectionID,
		OrganizationID:      &organizationID,
	})
}

func (r *Repository) DeleteSourceIP(
	ctx context.Context,
	organizationID uuid.UUID,
	connectionID uuid.UUID,
	id uuid.UUID,
) error {
	return r.queries.DeleteCarrierConnectionSourceIP(
		ctx,
		sqlc.DeleteCarrierConnectionSourceIPParams{
			ID:                  id,
			CarrierConnectionID: connectionID,
			OrganizationID:      &organizationID,
		},
	)
}

func (r *Repository) ListProviders(ctx context.Context) ([]sqlc.CarrierProvider, error) {
	return r.queries.ListCarrierProviders(ctx)
}

func (r *Repository) GetProvider(ctx context.Context, id uuid.UUID) (sqlc.CarrierProvider, error) {
	return r.queries.GetCarrierProviderByID(ctx, id)
}

func (r *Repository) ListConnectionTrunks(ctx context.Context, organizationID, connectionID uuid.UUID) ([]sqlc.Trunk, error) {
	return r.queries.ListTrunksByCarrierConnectionID(ctx, sqlc.ListTrunksByCarrierConnectionIDParams{
		CarrierConnectionID: &connectionID,
		OrganizationID:      &organizationID,
	})
}

func (r *Repository) ListTrunkEndpoints(ctx context.Context, organizationID, trunkID uuid.UUID) ([]sqlc.TrunkEndpoint, error) {
	return r.queries.ListTrunkEndpoints(ctx, sqlc.ListTrunkEndpointsParams{
		TrunkID:        trunkID,
		OrganizationID: &organizationID,
	})
}
