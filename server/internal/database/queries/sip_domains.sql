-- name: CreateSipDomain :one
INSERT INTO sip_domains (
    tenant_id,
    domain
)
SELECT
    sqlc.arg(tenant_id),
    sqlc.arg(domain)
FROM tenants AS t
WHERE t.id = sqlc.arg(tenant_id)
AND t.status = 'active'
AND t.deleted_at IS NULL
RETURNING *;

-- name: GetSipDomainByID :one
SELECT sd.*
FROM sip_domains AS sd
JOIN tenants AS t ON t.id = sd.tenant_id
WHERE sd.id = sqlc.arg(id)
AND sd.tenant_id = sqlc.arg(tenant_id)
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetSipDomainByDomain :one
SELECT sd.*
FROM sip_domains AS sd
JOIN tenants AS t ON t.id = sd.tenant_id
WHERE sd.domain = sqlc.arg(domain)
AND sd.tenant_id = sqlc.arg(tenant_id)
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: ListSipDomainsByTenantID :many
SELECT sd.*
FROM sip_domains AS sd
JOIN tenants AS t ON t.id = sd.tenant_id
WHERE sd.tenant_id = sqlc.arg(tenant_id)
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY sd.created_at ASC;

-- name: UpdateSipDomain :one
UPDATE sip_domains
SET
    domain = COALESCE(sqlc.narg(domain), domain),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'active'
RETURNING *;

-- name: DisableSipDomain :exec
UPDATE sip_domains
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'active';

-- name: EnableSipDomain :exec
UPDATE sip_domains
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'disabled';
