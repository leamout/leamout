package licensing

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

const serializableAttempts = 3

// Repository persists licenses and deployments through SQLC within an organization boundary.
type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

func (r *Repository) Create(ctx context.Context, organizationID, subscriptionID uuid.UUID, maxDeployments int32, signingKeyID *string, issuedAt time.Time, expiresAt *time.Time) (License, error) {
	status := string(StatusPending)
	row, err := r.queries.CreateLicense(ctx, sqlc.CreateLicenseParams{
		SubscriptionID: &subscriptionID,
		Status:         &status,
		MaxDeployments: &maxDeployments,
		SigningKeyID:   signingKeyID,
		IssuedAt:       pgconv.NullableTimestamptz(&issuedAt),
		ExpiresAt:      pgconv.NullableTimestamptz(expiresAt),
		OrganizationID: organizationID,
	})
	if err != nil {
		return License{}, mapLicenseWriteError(err)
	}
	return licenseFromRow(row), nil
}

func (r *Repository) Get(ctx context.Context, organizationID, id uuid.UUID) (License, error) {
	row, err := r.queries.GetLicense(ctx, sqlc.GetLicenseParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return License{}, mapLicenseReadError(err)
	}
	return licenseFromRow(row), nil
}

func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]License, error) {
	rows, err := r.queries.ListLicensesByOrganization(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	result := make([]License, 0, len(rows))
	for _, row := range rows {
		result = append(result, licenseFromRow(row))
	}
	return result, nil
}

func (r *Repository) Transition(ctx context.Context, organizationID, id uuid.UUID, target Status) (License, error) {
	for attempt := 0; attempt < serializableAttempts; attempt++ {
		resolved, retry, err := r.transitionOnce(ctx, organizationID, id, target)
		if !retry {
			return resolved, err
		}
	}
	return License{}, ErrActivationConflict
}

func (r *Repository) transitionOnce(ctx context.Context, organizationID, id uuid.UUID, target Status) (License, bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return License{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	currentRow, err := queries.GetLicense(ctx, sqlc.GetLicenseParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return License{}, false, mapLicenseReadError(err)
	}
	current := licenseFromRow(currentRow)
	if err := validateTransition(current.Status, target); err != nil {
		return License{}, false, err
	}
	if current.Status == target {
		return current, false, nil
	}
	_, err = queries.UpdateLicenseStatus(ctx, sqlc.UpdateLicenseStatusParams{
		Status: targetString(target), OrganizationID: organizationID, ID: id,
	})
	if err != nil {
		if isSerializationFailure(err) {
			return License{}, true, nil
		}
		return License{}, false, mapLicenseWriteError(err)
	}
	row, err := queries.GetLicense(ctx, sqlc.GetLicenseParams{OrganizationID: organizationID, ID: id})
	if err != nil {
		return License{}, false, mapLicenseReadError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		if isSerializationFailure(err) {
			return License{}, true, nil
		}
		return License{}, false, err
	}
	return licenseFromRow(row), false, nil
}

func (r *Repository) UpdateExpiration(ctx context.Context, organizationID, id uuid.UUID, expiresAt *time.Time) (License, error) {
	_, err := r.queries.UpdateLicenseExpiration(ctx, sqlc.UpdateLicenseExpirationParams{
		ExpiresAt: pgconv.NullableTimestamptz(expiresAt), OrganizationID: organizationID, ID: id,
	})
	if err != nil {
		return License{}, mapLicenseWriteError(err)
	}
	return r.Get(ctx, organizationID, id)
}

func (r *Repository) ActivateDeployment(ctx context.Context, organizationID, licenseID uuid.UUID, input ActivateDeploymentInput, at time.Time) (Deployment, error) {
	for attempt := 0; attempt < serializableAttempts; attempt++ {
		deployment, retry, err := r.activateDeploymentOnce(ctx, organizationID, licenseID, input, at)
		if !retry {
			return deployment, err
		}
	}
	return Deployment{}, ErrActivationConflict
}

func (r *Repository) activateDeploymentOnce(ctx context.Context, organizationID, licenseID uuid.UUID, input ActivateDeploymentInput, at time.Time) (Deployment, bool, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return Deployment{}, false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	existing, err := queries.GetDeployment(ctx, sqlc.GetDeploymentParams{
		LicenseID: licenseID, DeploymentID: input.DeploymentID, OrganizationID: organizationID,
	})
	if err == nil {
		deployment := deploymentFromRow(organizationID, existing)
		if deployment.Status == DeploymentStatusDeactivated {
			return Deployment{}, false, ErrDeploymentInactive
		}
		return deployment, false, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return Deployment{}, false, err
	}

	licenseRow, err := queries.GetLicense(ctx, sqlc.GetLicenseParams{OrganizationID: organizationID, ID: licenseID})
	if err != nil {
		return Deployment{}, false, mapLicenseReadError(err)
	}
	license := licenseFromRow(licenseRow)
	if license.Status != StatusActive || (license.ExpiresAt != nil && !license.ExpiresAt.After(at)) {
		return Deployment{}, false, ErrLicenseUnavailable
	}
	count, err := queries.CountActiveDeploymentsByLicense(ctx, sqlc.CountActiveDeploymentsByLicenseParams{
		LicenseID: licenseID, OrganizationID: organizationID,
	})
	if err != nil {
		return Deployment{}, false, err
	}
	if count >= int64(license.MaxDeployments) {
		return Deployment{}, false, ErrDeploymentLimitReached
	}
	row, err := queries.CreateDeployment(ctx, sqlc.CreateDeploymentParams{
		DeploymentID: input.DeploymentID, Name: input.Name, LicenseID: licenseID, OrganizationID: organizationID,
	})
	if err != nil {
		if isSerializationFailure(err) {
			return Deployment{}, true, nil
		}
		return Deployment{}, false, mapDeploymentWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		if isSerializationFailure(err) {
			return Deployment{}, true, nil
		}
		return Deployment{}, false, err
	}
	return deploymentFromRow(organizationID, row), false, nil
}

func (r *Repository) ListDeployments(ctx context.Context, organizationID, licenseID uuid.UUID) ([]Deployment, error) {
	rows, err := r.queries.ListDeploymentsByLicense(ctx, sqlc.ListDeploymentsByLicenseParams{
		LicenseID: licenseID, OrganizationID: organizationID,
	})
	if err != nil {
		return nil, err
	}
	result := make([]Deployment, 0, len(rows))
	for _, row := range rows {
		result = append(result, deploymentFromRow(organizationID, row))
	}
	return result, nil
}

func (r *Repository) TouchDeployment(ctx context.Context, organizationID, licenseID uuid.UUID, deploymentID string, at time.Time) (Deployment, error) {
	_, err := r.queries.TouchDeployment(ctx, sqlc.TouchDeploymentParams{
		LastSeenAt: pgconv.NullableTimestamptz(&at), LicenseID: licenseID, DeploymentID: deploymentID, OrganizationID: organizationID,
	})
	if err != nil {
		return Deployment{}, mapDeploymentWriteError(err)
	}
	return r.getDeployment(ctx, organizationID, licenseID, deploymentID)
}

func (r *Repository) DeactivateDeployment(ctx context.Context, organizationID, licenseID uuid.UUID, deploymentID string) (Deployment, error) {
	_, err := r.queries.DeactivateDeployment(ctx, sqlc.DeactivateDeploymentParams{
		LicenseID: licenseID, DeploymentID: deploymentID, OrganizationID: organizationID,
	})
	if err != nil {
		return Deployment{}, mapDeploymentWriteError(err)
	}
	return r.getDeployment(ctx, organizationID, licenseID, deploymentID)
}

func (r *Repository) getDeployment(ctx context.Context, organizationID, licenseID uuid.UUID, deploymentID string) (Deployment, error) {
	row, err := r.queries.GetDeployment(ctx, sqlc.GetDeploymentParams{
		LicenseID: licenseID, DeploymentID: deploymentID, OrganizationID: organizationID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Deployment{}, ErrDeploymentNotFound
		}
		return Deployment{}, err
	}
	return deploymentFromRow(organizationID, row), nil
}

func licenseFromRow(row sqlc.License) License {
	return License{
		ID: row.ID, OrganizationID: row.OrganizationID, SubscriptionID: row.SubscriptionID,
		Status: Status(row.Status), MaxDeployments: row.MaxDeployments, SigningKeyID: row.SigningKeyID,
		IssuedAt: pgconv.TimestamptzToTime(row.IssuedAt), ExpiresAt: pgconv.TimestamptzToTimePtr(row.ExpiresAt),
		CreatedAt: pgconv.TimestamptzToTime(row.CreatedAt), UpdatedAt: pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func deploymentFromRow(organizationID uuid.UUID, row sqlc.Deployment) Deployment {
	return Deployment{
		ID: row.ID, OrganizationID: organizationID, LicenseID: row.LicenseID, DeploymentID: row.DeploymentID,
		Name: row.Name, Status: DeploymentStatus(row.Status), ActivatedAt: pgconv.TimestamptzToTime(row.ActivatedAt),
		LastSeenAt: pgconv.TimestamptzToTimePtr(row.LastSeenAt), DeactivatedAt: pgconv.TimestamptzToTimePtr(row.DeactivatedAt),
		CreatedAt: pgconv.TimestamptzToTime(row.CreatedAt), UpdatedAt: pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func targetString(status Status) string { return string(status) }

func mapLicenseReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrLicenseNotFound
	}
	return err
}

func mapLicenseWriteError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrLicenseUnavailable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23514" {
		return ErrInvalidExpiration
	}
	return err
}

func mapDeploymentWriteError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDeploymentNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "uq_deployments_license_deployment" {
		return ErrActivationConflict
	}
	return err
}

func isSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "40001"
}
