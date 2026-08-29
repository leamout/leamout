-- name: CreateSubscription :one
INSERT INTO subscriptions (
    organization_id,
    plan_id,
    status,
    starts_at,
    renews_at,
    ends_at,
    billing_provider,
    provider_subscription_id
) VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(plan_id),
    COALESCE(sqlc.narg(status), 'pending'),
    COALESCE(sqlc.narg(starts_at), NOW()),
    sqlc.narg(renews_at),
    sqlc.narg(ends_at),
    sqlc.narg(billing_provider),
    sqlc.narg(provider_subscription_id)
)
RETURNING *;

-- name: GetSubscription :one
SELECT *
FROM subscriptions
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetCurrentSubscription :one
SELECT *
FROM subscriptions
WHERE organization_id = sqlc.arg(organization_id)
  AND status IN ('active', 'past_due')
ORDER BY starts_at DESC, created_at DESC
LIMIT 1;

-- name: GetSubscriptionByProviderID :one
SELECT *
FROM subscriptions
WHERE billing_provider = sqlc.arg(billing_provider)
  AND provider_subscription_id = sqlc.arg(provider_subscription_id)
LIMIT 1;

-- name: ListSubscriptionsByOrganization :many
SELECT *
FROM subscriptions
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: UpdateSubscription :one
UPDATE subscriptions
SET
    plan_id = COALESCE(sqlc.narg(plan_id), plan_id),
    status = COALESCE(sqlc.narg(status), status),
    renews_at = COALESCE(sqlc.narg(renews_at), renews_at),
    ends_at = COALESCE(sqlc.narg(ends_at), ends_at),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;

-- name: SetSubscriptionProvider :one
UPDATE subscriptions
SET
    billing_provider = sqlc.arg(billing_provider),
    provider_subscription_id = sqlc.arg(provider_subscription_id),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;
