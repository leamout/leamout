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
    canceled_at,
    provider,
    provider_customer_id,
    provider_subscription_id,
    provider_event_created_at
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
    sqlc.narg(canceled_at),
    sqlc.narg(provider),
    sqlc.narg(provider_customer_id),
    sqlc.narg(provider_subscription_id),
    sqlc.narg(provider_event_created_at)
)
ON CONFLICT (organization_id)
DO UPDATE SET
    plan_id = EXCLUDED.plan_id,
    status = EXCLUDED.status,
    current_period_start = EXCLUDED.current_period_start,
    current_period_end = EXCLUDED.current_period_end,
    trial_ends_at = EXCLUDED.trial_ends_at,
    cancel_at_period_end = EXCLUDED.cancel_at_period_end,
    canceled_at = EXCLUDED.canceled_at,
    provider = EXCLUDED.provider,
    provider_customer_id = EXCLUDED.provider_customer_id,
    provider_subscription_id = EXCLUDED.provider_subscription_id,
    provider_event_created_at = EXCLUDED.provider_event_created_at
WHERE subscriptions.provider_event_created_at IS NULL
   OR EXCLUDED.provider_event_created_at IS NULL
   OR EXCLUDED.provider_event_created_at >= subscriptions.provider_event_created_at
RETURNING *;

-- name: GetEffectiveOrganizationEntitlement :one
SELECT
    COALESCE(
        oe.enabled,
        CASE
            WHEN jsonb_typeof(p.entitlements -> (sqlc.arg(capability)::text)) = 'boolean'
                THEN (p.entitlements ->> (sqlc.arg(capability)::text))::boolean
            ELSE FALSE
        END,
        FALSE
    )::boolean AS enabled
FROM organizations o
LEFT JOIN entitlements oe
  ON oe.organization_id = o.id
 AND oe.capability = (sqlc.arg(capability)::text)
LEFT JOIN subscriptions s
  ON s.organization_id = o.id
 AND s.status IN ('trialing', 'active')
LEFT JOIN plans p
  ON p.id = s.plan_id
WHERE o.id = sqlc.arg(organization_id)
LIMIT 1;
