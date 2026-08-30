package deployments

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

var (
	ErrLicenseUnavailable     = errors.New("license is unavailable for activation")
	ErrDeploymentLimitReached = errors.New("license deployment limit reached")
	ErrDeploymentDeactivated  = errors.New("deployment was previously deactivated")
)

// Repository persists deployment state. Activation serializes on the license
// row so capacity checks and insertion cannot race for the same license.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	if db == nil {
		panic("deployments: database is required")
	}
	return &Repository{db: db}
}

// Activate creates a deployment or returns its existing active record. A
// deactivated deployment ID cannot be reused; callers must choose a new stable
// deployment identity.
func (r *Repository) Activate(ctx context.Context, request ActivateRequest) (Deployment, error) {
	if err := ValidateActivate(request); err != nil {
		return Deployment{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Deployment{}, fmt.Errorf("begin deployment activation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	created, err := activate(ctx, transactionStore{queries: sqlc.New(tx)}, request)
	if err != nil {
		return Deployment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Deployment{}, fmt.Errorf("commit deployment activation: %w", err)
	}
	return created, nil
}

type activationStore interface {
	lockLicense(context.Context, ActivateRequest) (int32, error)
	get(context.Context, ActivateRequest) (Deployment, error)
	countActive(context.Context, ActivateRequest) (int64, error)
	insert(context.Context, ActivateRequest) (Deployment, error)
}

func activate(ctx context.Context, store activationStore, request ActivateRequest) (Deployment, error) {
	maxDeployments, err := store.lockLicense(ctx, request)
	if err != nil {
		return Deployment{}, err
	}
	existing, err := store.get(ctx, request)
	if err == nil {
		if existing.Status == StatusActive {
			return existing, nil
		}
		return Deployment{}, ErrDeploymentDeactivated
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Deployment{}, fmt.Errorf("get existing deployment: %w", err)
	}
	active, err := store.countActive(ctx, request)
	if err != nil {
		return Deployment{}, fmt.Errorf("count active deployments: %w", err)
	}
	if active >= int64(maxDeployments) {
		return Deployment{}, ErrDeploymentLimitReached
	}
	created, err := store.insert(ctx, request)
	if err != nil {
		return Deployment{}, fmt.Errorf("insert deployment: %w", err)
	}
	return created, nil
}

type transactionStore struct {
	queries *sqlc.Queries
}

func (s transactionStore) lockLicense(ctx context.Context, request ActivateRequest) (int32, error) {
	maxDeployments, err := s.queries.LockLicenseForDeploymentActivation(ctx, sqlc.LockLicenseForDeploymentActivationParams{
		LicenseID:      request.LicenseID,
		OrganizationID: request.OrganizationID,
		ActivationTime: pgtype.Timestamptz{Time: request.At.UTC(), Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrLicenseUnavailable
	}
	if err != nil {
		return 0, fmt.Errorf("lock license for deployment activation: %w", err)
	}
	return maxDeployments, nil
}

func (s transactionStore) countActive(ctx context.Context, request ActivateRequest) (int64, error) {
	return s.queries.CountActiveDeploymentsByLicense(ctx, sqlc.CountActiveDeploymentsByLicenseParams{
		LicenseID: request.LicenseID, OrganizationID: request.OrganizationID,
	})
}

func (s transactionStore) get(ctx context.Context, request ActivateRequest) (Deployment, error) {
	value, err := s.queries.GetDeploymentForActivation(ctx, sqlc.GetDeploymentForActivationParams{
		LicenseID: request.LicenseID, DeploymentID: request.DeploymentID,
	})
	return fromRow(value), err
}

func (s transactionStore) insert(ctx context.Context, request ActivateRequest) (Deployment, error) {
	value, err := s.queries.CreateDeploymentAt(ctx, sqlc.CreateDeploymentAtParams{
		LicenseID: request.LicenseID, DeploymentID: request.DeploymentID, Name: request.Name,
		ActivatedAt: pgtype.Timestamptz{Time: request.At.UTC(), Valid: true},
	})
	return fromRow(value), err
}

func fromRow(value sqlc.Deployment) Deployment {
	return Deployment{
		ID: value.ID, LicenseID: value.LicenseID, DeploymentID: value.DeploymentID, Name: value.Name,
		Status: Status(value.Status), ActivatedAt: pgconv.TimestamptzToTime(value.ActivatedAt),
		LastSeenAt: pgconv.TimestamptzToTimePtr(value.LastSeenAt), DeactivatedAt: pgconv.TimestamptzToTimePtr(value.DeactivatedAt),
		CreatedAt: pgconv.TimestamptzToTime(value.CreatedAt), UpdatedAt: pgconv.TimestamptzToTime(value.UpdatedAt),
	}
}
