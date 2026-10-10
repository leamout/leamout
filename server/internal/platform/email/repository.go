package email

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/integrations/ses"
)

type Repository struct{ queries *sqlc.Queries }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{queries: sqlc.New(pool)} }
func (r *Repository) Claim(ctx context.Context, lease uuid.UUID) (Delivery, error) {
	if err := r.queries.ExpireEmailDeliveries(ctx); err != nil {
		return Delivery{}, err
	}
	if err := r.queries.FailExhaustedEmailDeliveries(ctx); err != nil {
		return Delivery{}, err
	}
	row, err := r.queries.ClaimEmailDelivery(ctx, &lease)
	if err != nil {
		return Delivery{}, err
	}
	payload := ""
	if row.EncryptedData != nil {
		payload = *row.EncryptedData
	}
	return Delivery{ID: row.ID, To: row.Recipient, Template: row.Template, Payload: payload, Attempts: int(row.Attempts), ExpiresAt: pgconv.TimestamptzToTime(row.ExpiresAt)}, nil
}
func (r *Repository) Complete(ctx context.Context, id, lease uuid.UUID, result ses.Result) error {
	return r.queries.CompleteEmailDelivery(ctx, sqlc.CompleteEmailDeliveryParams{ID: id, LockToken: &lease, ProviderMessageID: &result.MessageID})
}
func (r *Repository) Fail(ctx context.Context, d Delivery, lease uuid.UUID, code string, permanent bool) error {
	return r.queries.FailEmailDelivery(ctx, sqlc.FailEmailDeliveryParams{ID: d.ID, LockToken: &lease, Terminal: permanent || d.Attempts >= 5, ErrorCode: &code, DelaySeconds: int32(1 << uint(d.Attempts))})
}

// Ready checks only delivery state and expiry, independent of the originating domain.
func (r *Repository) Ready(ctx context.Context, id, lease uuid.UUID) (bool, error) {
	return r.queries.IsEmailDeliveryReady(ctx, sqlc.IsEmailDeliveryReadyParams{ID: id, LockToken: &lease})
}
