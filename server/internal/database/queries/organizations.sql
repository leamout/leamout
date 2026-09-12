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
    COUNT(DISTINCT om.user_id) FILTER (WHERE om.status = 'active')::BIGINT AS member_count,
    COUNT(DISTINCT w.id) FILTER (WHERE w.status <> 'closed')::BIGINT AS wallet_count,
    COALESCE(string_agg(DISTINCT w.currency, ', ' ORDER BY w.currency) FILTER (WHERE w.status <> 'closed'), '—') AS currencies,
    to_char(o.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at
FROM organizations AS o
LEFT JOIN organization_members AS om ON om.organization_id = o.id
LEFT JOIN wallets AS w ON w.organization_id = o.id
WHERE o.deleted_at IS NULL
GROUP BY o.id, o.name, o.status, o.created_at
ORDER BY o.created_at DESC
LIMIT 100;

-- name: GetBackofficeOrganization :one
SELECT
    o.id::TEXT AS id,
    o.name,
    o.status,
    COUNT(DISTINCT om.user_id) FILTER (WHERE om.status = 'active')::BIGINT AS member_count,
    COUNT(DISTINCT w.id) FILTER (WHERE w.status <> 'closed')::BIGINT AS wallet_count,
    COALESCE(string_agg(DISTINCT w.currency, ', ' ORDER BY w.currency) FILTER (WHERE w.status <> 'closed'), '—') AS currencies,
    'prepaid'::TEXT AS billing_model,
    to_char(o.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at,
    to_char(o.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS updated_at
FROM organizations AS o
LEFT JOIN organization_members AS om ON om.organization_id = o.id
LEFT JOIN wallets AS w ON w.organization_id = o.id
WHERE o.id = sqlc.arg(id)
  AND o.deleted_at IS NULL
GROUP BY
    o.id,
    o.name,
    o.status,
    o.created_at,
    o.updated_at
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
