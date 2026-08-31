-- name: CreateSubscription :one
INSERT INTO subscriptions (
    organization_id,
    plan_id,
    price_id,
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
    pr.id AS price_id,
    COALESCE(sqlc.narg(status), 'pending') AS status,
    COALESCE(sqlc.narg(starts_at), NOW()) AS starts_at,
    sqlc.narg(renews_at) AS renews_at,
    sqlc.narg(ends_at) AS ends_at,
    sqlc.narg(billing_provider) AS billing_provider,
    sqlc.narg(provider_subscription_id) AS provider_subscription_id
FROM organizations AS o
JOIN plans AS pl ON pl.id = sqlc.arg(plan_id)
JOIN products AS p ON p.id = pl.product_id
JOIN prices AS pr
  ON pr.id = sqlc.arg(price_id)
 AND pr.plan_id = pl.id
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND pl.active = true
  AND p.active = true
  AND pr.active = true
  AND pr.effective_from <= COALESCE(sqlc.narg(starts_at), NOW())
  AND (pr.effective_until IS NULL OR pr.effective_until > COALESCE(sqlc.narg(starts_at), NOW()))
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

-- name: UpdateSubscriptionPeriod :one
UPDATE subscriptions AS s
SET
    renews_at = COALESCE(sqlc.narg(renews_at), s.renews_at),
    ends_at = COALESCE(sqlc.narg(ends_at), s.ends_at),
    updated_at = NOW()
FROM organizations AS o
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.id = sqlc.arg(id)
  AND o.id = s.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING s.*;

-- name: ChangeSubscriptionPrice :one
UPDATE subscriptions AS s
SET
    plan_id = pr.plan_id,
    price_id = pr.id,
    updated_at = NOW()
FROM organizations AS o,
     prices AS pr
JOIN plans AS pl ON pl.id = pr.plan_id
JOIN products AS p ON p.id = pl.product_id
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.id = sqlc.arg(id)
  AND o.id = s.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND pr.id = sqlc.arg(price_id)
  AND pr.plan_id = sqlc.arg(plan_id)
  AND pr.active = true
  AND pr.effective_from <= NOW()
  AND (pr.effective_until IS NULL OR pr.effective_until > NOW())
  AND pl.active = true
  AND p.active = true
RETURNING s.*;

-- name: CompareAndSetSubscriptionStatus :one
UPDATE subscriptions AS s
SET
    status = sqlc.arg(status),
    updated_at = NOW()
FROM organizations AS o
WHERE s.organization_id = sqlc.arg(organization_id)
  AND s.id = sqlc.arg(id)
  AND s.status = sqlc.arg(expected_status)
  AND o.id = s.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
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
