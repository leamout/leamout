package licensing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// SubscriptionAuthority is the subscription state needed to resolve a
// subscription-backed license without leaking billing-provider details.
type SubscriptionAuthority struct {
	ID       uuid.UUID
	PlanID   uuid.UUID
	Status   string
	StartsAt time.Time
	EndsAt   *time.Time
}

// AuthorityStore loads organization-scoped state required for issuance.
type AuthorityStore interface {
	GetLicense(context.Context, uuid.UUID, uuid.UUID) (License, error)
	GetSubscription(context.Context, uuid.UUID, uuid.UUID) (SubscriptionAuthority, error)
}

type SQLRepository struct {
	queries *sqlc.Queries
}

func NewSQLRepository(db *pgxpool.Pool) *SQLRepository {
	if db == nil {
		panic("licensing: database is required")
	}
	return &SQLRepository{queries: sqlc.New(db)}
}

func (r *SQLRepository) GetLicense(ctx context.Context, organizationID, licenseID uuid.UUID) (License, error) {
	row, err := r.queries.GetLicense(ctx, sqlc.GetLicenseParams{OrganizationID: organizationID, ID: licenseID})
	if err != nil {
		return License{}, fmt.Errorf("get license: %w", err)
	}
	return License{
		ID: row.ID.String(), OrganizationID: row.OrganizationID.String(), SubscriptionID: uuidString(row.SubscriptionID),
		Status: Status(row.Status), MaxDeployments: int(row.MaxDeployments), SigningKeyID: stringValue(row.SigningKeyID),
		IssuedAt: pgconv.TimestamptzToTime(row.IssuedAt), ExpiresAt: pgconv.TimestamptzToTimePtr(row.ExpiresAt),
	}, nil
}

func (r *SQLRepository) GetSubscription(ctx context.Context, organizationID, subscriptionID uuid.UUID) (SubscriptionAuthority, error) {
	row, err := r.queries.GetSubscription(ctx, sqlc.GetSubscriptionParams{OrganizationID: organizationID, ID: subscriptionID})
	if err != nil {
		return SubscriptionAuthority{}, fmt.Errorf("get subscription authority: %w", err)
	}
	return SubscriptionAuthority{
		ID: row.ID, PlanID: row.PlanID, Status: row.Status,
		StartsAt: pgconv.TimestamptzToTime(row.StartsAt), EndsAt: pgconv.TimestamptzToTimePtr(row.EndsAt),
	}, nil
}

func uuidString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
