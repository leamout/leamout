package metering

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

// Repository persists meters and immutable usage events through SQLC.
type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) GetMeter(ctx context.Context, key string) (Meter, error) {
	row, err := r.queries.GetMeterByKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Meter{}, ErrMeterNotFound
		}
		return Meter{}, err
	}
	return meterFromRow(row), nil
}

func (r *Repository) CreateUsageEvent(ctx context.Context, organizationID uuid.UUID, input RecordInput) (UsageEvent, error) {
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
		return UsageEvent{}, err
	}
	return usageEventFromRow(row), nil
}

func (r *Repository) GetUsageEventByIdempotencyKey(ctx context.Context, organizationID uuid.UUID, key string) (UsageEvent, error) {
	row, err := r.queries.GetUsageEventByIdempotencyKey(ctx, sqlc.GetUsageEventByIdempotencyKeyParams{
		OrganizationID: organizationID,
		IdempotencyKey: key,
	})
	if err != nil {
		return UsageEvent{}, err
	}
	return usageEventFromRow(row), nil
}

func meterFromRow(row sqlc.Meter) Meter {
	return Meter{
		ID:        row.ID,
		Key:       row.Key,
		Name:      row.Name,
		Unit:      row.Unit,
		Active:    row.Active,
		CreatedAt: pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt: pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}

func usageEventFromRow(row sqlc.UsageEvent) UsageEvent {
	return UsageEvent{
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
