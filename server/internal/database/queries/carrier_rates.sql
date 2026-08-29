-- name: CreateCarrierRate :one
INSERT INTO carrier_rates (
    plan_id,
    meter_id,
    carrier_provider_id,
    direction,
    country_code,
    network,
    currency,
    unit_amount_micros,
    unit_size,
    effective_from,
    effective_until,
    active
)
SELECT
    pl.id AS plan_id,
    m.id AS meter_id,
    sqlc.narg(carrier_provider_id)::uuid AS carrier_provider_id,
    sqlc.narg(direction) AS direction,
    sqlc.narg(country_code) AS country_code,
    sqlc.narg(network) AS network,
    sqlc.arg(currency) AS currency,
    sqlc.arg(unit_amount_micros) AS unit_amount_micros,
    COALESCE(sqlc.narg(unit_size), 1) AS unit_size,
    COALESCE(sqlc.narg(effective_from), NOW()) AS effective_from,
    sqlc.narg(effective_until) AS effective_until,
    COALESCE(sqlc.narg(active), true) AS active
FROM plans AS pl
JOIN products AS p ON p.id = pl.product_id
JOIN meters AS m ON m.id = sqlc.arg(meter_id)
WHERE pl.id = sqlc.arg(plan_id)
  AND pl.active = true
  AND p.active = true
  AND m.active = true
  AND (
      sqlc.narg(carrier_provider_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM carrier_providers AS cp
          WHERE cp.id = sqlc.narg(carrier_provider_id)::uuid
            AND cp.status = 'active'
      )
  )
RETURNING *;

-- name: GetCarrierRateByID :one
SELECT *
FROM carrier_rates
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: ListCarrierRatesByPlan :many
SELECT *
FROM carrier_rates
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY effective_from DESC, created_at DESC;

-- name: ResolveCarrierRate :one
SELECT cr.*
FROM subscriptions AS s
JOIN organizations AS o ON o.id = s.organization_id
JOIN plans AS pl ON pl.id = s.plan_id
JOIN products AS p ON p.id = pl.product_id
JOIN meters AS m ON m.id = sqlc.arg(meter_id)
JOIN carrier_rates AS cr
  ON cr.plan_id = s.plan_id
 AND cr.meter_id = m.id
WHERE s.id = sqlc.arg(subscription_id)
  AND s.organization_id = sqlc.arg(organization_id)
  AND s.status IN ('active', 'past_due')
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND pl.active = true
  AND p.active = true
  AND m.active = true
  AND cr.active = true
  AND cr.effective_from <= sqlc.arg(at_time)
  AND (cr.effective_until IS NULL OR cr.effective_until > sqlc.arg(at_time))
  AND (cr.carrier_provider_id IS NULL OR cr.carrier_provider_id = sqlc.narg(carrier_provider_id)::uuid)
  AND (cr.direction IS NULL OR cr.direction = sqlc.narg(direction))
  AND (cr.country_code IS NULL OR cr.country_code = sqlc.narg(country_code))
  AND (cr.network IS NULL OR cr.network = sqlc.narg(network))
ORDER BY
    (cr.carrier_provider_id IS NOT NULL)::int
    + (cr.direction IS NOT NULL)::int
    + (cr.country_code IS NOT NULL)::int
    + (cr.network IS NOT NULL)::int DESC,
    cr.effective_from DESC
LIMIT 1;

-- name: SetCarrierRateActive :one
UPDATE carrier_rates AS cr
SET
    active = sqlc.arg(active),
    updated_at = NOW()
FROM plans AS pl, products AS p, meters AS m
WHERE cr.id = sqlc.arg(id)
  AND pl.id = cr.plan_id
  AND p.id = pl.product_id
  AND m.id = cr.meter_id
  AND (
      sqlc.arg(active) = false
      OR (
          pl.active = true
          AND p.active = true
          AND m.active = true
          AND (
              cr.carrier_provider_id IS NULL
              OR EXISTS (
                  SELECT 1
                  FROM carrier_providers AS cp
                  WHERE cp.id = cr.carrier_provider_id
                    AND cp.status = 'active'
              )
          )
      )
  )
RETURNING *;
