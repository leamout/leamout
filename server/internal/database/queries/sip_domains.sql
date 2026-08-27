-- name: CreateSipDomain :one
INSERT INTO sip_domains (
    organization_id,
    domain
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(domain)
FROM organizations AS t
WHERE t.id = sqlc.arg(organization_id)
AND t.status = 'active'
AND t.deleted_at IS NULL
RETURNING *;

-- name: GetSipDomainByID :one
SELECT sd.*
FROM sip_domains AS sd
JOIN organizations AS t ON t.id = sd.organization_id
WHERE sd.id = sqlc.arg(id)
AND sd.organization_id = sqlc.arg(organization_id)
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetSipDomainByDomain :one
SELECT sd.*
FROM sip_domains AS sd
JOIN organizations AS t ON t.id = sd.organization_id
WHERE sd.domain = sqlc.arg(domain)
AND sd.organization_id = sqlc.arg(organization_id)
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: ListSipDomainsByOrganizationID :many
SELECT sd.*
FROM sip_domains AS sd
JOIN organizations AS t ON t.id = sd.organization_id
WHERE sd.organization_id = sqlc.arg(organization_id)
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
AND organization_id = sqlc.arg(organization_id)
AND status = 'active'
RETURNING *;

-- name: DisableSipDomain :exec
UPDATE sip_domains
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND organization_id = sqlc.arg(organization_id)
AND status = 'active';

-- name: EnableSipDomain :exec
UPDATE sip_domains
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND organization_id = sqlc.arg(organization_id)
AND status = 'disabled';
