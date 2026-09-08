-- name: EnsureProviderCDRPollCursor :one
INSERT INTO provider_cdr_poll_cursors (
    provider,
    direction,
    window_date,
    page,
    next_attempt_at
)
VALUES (
    sqlc.arg(provider),
    sqlc.arg(direction),
    sqlc.arg(window_date),
    1,
    now()
)
ON CONFLICT (provider, direction)
DO UPDATE SET provider = EXCLUDED.provider
RETURNING *;

-- name: InsertProviderCDRPage :one
INSERT INTO provider_cdr_pages (
    provider,
    direction,
    window_date,
    page,
    record_count,
    payload_sha256,
    raw
)
VALUES (
    sqlc.arg(provider),
    sqlc.arg(direction),
    sqlc.arg(window_date),
    sqlc.arg(page),
    sqlc.arg(record_count),
    sqlc.arg(payload_sha256),
    sqlc.arg(raw)
)
ON CONFLICT (provider, direction, window_date, page, payload_sha256)
DO UPDATE SET received_at = provider_cdr_pages.received_at
RETURNING *;

-- name: AdvanceProviderCDRPollCursor :one
UPDATE provider_cdr_poll_cursors
SET
    window_date = sqlc.arg(window_date),
    page = sqlc.arg(page),
    attempt_count = 0,
    next_attempt_at = sqlc.arg(next_attempt_at),
    last_error = NULL,
    last_success_at = now()
WHERE provider = sqlc.arg(provider)
  AND direction = sqlc.arg(direction)
RETURNING *;

-- name: FailProviderCDRPollCursor :one
UPDATE provider_cdr_poll_cursors
SET
    attempt_count = attempt_count + 1,
    next_attempt_at = sqlc.arg(next_attempt_at),
    last_error = sqlc.arg(last_error)
WHERE provider = sqlc.arg(provider)
  AND direction = sqlc.arg(direction)
RETURNING *;
