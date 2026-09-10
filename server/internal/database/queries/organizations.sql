-- name: CreateOrganization :one
INSERT INTO organizations (
    name
) VALUES (
    sqlc.arg(name)
)
RETURNING *;

-- name: CreateOrganizationWithOwner :one
WITH new_organization AS (
    INSERT INTO organizations (
        name
    )
    SELECT sqlc.arg(name)
    FROM users AS u
    WHERE u.id = sqlc.arg(user_id)
    AND u.disabled_at IS NULL
    RETURNING *
), owner_membership AS (
    INSERT INTO organization_members (
        organization_id,
        user_id,
        role
    )
    SELECT
        o.id,
        sqlc.arg(user_id),
        'owner'
    FROM new_organization AS o
    RETURNING organization_id
)
SELECT o.*
FROM new_organization AS o
JOIN owner_membership AS om ON om.organization_id = o.id;

-- name: GetOrganizationByID :one
SELECT *
FROM organizations
WHERE id = sqlc.arg(id)
AND status = 'active'
AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateOrganization :one
UPDATE organizations
SET
    name = COALESCE(sqlc.narg(name), name),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND status = 'active'
AND deleted_at IS NULL
RETURNING *;

-- name: DeleteOrganization :exec
UPDATE organizations
SET
    status = 'disabled',
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: ListOrganizationsByUserID :many
SELECT t.*, tm.role AS member_role
FROM organizations AS t
JOIN organization_members AS tm ON tm.organization_id = t.id
JOIN users AS u ON u.id = tm.user_id
WHERE tm.user_id = sqlc.arg(user_id)
AND u.disabled_at IS NULL
AND tm.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY t.created_at DESC;

-- name: ListBackofficeOrganizations :many
SELECT
    o.id::TEXT AS id,
    o.name,
    o.status,
    COUNT(om.user_id) FILTER (WHERE om.status = 'active')::BIGINT AS member_count,
    COALESCE(p.name, '—') AS plan_name,
    to_char(o.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at
FROM organizations AS o
LEFT JOIN organization_members AS om ON om.organization_id = o.id
LEFT JOIN subscriptions AS s
    ON s.organization_id = o.id
   AND s.status IN ('active', 'past_due')
LEFT JOIN plans AS p ON p.id = s.plan_id
WHERE o.deleted_at IS NULL
GROUP BY o.id, o.name, o.status, p.name, o.created_at
ORDER BY o.created_at DESC
LIMIT 100;

-- name: GetBackofficeOrganization :one
SELECT
    o.id::TEXT AS id,
    o.name,
    o.status,
    COUNT(om.user_id) FILTER (WHERE om.status = 'active')::BIGINT AS member_count,
    COALESCE(subscription.plan_name, '—') AS plan_name,
    COALESCE(subscription.status, 'none') AS subscription_status,
    COALESCE(subscription.billing_provider, '—') AS billing_provider,
    COALESCE(subscription.provider_subscription_id, '—') AS provider_subscription_id,
    COALESCE(to_char(subscription.renews_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'), '—') AS renews_at,
    COALESCE(to_char(subscription.ends_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI'), '—') AS ends_at,
    to_char(o.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at,
    to_char(o.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS updated_at
FROM organizations AS o
LEFT JOIN organization_members AS om ON om.organization_id = o.id
LEFT JOIN LATERAL (
    SELECT
        p.name AS plan_name,
        s.status,
        s.billing_provider,
        s.provider_subscription_id,
        s.renews_at,
        s.ends_at
    FROM subscriptions AS s
    JOIN plans AS p ON p.id = s.plan_id
    WHERE s.organization_id = o.id
    ORDER BY
        CASE WHEN s.status IN ('active', 'past_due') THEN 0 ELSE 1 END,
        s.created_at DESC
    LIMIT 1
) AS subscription ON TRUE
WHERE o.id = sqlc.arg(id)
  AND o.deleted_at IS NULL
GROUP BY
    o.id,
    o.name,
    o.status,
    o.created_at,
    o.updated_at,
    subscription.plan_name,
    subscription.status,
    subscription.billing_provider,
    subscription.provider_subscription_id,
    subscription.renews_at,
    subscription.ends_at
LIMIT 1;

-- name: ListBackofficeOrganizationMembers :many
SELECT
    u.id::TEXT AS user_id,
    COALESCE(u.name, '—') AS name,
    u.email::TEXT AS email,
    om.role,
    om.status,
    u.email_verified,
    u.is_platform_admin,
    CASE WHEN u.disabled_at IS NULL THEN 'active' ELSE 'disabled' END AS user_status,
    to_char(om.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS joined_at
FROM organization_members AS om
JOIN users AS u ON u.id = om.user_id
WHERE om.organization_id = sqlc.arg(organization_id)
ORDER BY
    CASE om.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 ELSE 2 END,
    om.created_at ASC;
