-- name: CreateTenant :one
INSERT INTO tenants (
    slug,
    name
) VALUES (
    sqlc.arg(slug),
    sqlc.arg(name)
)
RETURNING *;

-- name: GetTenantByID :one
SELECT *
FROM tenants
WHERE id = sqlc.arg(id)
AND status = 'active'
AND deleted_at IS NULL
LIMIT 1;

-- name: GetTenantBySlug :one
SELECT *
FROM tenants
WHERE slug = sqlc.arg(slug)
AND status = 'active'
AND deleted_at IS NULL
LIMIT 1;

-- name: UpdateTenant :one
UPDATE tenants
SET
    name = COALESCE(sqlc.narg(name), name),
    slug = COALESCE(sqlc.narg(slug), slug),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND status = 'active'
AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTenant :exec
UPDATE tenants
SET
    status = 'disabled',
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND deleted_at IS NULL;

-- name: ListTenantsByUserID :many
SELECT t.*
FROM tenants AS t
JOIN tenant_members AS tm ON tm.tenant_id = t.id
JOIN users AS u ON u.id = tm.user_id
WHERE tm.user_id = sqlc.arg(user_id)
AND u.disabled_at IS NULL
AND tm.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY t.created_at DESC;
