-- name: InsertBillingWebhookEvent :one
INSERT INTO billing_webhook_events (
    provider,
    environment,
    event_id,
    event_type,
    event_created_at,
    payload,
    payload_hash
)
VALUES (
    'polar',
    sqlc.arg(environment),
    sqlc.arg(event_id),
    sqlc.arg(event_type),
    sqlc.arg(event_created_at),
    sqlc.arg(payload)::jsonb,
    sqlc.arg(payload_hash)
)
ON CONFLICT (provider, environment, event_id)
DO UPDATE SET event_id = billing_webhook_events.event_id
WHERE billing_webhook_events.payload_hash = EXCLUDED.payload_hash
RETURNING *;

-- name: ClaimBillingWebhookEvent :one
UPDATE billing_webhook_events AS e
SET
    status = 'processing',
    attempts = attempts + 1,
    locked_at = NOW(),
    lease_token = sqlc.arg(lease_token)
WHERE e.id = (
    SELECT pending.id
    FROM billing_webhook_events AS pending
    WHERE pending.environment = sqlc.arg(environment)
      AND pending.attempts < sqlc.arg(max_attempts)
      AND pending.available_at <= NOW()
      AND (
          pending.status = 'pending'
          OR (
              pending.status = 'processing'
              AND pending.locked_at < NOW() - INTERVAL '60 seconds'
          )
      )
    ORDER BY pending.received_at, pending.id
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING e.*;

-- name: GetClaimedBillingWebhookEvent :one
SELECT * FROM billing_webhook_events
WHERE id = sqlc.arg(id) AND lease_token = sqlc.arg(lease_token) AND status = 'processing'
FOR UPDATE;

-- name: CompleteBillingWebhookEvent :execrows
UPDATE billing_webhook_events
SET
    status = sqlc.arg(status),
    processed_at = NOW(),
    locked_at = NULL,
    lease_token = NULL,
    last_error = NULL
WHERE id = sqlc.arg(id)
  AND lease_token = sqlc.arg(lease_token)
  AND status = 'processing';

-- name: FailBillingWebhookEvent :execrows
UPDATE billing_webhook_events
SET
    status = sqlc.arg(status),
    locked_at = NULL,
    lease_token = NULL,
    last_error = sqlc.arg(last_error),
    available_at = NOW() + sqlc.arg(retry_seconds)::integer * INTERVAL '1 second'
WHERE id = sqlc.arg(id)
  AND lease_token = sqlc.arg(lease_token)
  AND status = 'processing';

-- name: ReapExhaustedBillingWebhookEvents :exec
UPDATE billing_webhook_events
SET
    status = 'dead_letter',
    locked_at = NULL,
    lease_token = NULL,
    last_error = 'processing lease expired at retry limit'
WHERE environment = sqlc.arg(environment)
  AND status = 'processing'
  AND locked_at < NOW() - INTERVAL '60 seconds'
  AND attempts >= sqlc.arg(max_attempts);

-- name: DeleteOldBillingWebhookEvents :exec
DELETE FROM billing_webhook_events
WHERE status IN ('processed', 'ignored') AND processed_at < NOW() - INTERVAL '30 days';
