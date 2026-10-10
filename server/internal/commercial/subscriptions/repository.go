package subscriptions

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{queries: queries}
}

func (r *Repository) Get(ctx context.Context, organizationID uuid.UUID) (Subscription, error) {
	row, err := r.queries.GetOrganizationSubscription(ctx, organizationID)
	if err != nil {
		return Subscription{}, err
	}
	return fromRow(row), nil
}

func (r *Repository) Upsert(
	ctx context.Context,
	id uuid.UUID,
	organizationID uuid.UUID,
	req SyncRequest,
) (Subscription, error) {
	row, err := r.queries.UpsertOrganizationSubscription(ctx, sqlc.UpsertOrganizationSubscriptionParams{
		ID:                 id,
		OrganizationID:     organizationID,
		PlanID:             req.PlanID,
		Status:             req.Status,
		CurrentPeriodStart: pgconv.NullableTimestamptz(req.CurrentPeriodStart),
		CurrentPeriodEnd:   pgconv.NullableTimestamptz(req.CurrentPeriodEnd),
		TrialEndsAt:        pgconv.NullableTimestamptz(req.TrialEndsAt),
		CancelAtPeriodEnd:  req.CancelAtPeriodEnd,
		CanceledAt:         pgconv.NullableTimestamptz(req.CanceledAt),
	})
	if err != nil {
		return Subscription{}, err
	}
	return fromRow(row), nil
}

func fromRow(row sqlc.Subscription) Subscription {
	return Subscription{
		ID:                 row.ID,
		OrganizationID:     row.OrganizationID,
		PlanID:             row.PlanID,
		Status:             row.Status,
		CurrentPeriodStart: pgconv.TimestamptzToTimePtr(row.CurrentPeriodStart),
		CurrentPeriodEnd:   pgconv.TimestamptzToTimePtr(row.CurrentPeriodEnd),
		TrialEndsAt:        pgconv.TimestamptzToTimePtr(row.TrialEndsAt),
		CancelAtPeriodEnd:  row.CancelAtPeriodEnd,
		CanceledAt:         pgconv.TimestamptzToTimePtr(row.CanceledAt),
		CreatedAt:          pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:          pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}
