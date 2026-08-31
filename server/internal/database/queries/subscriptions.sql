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
)
SELECT
    o.id AS organization_id,
    pl.id AS plan_id,
    COALESCE(sqlc.narg(status), 'pending') AS status,
    COALESCE(sqlc.narg(starts_at), NOW()) AS starts_at,
    sqlc.narg(renews_at) AS renews_at,
    sqlc.narg(ends_at) AS ends_at,
    sqlc.narg(billing_provider) AS billing_provider,
    sqlc.narg(provider_subscription_id) AS provider_subscription_id
FROM organizations AS o
JOIN plans AS pl ON pl.id = sqlc.arg(plan_id)
JOIN products AS p ON p.id = pl.product_id
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND pl.active = true
  AND p.active = true
RETURNING *;

-- name: GetSubscription :one
SELECT s.*
FROM subscriptions AS s
JOIN organizations AS o ON o.id = s.organization_id
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.id = sqlc.arg(id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: GetCurrentSubscription :one
SELECT s.*
FROM subscriptions AS s
JOIN organizations AS o ON o.id = s.organization_id
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.status IN ('active', 'past_due')
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY s.starts_at DESC, s.created_at DESC
LIMIT 1;

-- name: GetSubscriptionByProviderID :one
SELECT s.*
FROM subscriptions AS s
JOIN organizations AS o ON o.id = s.organization_id
WHERE s.billing_provider = sqlc.arg(billing_provider)
  AND s.provider_subscription_id = sqlc.arg(provider_subscription_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListSubscriptionsByOrganization :many
SELECT s.*
FROM subscriptions AS s
JOIN organizations AS o ON o.id = s.organization_id
WHERE s.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY s.created_at DESC;

-- name: UpdateSubscription :one
UPDATE subscriptions AS s
SET
    plan_id = COALESCE(sqlc.narg(plan_id)::uuid, s.plan_id),
    status = COALESCE(sqlc.narg(status), s.status),
    renews_at = COALESCE(sqlc.narg(renews_at), s.renews_at),
    ends_at = COALESCE(sqlc.narg(ends_at), s.ends_at),
    updated_at = NOW()
FROM organizations AS o
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.id = sqlc.arg(id)
  AND o.id = s.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (
      sqlc.narg(plan_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM plans AS pl
          JOIN products AS p ON p.id = pl.product_id
          WHERE pl.id = sqlc.narg(plan_id)::uuid
            AND pl.active = true
            AND p.active = true
      )
  )
RETURNING s.*;

-- name: SetSubscriptionProvider :one
UPDATE subscriptions AS s
SET
    billing_provider = sqlc.arg(billing_provider),
    provider_subscription_id = sqlc.arg(provider_subscription_id),
    updated_at = NOW()
FROM organizations AS o
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.id = sqlc.arg(id)
  AND o.id = s.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING s.*;
