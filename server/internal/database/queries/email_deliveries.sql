-- name: CreateEmailDelivery :exec
INSERT INTO email_deliveries (
    id,
    recipient,
    template,
    encrypted_data,
    cancellation_key,
    expires_at
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(recipient),
    sqlc.arg(template),
    sqlc.arg(encrypted_data),
    sqlc.narg(cancellation_key),
    sqlc.arg(expires_at)
);

-- name: ExpireEmailDeliveries :exec
UPDATE email_deliveries
SET
    status = 'expired',
    encrypted_data = NULL,
    locked_at = NULL,
    lock_token = NULL
WHERE
    status IN ('pending', 'sending')
    AND expires_at <= now();

-- name: FailExhaustedEmailDeliveries :exec
UPDATE email_deliveries
SET
    status = 'failed',
    encrypted_data = NULL,
    lock_token = NULL,
    locked_at = NULL,
    last_error_code = 'attempts_exhausted'
WHERE
    attempts >= 5
    AND (
        status = 'pending'
        OR (
            status = 'sending'
            AND locked_at < now() - interval '60 seconds'
        )
    );

-- name: ClaimEmailDelivery :one
WITH candidate AS (
    SELECT id
    FROM email_deliveries
    WHERE
        expires_at > now()
        AND attempts < 5
        AND (
            (status = 'pending' AND available_at <= now())
            OR (
                status = 'sending'
                AND locked_at < now() - interval '60 seconds'
            )
        )
    ORDER BY available_at
    FOR UPDATE SKIP LOCKED
    LIMIT 1
)
UPDATE email_deliveries d
SET
    status = 'sending',
    attempts = d.attempts + 1,
    locked_at = now(),
    lock_token = sqlc.arg(lock_token)
FROM candidate
WHERE d.id = candidate.id
RETURNING d.*;

-- name: CompleteEmailDelivery :exec
UPDATE email_deliveries
SET
    status = 'sent',
    encrypted_data = NULL,
    provider_message_id = sqlc.arg(provider_message_id),
    sent_at = now(),
    locked_at = NULL,
    lock_token = NULL
WHERE
    id = sqlc.arg(id)
    AND lock_token = sqlc.arg(lock_token)
    AND status = 'sending';

-- name: FailEmailDelivery :exec
UPDATE email_deliveries
SET
    status = CASE
        WHEN sqlc.arg(terminal)::boolean THEN 'failed'
        ELSE 'pending'
    END,
    encrypted_data = CASE
        WHEN sqlc.arg(terminal)::boolean THEN NULL
        ELSE encrypted_data
    END,
    last_error_code = sqlc.arg(error_code),
    available_at = now() + sqlc.arg(delay_seconds)::integer * interval '1 second',
    locked_at = NULL,
    lock_token = NULL
WHERE
    id = sqlc.arg(id)
    AND lock_token = sqlc.arg(lock_token)
    AND status = 'sending';

-- name: IsEmailDeliveryReady :one
SELECT EXISTS (
    SELECT 1
    FROM email_deliveries
    WHERE
        id = sqlc.arg(id)
        AND lock_token = sqlc.arg(lock_token)
        AND status = 'sending'
        AND expires_at > now()
);

-- name: CancelEmailDeliveries :exec
UPDATE email_deliveries
SET
    status = 'cancelled',
    encrypted_data = NULL,
    locked_at = NULL,
    lock_token = NULL
WHERE
    cancellation_key = sqlc.arg(cancellation_key)
    AND status IN ('pending', 'sending');

-- name: CountEmailDeliveriesSince :one
SELECT count(*)
FROM email_deliveries
WHERE
    recipient = sqlc.arg(recipient)
    AND template = sqlc.arg(template)
    AND created_at > sqlc.arg(since)::timestamptz;
