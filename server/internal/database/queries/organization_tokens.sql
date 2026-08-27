-- name: CreateOrganizationToken :one
INSERT INTO organization_tokens (
    organization_id,
    created_by,
    name,
    description,
    token_hash,
    token_prefix,
    scopes,
    expires_at
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(created_by),
    sqlc.arg(name),
    sqlc.narg(description),
    sqlc.arg(token_hash),
    sqlc.arg(token_prefix),
    sqlc.arg(scopes)::jsonb,
    sqlc.narg(expires_at)::timestamptz
FROM organizations AS t
JOIN users AS u ON u.id = sqlc.arg(created_by)
JOIN organization_members AS tm ON tm.organization_id = t.id AND tm.user_id = u.id
WHERE t.id = sqlc.arg(organization_id)
AND t.status = 'active'
AND t.deleted_at IS NULL
AND u.disabled_at IS NULL
AND tm.status = 'active'
AND tm.role = 'admin'
RETURNING *;

-- name: GetOrganizationTokenByTokenHash :one
SELECT ak.*
FROM organization_tokens AS ak
JOIN organizations AS t ON t.id = ak.organization_id
WHERE ak.token_hash = sqlc.arg(token_hash)
AND (ak.expires_at IS NULL OR ak.expires_at > NOW())
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetOrganizationTokenByID :one
SELECT ak.*
FROM organization_tokens AS ak
WHERE ak.id = sqlc.arg(id)
AND ak.organization_id = sqlc.arg(organization_id)
LIMIT 1;

-- name: ListOrganizationTokensByOrganizationID :many
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
FROM organization_tokens AS ak
JOIN users AS u ON u.id = ak.created_by
WHERE ak.organization_id = sqlc.arg(organization_id)
ORDER BY ak.created_at DESC;

-- name: UpdateOrganizationToken :one
UPDATE organization_tokens
SET
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    scopes = COALESCE(sqlc.narg(scopes)::jsonb, scopes),
    expires_at = COALESCE(sqlc.narg(expires_at), expires_at),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: TouchOrganizationToken :exec
UPDATE organization_tokens
SET
    last_used_at = NOW(),
    last_used_ip = sqlc.narg(last_used_ip)::inet,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND (last_used_at IS NULL OR last_used_at < NOW() - INTERVAL '1 minute');
