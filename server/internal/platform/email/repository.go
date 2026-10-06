package email

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Claim uses a lease token to prevent stale workers from completing another worker's job.
func (r *Repository) Claim(ctx context.Context, lease uuid.UUID) (Delivery, error) {
	_, err := r.pool.Exec(ctx, `UPDATE email_deliveries d SET status='expired', encrypted_data=NULL, locked_at=NULL, lock_token=NULL
 WHERE status IN ('pending','sending') AND (expires_at <= now() OR (template='otp' AND NOT EXISTS (
 SELECT 1 FROM auth_challenges c JOIN auth_transactions t ON t.id=c.auth_transaction_id
 WHERE c.id=d.challenge_id AND c.consumed_at IS NULL AND c.expires_at>now() AND c.attempts<c.max_attempts
 AND t.expires_at>now() AND t.state NOT IN ('authenticated','expired'))))`)
	if err != nil {
		return Delivery{}, err
	}
	_, err = r.pool.Exec(ctx, `UPDATE email_deliveries SET status='failed', encrypted_data=NULL, lock_token=NULL, locked_at=NULL, last_error_code='attempts_exhausted' WHERE attempts>=5 AND (status='pending' OR (status='sending' AND locked_at<now()-interval '60 seconds'))`)
	if err != nil {
		return Delivery{}, err
	}
	var d Delivery
	err = r.pool.QueryRow(ctx, `WITH candidate AS (
 SELECT id FROM email_deliveries WHERE expires_at>now() AND attempts<5 AND
 ((status='pending' AND available_at<=now()) OR (status='sending' AND locked_at<now()-interval '60 seconds'))
 ORDER BY available_at FOR UPDATE SKIP LOCKED LIMIT 1)
 UPDATE email_deliveries d SET status='sending', attempts=attempts+1, locked_at=now(), lock_token=$1
 FROM candidate WHERE d.id=candidate.id RETURNING d.id,d.recipient,d.template,d.encrypted_data,d.attempts,d.expires_at`, lease).Scan(&d.ID, &d.To, &d.Template, &d.Payload, &d.Attempts, &d.ExpiresAt)
	return d, err
}
func (r *Repository) Complete(ctx context.Context, id, lease uuid.UUID, result Result) error {
	_, err := r.pool.Exec(ctx, `UPDATE email_deliveries SET status='sent', encrypted_data=NULL, provider_message_id=$3, sent_at=now(), locked_at=NULL, lock_token=NULL WHERE id=$1 AND lock_token=$2 AND status='sending'`, id, lease, result.MessageID)
	return err
}
func (r *Repository) Fail(ctx context.Context, d Delivery, lease uuid.UUID, code string, permanent bool) error {
	terminal := permanent || d.Attempts >= 5
	delay := time.Duration(1<<uint(d.Attempts)) * time.Second
	_, err := r.pool.Exec(ctx, `UPDATE email_deliveries SET status=CASE WHEN $3 THEN 'failed' ELSE 'pending' END,
 encrypted_data=CASE WHEN $3 THEN NULL ELSE encrypted_data END, last_error_code=$4, available_at=now()+$5*interval '1 second', locked_at=NULL, lock_token=NULL
 WHERE id=$1 AND lock_token=$2 AND status='sending'`, d.ID, lease, terminal, code, int(delay.Seconds()))
	return err
}

// Ready rechecks the lease and OTP lifecycle immediately before the provider call.
func (r *Repository) Ready(ctx context.Context, id, lease uuid.UUID) (bool, error) {
	var ready bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM email_deliveries d WHERE d.id=$1 AND d.lock_token=$2 AND d.status='sending' AND d.expires_at>now() AND (d.template<>'otp' OR EXISTS (
 SELECT 1 FROM auth_challenges c JOIN auth_transactions t ON t.id=c.auth_transaction_id WHERE c.id=d.challenge_id AND c.consumed_at IS NULL AND c.attempts<c.max_attempts AND c.expires_at>now() AND t.expires_at>now() AND t.state NOT IN ('authenticated','expired'))))`, id, lease).Scan(&ready)
	return ready, err
}
