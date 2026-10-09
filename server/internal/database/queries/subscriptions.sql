-- name: GetOrganizationSubscription :one
SELECT *
FROM subscriptions
WHERE organization_id = sqlc.arg(organization_id)
LIMIT 1;

-- name: UpsertOrganizationSubscription :one
INSERT INTO subscriptions (
    id,
    organization_id,
    plan_id,
    status,
    current_period_start,
    current_period_end,
    trial_ends_at,
    cancel_at_period_end,
    canceled_at
)
VALUES (
    sqlc.arg(id),
    sqlc.arg(organization_id),
    sqlc.arg(plan_id),
    sqlc.arg(status),
    sqlc.narg(current_period_start),
    sqlc.narg(current_period_end),
    sqlc.narg(trial_ends_at),
    sqlc.arg(cancel_at_period_end),
    sqlc.narg(canceled_at)
)
ON CONFLICT (organization_id)
DO UPDATE SET
    plan_id = EXCLUDED.plan_id,
    status = EXCLUDED.status,
    current_period_start = EXCLUDED.current_period_start,
    current_period_end = EXCLUDED.current_period_end,
    trial_ends_at = EXCLUDED.trial_ends_at,
    cancel_at_period_end = EXCLUDED.cancel_at_period_end,
    canceled_at = EXCLUDED.canceled_at
RETURNING *;

-- name: GetEffectiveOrganizationEntitlement :one
SELECT
    COALESCE(oe.enabled, pe.enabled, FALSE)::boolean AS enabled
FROM organizations o
LEFT JOIN entitlements oe
  ON oe.organization_id = o.id
 AND oe.capability = sqlc.arg(capability)
LEFT JOIN subscriptions s
  ON s.organization_id = o.id
 AND s.status IN ('trialing', 'active')
LEFT JOIN subscription_plan_entitlements pe
  ON pe.plan_id = s.plan_id
 AND pe.capability = sqlc.arg(capability)
WHERE o.id = sqlc.arg(organization_id)
LIMIT 1;
