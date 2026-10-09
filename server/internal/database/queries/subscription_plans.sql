-- name: ListActiveSubscriptionPlans :many
SELECT *
FROM subscription_plans
WHERE status = 'active'
ORDER BY amount_minor ASC, created_at ASC;

-- name: GetSubscriptionPlanByID :one
SELECT *
FROM subscription_plans
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetSubscriptionPlanByCode :one
SELECT *
FROM subscription_plans
WHERE code = sqlc.arg(code)
LIMIT 1;

-- name: ListSubscriptionPlanEntitlements :many
SELECT capability, enabled
FROM subscription_plan_entitlements
WHERE plan_id = sqlc.arg(plan_id)
ORDER BY capability;

-- name: UpsertSubscriptionPlan :one
INSERT INTO subscription_plans (
    id,
    code,
    name,
    description,
    currency,
    amount_minor,
    billing_interval,
    status
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.narg(description),
    sqlc.arg(currency),
    sqlc.arg(amount_minor),
    sqlc.arg(billing_interval),
    sqlc.arg(status)
)
ON CONFLICT (code)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    currency = EXCLUDED.currency,
    amount_minor = EXCLUDED.amount_minor,
    billing_interval = EXCLUDED.billing_interval,
    status = EXCLUDED.status
RETURNING *;

-- name: UpsertSubscriptionPlanEntitlement :exec
INSERT INTO subscription_plan_entitlements (
    plan_id,
    capability,
    enabled
)
VALUES (
    sqlc.arg(plan_id),
    sqlc.arg(capability),
    sqlc.arg(enabled)
)
ON CONFLICT (plan_id, capability)
DO UPDATE SET enabled = EXCLUDED.enabled;
