package usage

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository persists immutable usage events through SQLC.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) CreateEvent(
	ctx context.Context,
	organizationID uuid.UUID,
	input RecordInput,
) (Event, error) {
	row, err := r.queries.CreateUsageEvent(ctx, sqlc.CreateUsageEventParams{
		SubscriptionID: input.SubscriptionID,
		Quantity:       input.Quantity,
		SourceType:     input.SourceType,
		SourceID:       input.SourceID,
		IdempotencyKey: input.IdempotencyKey,
		Dimensions:     input.Dimensions,
		OccurredAt:     pgconv.NullableTimestamptz(&input.OccurredAt),
		MeterID:        input.MeterID,
		OrganizationID: organizationID,
	})
	if err != nil {
		return Event{}, err
	}

	return eventFromRow(row), nil
}

func (r *Repository) GetByIdempotencyKey(
	ctx context.Context,
	organizationID uuid.UUID,
	key string,
) (Event, error) {
	row, err := r.queries.GetUsageEventByIdempotencyKey(
		ctx,
		sqlc.GetUsageEventByIdempotencyKeyParams{
			OrganizationID: organizationID,
			IdempotencyKey: key,
		},
	)
	if err != nil {
		return Event{}, err
	}

	return eventFromRow(row), nil
}

func eventFromRow(row sqlc.UsageEvent) Event {
	return Event{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		SubscriptionID: row.SubscriptionID,
		MeterID:        row.MeterID,
		Quantity:       row.Quantity,
		SourceType:     row.SourceType,
		SourceID:       row.SourceID,
		IdempotencyKey: row.IdempotencyKey,
		Dimensions:     row.Dimensions,
		OccurredAt:     pgconv.TimestamptzToTime(row.OccurredAt),
		CreatedAt:      pgconv.TimestamptzToTime(row.CreatedAt),
	}
}
