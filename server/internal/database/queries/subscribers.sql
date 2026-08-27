-- name: CreateSubscriber :one
INSERT INTO subscribers (
    organization_id,
    sip_domain_id,
    username,
    domain,
    ha1_md5,
    ha1_sha256,
    ha1_sha512_256,
    display_name
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(sip_domain_id),
    sqlc.arg(username),
    sd.domain,
    sqlc.arg(ha1_md5),
    sqlc.narg(ha1_sha256),
    sqlc.narg(ha1_sha512_256),
    sqlc.narg(display_name)
FROM sip_domains AS sd
JOIN organizations AS t ON t.id = sd.organization_id
WHERE sd.id = sqlc.arg(sip_domain_id)
AND sd.organization_id = sqlc.arg(organization_id)
AND sd.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
RETURNING *;

-- name: GetSubscriberByID :one
SELECT s.*
FROM subscribers AS s
JOIN organizations AS t ON t.id = s.organization_id
WHERE s.id = sqlc.arg(id)
AND s.organization_id = sqlc.arg(organization_id)
AND s.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetSubscriberBySIPIdentity :one
SELECT s.*
FROM subscribers AS s
JOIN organizations AS t ON t.id = s.organization_id
WHERE s.domain = sqlc.arg(domain)
AND s.username = sqlc.arg(username)
AND s.organization_id = sqlc.arg(organization_id)
AND s.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
LIMIT 1;

-- name: ListSubscribersByOrganizationID :many
SELECT s.*
FROM subscribers AS s
JOIN organizations AS t ON t.id = s.organization_id
WHERE s.organization_id = sqlc.arg(organization_id)
AND s.status = 'active'
AND t.status = 'active'
AND t.deleted_at IS NULL
ORDER BY s.created_at ASC;

-- name: ListSubscribersBySipDomainID :many
SELECT s.*
FROM subscribers AS s
JOIN sip_domains AS sd ON sd.id = s.sip_domain_id
JOIN organizations AS t ON t.id = sd.organization_id
WHERE s.sip_domain_id = sqlc.arg(sip_domain_id)
AND s.organization_id = sqlc.arg(organization_id)
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
AND organization_id = sqlc.arg(organization_id)
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
AND organization_id = sqlc.arg(organization_id)
AND status = 'active'
RETURNING *;

-- name: DisableSubscriber :exec
UPDATE subscribers
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND organization_id = sqlc.arg(organization_id)
AND status = 'active';

-- name: EnableSubscriber :exec
UPDATE subscribers
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
AND organization_id = sqlc.arg(organization_id)
AND status = 'disabled';
