-- name: ListActivePlans :many
SELECT *
FROM plans
WHERE status = 'active'
ORDER BY amount_minor ASC NULLS LAST, created_at ASC;

-- name: GetPlanByID :one
SELECT *
FROM plans
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetPlanByCode :one
SELECT *
FROM plans
WHERE code = sqlc.arg(code)
LIMIT 1;

-- name: UpsertPlan :one
INSERT INTO plans (
    id,
    code,
    name,
    description,
    pricing_type,
    currency,
    amount_minor,
    billing_interval,
    entitlements,
    limits,
    status
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.narg(description),
    sqlc.arg(pricing_type),
    sqlc.arg(currency),
    sqlc.narg(amount_minor),
    sqlc.arg(billing_interval),
    sqlc.arg(entitlements),
    sqlc.arg(limits),
    sqlc.arg(status)
)
ON CONFLICT (code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    pricing_type = EXCLUDED.pricing_type,
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    billing_interval = EXCLUDED.billing_interval,
    entitlements = EXCLUDED.entitlements,
    limits = EXCLUDED.limits,
    status = EXCLUDED.status
RETURNING *;
