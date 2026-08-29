-- name: CreateUsageEvent :one
INSERT INTO usage_events (
    organization_id,
    subscription_id,
    meter_id,
    quantity,
    source_type,
    source_id,
    idempotency_key,
    dimensions,
    occurred_at
) VALUES (
    sqlc.arg(organization_id),
    sqlc.narg(subscription_id),
    sqlc.arg(meter_id),
    sqlc.arg(quantity),
    sqlc.arg(source_type),
    sqlc.arg(source_id),
    sqlc.arg(idempotency_key),
    COALESCE(sqlc.narg(dimensions), '{}'::jsonb),
    sqlc.arg(occurred_at)
)
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING *;

-- name: GetUsageEventByIdempotencyKey :one
SELECT *
FROM usage_events
WHERE idempotency_key = sqlc.arg(idempotency_key)
LIMIT 1;

-- name: ListUsageEventsByOrganization :many
SELECT *
FROM usage_events
WHERE organization_id = sqlc.arg(organization_id)
  AND occurred_at >= sqlc.arg(period_start)
  AND occurred_at < sqlc.arg(period_end)
ORDER BY occurred_at ASC;

-- name: ListUsageEventsByMeter :many
SELECT *
FROM usage_events
WHERE organization_id = sqlc.arg(organization_id)
  AND meter_id = sqlc.arg(meter_id)
  AND occurred_at >= sqlc.arg(period_start)
  AND occurred_at < sqlc.arg(period_end)
ORDER BY occurred_at ASC;

-- name: SumUsageByMeter :one
SELECT COALESCE(SUM(quantity), 0)::bigint
FROM usage_events
WHERE organization_id = sqlc.arg(organization_id)
  AND meter_id = sqlc.arg(meter_id)
  AND occurred_at >= sqlc.arg(period_start)
  AND occurred_at < sqlc.arg(period_end);
