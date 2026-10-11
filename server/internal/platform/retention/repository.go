package retention

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(queries *sqlc.Queries) *Repository {
	return &Repository{
		queries: queries,
	}
}
func (r *Repository) List(ctx context.Context, organizationID uuid.UUID) ([]Policy, error) {
	rows, err := r.queries.ListRetentionPolicies(ctx, organizationID)
	if err != nil {
		return nil, err
	}
	out := make([]Policy, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}
func (r *Repository) Upsert(ctx context.Context, organizationID uuid.UUID, resource string, req UpsertRequest) (Policy, error) {
	row, err := r.queries.UpsertRetentionPolicy(ctx, sqlc.UpsertRetentionPolicyParams{
		OrganizationID: organizationID,
		Resource:       resource,
		RetentionDays:  req.RetentionDays,
		Enabled:        req.Enabled,
	})
	if err != nil {
		return Policy{}, err
	}
	return fromRow(row), nil
}
func (r *Repository) Delete(ctx context.Context, organizationID uuid.UUID, resource string) error {
	return r.queries.DeleteRetentionPolicy(ctx, sqlc.DeleteRetentionPolicyParams{
		OrganizationID: organizationID,
		Resource:       resource,
	})
}
func (r *Repository) ListEnabled(ctx context.Context) ([]Policy, error) {
	rows, err := r.queries.ListEnabledRetentionPolicies(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Policy, 0, len(rows))
	for _, row := range rows {
		out = append(out, fromRow(row))
	}
	return out, nil
}
func (r *Repository) ListEffectiveRecordingRetention(
	ctx context.Context,
) ([]EffectiveRecordingRetention, error) {
	rows, err := r.queries.ListEffectiveRecordingRetention(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]EffectiveRecordingRetention, 0, len(rows))
	for _, row := range rows {
		out = append(out, EffectiveRecordingRetention{
			OrganizationID: row.OrganizationID,
			RetentionDays:  row.RetentionDays,
		})
	}
	return out, nil
}

func (r *Repository) ListExpiredRecordings(ctx context.Context, organizationID uuid.UUID, before time.Time, batchSize int32) ([]uuid.UUID, error) {
	return r.queries.ListExpiredRecordingsForRetention(ctx, sqlc.ListExpiredRecordingsForRetentionParams{
		OrganizationID: organizationID,
		CreatedBefore: pgtype.Timestamptz{
			Time:  before.UTC(),
			Valid: true,
		},
		BatchSize: batchSize,
	})
}
func fromRow(row sqlc.RetentionPolicy) Policy {
	return Policy{
		OrganizationID: row.OrganizationID,
		Resource:       row.Resource,
		RetentionDays:  row.RetentionDays,
		Enabled:        row.Enabled,
		CreatedAt:      pgconv.TimestamptzToTime(row.CreatedAt),
		UpdatedAt:      pgconv.TimestamptzToTime(row.UpdatedAt),
	}
}
