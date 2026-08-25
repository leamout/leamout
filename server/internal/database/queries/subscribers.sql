-- name: CreateSubscriber :one
INSERT INTO subscribers (
    tenant_id,
    sip_domain_id,
    username,
    domain,
    ha1_md5,
    ha1_sha256,
    ha1_sha512_256,
    display_name
)
SELECT
    sqlc.arg(tenant_id),
    sqlc.arg(sip_domain_id),
    sqlc.arg(username),
    sd.domain,
    sqlc.arg(ha1_md5),
    sqlc.narg(ha1_sha256),
    sqlc.narg(ha1_sha512_256),
    sqlc.narg(display_name)
FROM sip_domains AS sd
JOIN tenants AS t ON t.id = sd.tenant_id
WHERE sd.id = sqlc.arg(sip_domain_id)
AND sd.tenant_id = sqlc.arg(tenant_id)
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
RETURNING *;

-- name: GetSubscriberByID :one
SELECT s.*
FROM subscribers AS s
JOIN tenants AS t ON t.id = s.tenant_id
WHERE s.id = sqlc.arg(id)
AND s.tenant_id = sqlc.arg(tenant_id)
AND s.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetSubscriberBySIPIdentity :one
SELECT s.*
FROM subscribers AS s
JOIN tenants AS t ON t.id = s.tenant_id
WHERE s.domain = sqlc.arg(domain)
AND s.username = sqlc.arg(username)
AND s.tenant_id = sqlc.arg(tenant_id)
AND s.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: ListSubscribersByTenantID :many
SELECT s.*
FROM subscribers AS s
JOIN tenants AS t ON t.id = s.tenant_id
WHERE s.tenant_id = sqlc.arg(tenant_id)
AND s.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY s.created_at ASC;

-- name: ListSubscribersBySipDomainID :many
SELECT s.*
FROM subscribers AS s
JOIN sip_domains AS sd ON sd.id = s.sip_domain_id
JOIN tenants AS t ON t.id = sd.tenant_id
WHERE s.sip_domain_id = sqlc.arg(sip_domain_id)
AND s.tenant_id = sqlc.arg(tenant_id)
AND s.status = 'active'
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY s.created_at ASC;

-- name: UpdateSubscriber :one
UPDATE subscribers
SET
    display_name = COALESCE(sqlc.narg(display_name), display_name),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'active'
RETURNING *;

-- name: SetSubscriberPassword :one
UPDATE subscribers
SET
    ha1_md5 = sqlc.arg(ha1_md5),
    ha1_sha256 = COALESCE(sqlc.narg(ha1_sha256), ha1_sha256),
    ha1_sha512_256 = COALESCE(sqlc.narg(ha1_sha512_256), ha1_sha512_256),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'active'
RETURNING *;

-- name: DisableSubscriber :exec
UPDATE subscribers
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'active';

-- name: EnableSubscriber :exec
UPDATE subscribers
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND tenant_id = sqlc.arg(tenant_id)
AND status = 'disabled';
