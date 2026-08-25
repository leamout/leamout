-- name: CreateAPIKey :one
INSERT INTO api_keys (
    tenant_id,
    created_by,
    name,
    description,
    token_hash,
    token_prefix,
    scopes,
    expires_at
)
SELECT
    sqlc.arg(tenant_id),
    sqlc.arg(created_by),
    sqlc.arg(name),
    sqlc.narg(description),
    sqlc.arg(token_hash),
    sqlc.arg(token_prefix),
    sqlc.arg(scopes)::jsonb,
    sqlc.narg(expires_at)::timestamptz
FROM tenants AS t
JOIN users AS u ON u.id = sqlc.arg(created_by)
JOIN tenant_members AS tm ON tm.tenant_id = t.id AND tm.user_id = u.id
WHERE t.id = sqlc.arg(tenant_id)
AND t.status = 'active'
AND t.deleted_at IS NULL
AND u.disabled_at IS NULL
AND tm.status = 'active'
AND tm.role = 'admin'
RETURNING *;

-- name: GetAPIKeyByTokenHash :one
SELECT ak.*
FROM api_keys AS ak
JOIN tenants AS t ON t.id = ak.tenant_id
WHERE ak.token_hash = sqlc.arg(token_hash)
AND (ak.expires_at IS NULL OR ak.expires_at > NOW())
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetAPIKeyByID :one
SELECT ak.*
FROM api_keys AS ak
WHERE ak.id = sqlc.arg(id)
AND ak.tenant_id = sqlc.arg(tenant_id)
LIMIT 1;

-- name: ListAPIKeysByTenantID :many
SELECT
    ak.id,
    ak.name,
    ak.description,
    ak.token_prefix,
    ak.scopes,
    ak.last_used_at,
    ak.last_used_ip,
    ak.expires_at,
    ak.created_at,
    u.name AS created_by
FROM api_keys AS ak
JOIN users AS u ON u.id = ak.created_by
WHERE ak.tenant_id = sqlc.arg(tenant_id)
ORDER BY ak.created_at DESC;

-- name: UpdateAPIKey :one
UPDATE api_keys
SET
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    scopes = COALESCE(sqlc.narg(scopes)::jsonb, scopes),
    expires_at = COALESCE(sqlc.narg(expires_at), expires_at),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
RETURNING *;

-- name: TouchAPIKey :exec
UPDATE api_keys
SET
    last_used_at = NOW(),
    last_used_ip = sqlc.narg(last_used_ip)::inet,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND (last_used_at IS NULL OR last_used_at < NOW() - INTERVAL '1 minute');
