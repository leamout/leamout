-- name: AddTenantMember :one
INSERT INTO tenant_members (
    tenant_id,
    user_id,
    role
)
SELECT
    sqlc.arg(tenant_id),
    sqlc.arg(user_id),
    sqlc.arg(role)
FROM users AS u
WHERE u.id = sqlc.arg(user_id)
AND u.disabled_at IS NULL
RETURNING *;

-- name: GetTenantMember :one
SELECT tm.*
FROM tenant_members AS tm
JOIN users AS u ON u.id = tm.user_id
WHERE tm.tenant_id = sqlc.arg(tenant_id)
AND tm.user_id = sqlc.arg(user_id)
AND tm.status = 'active'
AND u.disabled_at IS NULL
LIMIT 1;

-- name: ListMembersByTenantID :many
SELECT tm.*
FROM tenant_members AS tm
JOIN users AS u ON u.id = tm.user_id
WHERE tm.tenant_id = sqlc.arg(tenant_id)
AND tm.status = 'active'
AND u.disabled_at IS NULL
ORDER BY tm.created_at ASC;

-- name: ListTenantMembershipsByUserID :many
SELECT tm.*
FROM tenant_members AS tm
JOIN tenants AS t ON t.id = tm.tenant_id
WHERE tm.user_id = sqlc.arg(user_id)
AND tm.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY tm.created_at ASC;

-- name: UpdateMemberRole :one
UPDATE tenant_members
SET
    role = sqlc.arg(role),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
AND user_id = sqlc.arg(user_id)
AND status = 'active'
RETURNING *;

-- name: DisableTenantMember :exec
UPDATE tenant_members
SET
    status = 'disabled',
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
AND user_id = sqlc.arg(user_id)
AND status = 'active';

-- name: EnableTenantMember :exec
UPDATE tenant_members
SET
    status = 'active',
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
AND user_id = sqlc.arg(user_id)
AND status = 'disabled';

-- name: RemoveTenantMember :exec
DELETE FROM tenant_members
WHERE tenant_id = sqlc.arg(tenant_id)
AND user_id = sqlc.arg(user_id);
