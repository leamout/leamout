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
) VALUES (
    sqlc.arg(plan_id),
    sqlc.arg(meter_id),
    sqlc.narg(carrier_provider_id),
    sqlc.narg(direction),
    sqlc.narg(country_code),
    sqlc.narg(network),
    sqlc.arg(currency),
    sqlc.arg(unit_amount_micros),
    COALESCE(sqlc.narg(unit_size), 1),
    COALESCE(sqlc.narg(effective_from), NOW()),
    sqlc.narg(effective_until),
    COALESCE(sqlc.narg(active), true)
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
SELECT *
FROM carrier_rates
WHERE plan_id = sqlc.arg(plan_id)
  AND meter_id = sqlc.arg(meter_id)
  AND active = true
  AND effective_from <= sqlc.arg(at_time)
  AND (effective_until IS NULL OR effective_until > sqlc.arg(at_time))
  AND (carrier_provider_id IS NULL OR carrier_provider_id = sqlc.narg(carrier_provider_id))
  AND (direction IS NULL OR direction = sqlc.narg(direction))
  AND (country_code IS NULL OR country_code = sqlc.narg(country_code))
  AND (network IS NULL OR network = sqlc.narg(network))
ORDER BY
    (carrier_provider_id IS NOT NULL)::int
    + (direction IS NOT NULL)::int
    + (country_code IS NOT NULL)::int
    + (network IS NOT NULL)::int DESC,
    effective_from DESC
LIMIT 1;

-- name: SetCarrierRateActive :one
UPDATE carrier_rates
SET
    active = sqlc.arg(active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;
