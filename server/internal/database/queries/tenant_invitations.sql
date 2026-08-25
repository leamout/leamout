-- name: CreateInvitation :one
INSERT INTO tenant_invitations (
    tenant_id,
    invited_by,
    email,
    role,
    token_hash,
    expires_at
)
SELECT
    sqlc.arg(tenant_id),
    sqlc.arg(invited_by),
    sqlc.arg(email),
    sqlc.arg(role),
    sqlc.arg(token_hash),
    sqlc.arg(expires_at)
FROM tenants AS t
JOIN users AS u ON u.id = sqlc.arg(invited_by)
JOIN tenant_members AS tm ON tm.tenant_id = t.id AND tm.user_id = u.id
WHERE t.id = sqlc.arg(tenant_id)
AND t.status = 'active'
AND t.deleted_at IS NULL
AND u.disabled_at IS NULL
AND tm.status = 'active'
AND tm.role = 'admin'  -- Only admins can invite
RETURNING *;

-- name: GetInvitationByTokenHash :one
SELECT i.*
FROM tenant_invitations AS i
JOIN tenants AS t ON t.id = i.tenant_id
WHERE i.token_hash = sqlc.arg(token_hash)
AND i.status = 'pending'
AND i.expires_at > NOW()
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetPendingInvitationByEmailAndTenant :one
SELECT *
FROM tenant_invitations
WHERE tenant_id = sqlc.arg(tenant_id)
AND email = sqlc.arg(email)
AND status = 'pending'
AND expires_at > NOW()
LIMIT 1;

-- name: ListPendingInvitationsByTenantID :many
SELECT i.*
FROM tenant_invitations AS i
WHERE i.tenant_id = sqlc.arg(tenant_id)
AND i.status = 'pending'
AND i.expires_at > NOW()
ORDER BY i.created_at DESC;

-- name: ListInvitationsForEmail :many
SELECT i.*, t.slug AS tenant_slug, t.name AS tenant_name
FROM tenant_invitations AS i
JOIN tenants AS t ON t.id = i.tenant_id
WHERE i.email = sqlc.arg(email)
AND i.status = 'pending'
AND i.expires_at > NOW()
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY i.created_at DESC;

-- name: AcceptInvitation :one
UPDATE tenant_invitations
SET
    status = 'accepted',
    accepted_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND status = 'pending'
AND expires_at > NOW()
RETURNING *;

-- name: DeclineInvitation :one
UPDATE tenant_invitations
SET
    status = 'declined',
    declined_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND status = 'pending'
RETURNING *;

-- name: RevokeInvitation :one
UPDATE tenant_invitations
SET
    status = 'revoked',
    revoked_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'pending'
RETURNING *;

-- name: ExpireStaleInvitations :exec
UPDATE tenant_invitations
SET
    status = 'expired',
    updated_at = NOW()
WHERE status = 'pending'
AND expires_at <= NOW();
