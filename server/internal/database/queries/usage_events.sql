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
)
SELECT
    o.id AS organization_id,
    sqlc.narg(subscription_id)::uuid AS subscription_id,
    m.id AS meter_id,
    sqlc.arg(quantity) AS quantity,
    sqlc.arg(source_type) AS source_type,
    sqlc.arg(source_id) AS source_id,
    sqlc.arg(idempotency_key) AS idempotency_key,
    COALESCE(sqlc.narg(dimensions), '{}'::jsonb) AS dimensions,
    sqlc.arg(occurred_at) AS occurred_at
FROM organizations AS o
JOIN meters AS m ON m.id = sqlc.arg(meter_id)
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND m.active = true
  AND (
      sqlc.narg(subscription_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM subscriptions AS s
          WHERE s.id = sqlc.narg(subscription_id)::uuid
            AND s.organization_id = o.id
      )
  )
ON CONFLICT (idempotency_key) DO NOTHING
RETURNING *;

-- name: GetUsageEventByIdempotencyKey :one
SELECT ue.*
FROM usage_events AS ue
JOIN organizations AS o ON o.id = ue.organization_id
WHERE ue.organization_id = sqlc.arg(organization_id)
  AND ue.idempotency_key = sqlc.arg(idempotency_key)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListUsageEventsByOrganization :many
SELECT ue.*
FROM usage_events AS ue
JOIN organizations AS o ON o.id = ue.organization_id
WHERE ue.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND ue.occurred_at >= sqlc.arg(period_start)
  AND ue.occurred_at < sqlc.arg(period_end)
ORDER BY ue.occurred_at ASC;

-- name: ListUsageEventsByMeter :many
SELECT ue.*
FROM usage_events AS ue
JOIN organizations AS o ON o.id = ue.organization_id
WHERE ue.organization_id = sqlc.arg(organization_id)
  AND ue.meter_id = sqlc.arg(meter_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND ue.occurred_at >= sqlc.arg(period_start)
  AND ue.occurred_at < sqlc.arg(period_end)
ORDER BY ue.occurred_at ASC;

-- name: SumUsageByMeter :one
SELECT COALESCE(SUM(ue.quantity), 0)::bigint
FROM usage_events AS ue
JOIN organizations AS o ON o.id = ue.organization_id
WHERE ue.organization_id = sqlc.arg(organization_id)
  AND ue.meter_id = sqlc.arg(meter_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND ue.occurred_at >= sqlc.arg(period_start)
  AND ue.occurred_at < sqlc.arg(period_end);
