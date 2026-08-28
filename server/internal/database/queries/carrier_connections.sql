-- name: CreateCarrierConnection :one
INSERT INTO carrier_connections (
    organization_id,
    provider_id,
    name,
    status,
    outbound_auth_method,
    auth_username,
    auth_secret_ciphertext,
    inbound_enabled,
    inbound_auth_method,
    inbound_username,
    inbound_secret_ciphertext,
    max_cps,
    max_concurrent_calls,
    max_daily_minutes,
    codecs,
    supports_video,
    supports_fax
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(provider_id),
    sqlc.arg(name),
    COALESCE(sqlc.narg(status), 'active'),
    COALESCE(sqlc.narg(outbound_auth_method), 'none'),
    sqlc.narg(auth_username),
    sqlc.narg(auth_secret_ciphertext),
    COALESCE(sqlc.narg(inbound_enabled), false),
    COALESCE(sqlc.narg(inbound_auth_method), 'ip'),
    sqlc.narg(inbound_username),
    sqlc.narg(inbound_secret_ciphertext),
    COALESCE(sqlc.narg(max_cps), 10),
    COALESCE(sqlc.narg(max_concurrent_calls), 100),
    sqlc.narg(max_daily_minutes),
    COALESCE(sqlc.narg(codecs), ARRAY['PCMU','PCMA']::TEXT[]),
    COALESCE(sqlc.narg(supports_video), false),
    COALESCE(sqlc.narg(supports_fax), false)
FROM organizations AS o
JOIN carrier_providers AS cp ON cp.id = sqlc.arg(provider_id)
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND cp.status = 'active'
RETURNING *;

-- name: GetCarrierConnectionByID :one
SELECT *
FROM carrier_connections
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
LIMIT 1;

-- name: ListCarrierConnectionsByOrganizationID :many
SELECT *
FROM carrier_connections
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: ListActiveCarrierConnectionsByOrganizationID :many
SELECT *
FROM carrier_connections
WHERE organization_id = sqlc.arg(organization_id)
  AND status = 'active'
ORDER BY created_at DESC;

-- name: UpdateCarrierConnection :one
UPDATE carrier_connections
SET
    name = COALESCE(sqlc.narg(name), name),
    status = COALESCE(sqlc.narg(status), status),
    outbound_auth_method = COALESCE(sqlc.narg(outbound_auth_method), outbound_auth_method),
    auth_username = COALESCE(sqlc.narg(auth_username), auth_username),
    auth_secret_ciphertext = COALESCE(sqlc.narg(auth_secret_ciphertext), auth_secret_ciphertext),
    inbound_enabled = COALESCE(sqlc.narg(inbound_enabled), inbound_enabled),
    inbound_auth_method = COALESCE(sqlc.narg(inbound_auth_method), inbound_auth_method),
    inbound_username = COALESCE(sqlc.narg(inbound_username), inbound_username),
    inbound_secret_ciphertext = COALESCE(sqlc.narg(inbound_secret_ciphertext), inbound_secret_ciphertext),
    max_cps = COALESCE(sqlc.narg(max_cps), max_cps),
    max_concurrent_calls = COALESCE(sqlc.narg(max_concurrent_calls), max_concurrent_calls),
    max_daily_minutes = COALESCE(sqlc.narg(max_daily_minutes), max_daily_minutes),
    codecs = COALESCE(sqlc.narg(codecs), codecs),
    supports_video = COALESCE(sqlc.narg(supports_video), supports_video),
    supports_fax = COALESCE(sqlc.narg(supports_fax), supports_fax),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: DisableCarrierConnection :exec
UPDATE carrier_connections
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'active';

-- name: EnableCarrierConnection :exec
UPDATE carrier_connections
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'disabled';

-- name: CreateCarrierConnectionSourceIP :one
INSERT INTO carrier_connection_source_ips (
    organization_id,
    carrier_connection_id,
    cidr
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(carrier_connection_id),
    sqlc.arg(cidr)
FROM carrier_connections AS cc
WHERE cc.id = sqlc.arg(carrier_connection_id)
  AND cc.organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: ListCarrierConnectionSourceIPs :many
SELECT *
FROM carrier_connection_source_ips
WHERE carrier_connection_id = sqlc.arg(carrier_connection_id)
  AND organization_id = sqlc.arg(organization_id)
ORDER BY cidr ASC;

-- name: DeleteCarrierConnectionSourceIP :exec
DELETE FROM carrier_connection_source_ips
WHERE id = sqlc.arg(id)
  AND carrier_connection_id = sqlc.arg(carrier_connection_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: ResolveCarrierConnectionBySourceIP :one
SELECT cc.*
FROM carrier_connection_source_ips AS src
JOIN carrier_connections AS cc
  ON cc.id = src.carrier_connection_id
 AND cc.organization_id = src.organization_id
WHERE sqlc.arg(source_ip)::INET <<= src.cidr
  AND cc.status = 'active'
  AND cc.inbound_enabled = true
  AND cc.inbound_auth_method = 'ip'
ORDER BY masklen(src.cidr) DESC
LIMIT 1;
