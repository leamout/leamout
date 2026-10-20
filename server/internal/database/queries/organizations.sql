-- name: CreateOrganization :one
INSERT INTO organizations (name)
VALUES (sqlc.arg(name))
RETURNING *;

-- name: CreateOrganizationWithOwner :one
WITH new_organization AS (
    INSERT INTO organizations (name)
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
JOIN owner_membership AS om
  ON om.organization_id = o.id;

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

-- name: LockOrganizationForCapacity :one
SELECT id FROM organizations
WHERE id = sqlc.arg(organization_id)
  AND status = 'active'
  AND deleted_at IS NULL
FOR UPDATE;

-- name: ListBackofficeOrganizations :many
SELECT
    o.id,
    o.name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::TEXT AS status,
    o.created_at
FROM organizations AS o
WHERE (
    sqlc.arg(search)::TEXT = ''
    OR strpos(lower(o.name), lower(sqlc.arg(search)::TEXT)) > 0
    OR o.id::TEXT = sqlc.arg(search)::TEXT
)
AND (
    sqlc.arg(status)::TEXT = ''
    OR (CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END) = sqlc.arg(status)::TEXT
)
ORDER BY o.created_at DESC, o.id DESC
LIMIT sqlc.arg(page_limit)::INT
OFFSET sqlc.arg(page_offset)::INT;

-- name: GetBackofficeOrganization :one
SELECT
    o.id,
    o.name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::TEXT AS status,
    o.created_at,
    o.updated_at,
    o.deleted_at,
    (SELECT count(*) FROM organization_members AS m WHERE m.organization_id = o.id AND m.status = 'active')::BIGINT AS active_members,
    (SELECT count(*) FROM voice_agents AS a WHERE a.organization_id = o.id)::BIGINT AS agents,
    (SELECT count(*) FROM trunks AS t WHERE t.organization_id = o.id)::BIGINT AS trunks,
    COALESCE(p.name, 'No subscription')::TEXT AS plan_name,
    COALESCE(s.status, 'none')::TEXT AS subscription_status,
    s.current_period_end
FROM organizations AS o
LEFT JOIN subscriptions AS s ON s.organization_id = o.id
LEFT JOIN plans AS p ON p.id = s.plan_id
WHERE o.id = sqlc.arg(id);
