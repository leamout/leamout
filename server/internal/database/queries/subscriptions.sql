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
    past_due_at,
    grace_ends_at,
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
    sqlc.narg(past_due_at),
    sqlc.narg(grace_ends_at),
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
    past_due_at = CASE
        WHEN subscriptions.status = 'past_due'
         AND EXCLUDED.status = 'past_due'
         AND subscriptions.provider_subscription_id IS NOT DISTINCT FROM EXCLUDED.provider_subscription_id
         AND subscriptions.past_due_at IS NOT NULL
         AND subscriptions.grace_ends_at IS NOT NULL
            THEN subscriptions.past_due_at
        ELSE EXCLUDED.past_due_at
    END,
    grace_ends_at = CASE
        WHEN subscriptions.status = 'past_due'
         AND EXCLUDED.status = 'past_due'
         AND subscriptions.provider_subscription_id IS NOT DISTINCT FROM EXCLUDED.provider_subscription_id
         AND subscriptions.past_due_at IS NOT NULL
         AND subscriptions.grace_ends_at IS NOT NULL
            THEN subscriptions.grace_ends_at
        ELSE EXCLUDED.grace_ends_at
    END,
    cancel_at_period_end = EXCLUDED.cancel_at_period_end,
    canceled_at = EXCLUDED.canceled_at,
    provider = EXCLUDED.provider,
    provider_customer_id = EXCLUDED.provider_customer_id,
    provider_subscription_id = EXCLUDED.provider_subscription_id,
    provider_event_created_at = EXCLUDED.provider_event_created_at
WHERE (
    subscriptions.provider_event_created_at IS NULL
    OR (
        EXCLUDED.provider_event_created_at IS NOT NULL
        AND EXCLUDED.provider_event_created_at >= subscriptions.provider_event_created_at
    )
)
AND (
    subscriptions.provider_subscription_id IS NULL
    OR subscriptions.provider_subscription_id = EXCLUDED.provider_subscription_id
    OR subscriptions.status = 'canceled'
)
AND (
    subscriptions.provider_customer_id IS NULL
    OR subscriptions.provider_customer_id = EXCLUDED.provider_customer_id
)
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
 AND s.current_period_start <= now()
 AND s.current_period_end > s.current_period_start
 AND (
     (
         s.status IN ('trialing', 'active')
         AND s.current_period_end > now()
         AND (s.status <> 'trialing' OR s.trial_ends_at IS NULL OR s.trial_ends_at > now())
     )
     OR (
         s.status = 'past_due'
         AND s.past_due_at <= now()
         AND s.grace_ends_at > now()
     )
 )
LEFT JOIN plans p
  ON p.id = s.plan_id
WHERE o.id = sqlc.arg(organization_id)
LIMIT 1;

-- name: ListBillingReconciliationTargets :many
SELECT
    organization_id,
    COALESCE(provider_subscription_id, '')::TEXT AS subscription_id,
    COALESCE(provider_customer_id, '')::TEXT AS customer_id
FROM subscriptions
WHERE provider = 'polar'
  AND provider_subscription_id IS NOT NULL
  AND provider_customer_id IS NOT NULL
  AND organization_id > sqlc.arg(after_organization_id)
ORDER BY organization_id
LIMIT 100;

-- name: ListBillingOrganizationReconciliationTargets :many
SELECT o.id AS organization_id, s.provider_subscription_id AS subscription_id, s.provider_customer_id AS customer_id
FROM organizations AS o
LEFT JOIN subscriptions AS s
  ON s.organization_id = o.id
 AND s.provider = 'polar'
WHERE o.id > sqlc.arg(after_organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY o.id
LIMIT 100;
