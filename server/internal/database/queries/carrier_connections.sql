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
    sqlc.arg(organization_id) AS organization_id,
    sqlc.arg(provider_id) AS provider_id,
    sqlc.arg(name) AS name,
    COALESCE(sqlc.narg(status), 'active') AS status,
    COALESCE(sqlc.narg(outbound_auth_method), 'none') AS outbound_auth_method,
    sqlc.narg(auth_username) AS auth_username,
    sqlc.narg(auth_secret_ciphertext) AS auth_secret_ciphertext,
    COALESCE(sqlc.narg(inbound_enabled), false) AS inbound_enabled,
    COALESCE(sqlc.narg(inbound_auth_method), 'ip') AS inbound_auth_method,
    sqlc.narg(inbound_username) AS inbound_username,
    sqlc.narg(inbound_secret_ciphertext) AS inbound_secret_ciphertext,
    COALESCE(sqlc.narg(max_cps), 10) AS max_cps,
    COALESCE(sqlc.narg(max_concurrent_calls), 100) AS max_concurrent_calls,
    sqlc.narg(max_daily_minutes) AS max_daily_minutes,
    COALESCE(sqlc.narg(codecs), ARRAY['PCMU','PCMA']::TEXT[]) AS codecs,
    COALESCE(sqlc.narg(supports_video), false) AS supports_video,
    COALESCE(sqlc.narg(supports_fax), false) AS supports_fax
FROM organizations AS o
JOIN carrier_providers AS cp ON cp.id = sqlc.arg(provider_id)
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND cp.status = 'active'
RETURNING *;

-- name: GetCarrierConnectionByID :one
SELECT
    id,
    organization_id,
    provider_id,
    name,
    status,
    outbound_auth_method,
    auth_username,
    auth_secret_ciphertext IS NOT NULL AS has_outbound_credentials,
    inbound_enabled,
    inbound_auth_method,
    inbound_username,
    inbound_secret_ciphertext IS NOT NULL AS has_inbound_credentials,
    max_cps,
    max_concurrent_calls,
    max_daily_minutes,
    codecs,
    supports_video,
    supports_fax,
    created_at,
    updated_at
FROM carrier_connections
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
LIMIT 1;

-- name: ListCarrierConnectionsByOrganizationID :many
SELECT
    id,
    organization_id,
    provider_id,
    name,
    status,
    outbound_auth_method,
    auth_username,
    auth_secret_ciphertext IS NOT NULL AS has_outbound_credentials,
    inbound_enabled,
    inbound_auth_method,
    inbound_username,
    inbound_secret_ciphertext IS NOT NULL AS has_inbound_credentials,
    max_cps,
    max_concurrent_calls,
    max_daily_minutes,
    codecs,
    supports_video,
    supports_fax,
    created_at,
    updated_at
FROM carrier_connections
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: ListActiveCarrierConnectionsByOrganizationID :many
SELECT
    id,
    organization_id,
    provider_id,
    name,
    status,
    outbound_auth_method,
    auth_username,
    auth_secret_ciphertext IS NOT NULL AS has_outbound_credentials,
    inbound_enabled,
    inbound_auth_method,
    inbound_username,
    inbound_secret_ciphertext IS NOT NULL AS has_inbound_credentials,
    max_cps,
    max_concurrent_calls,
    max_daily_minutes,
    codecs,
    supports_video,
    supports_fax,
    created_at,
    updated_at
FROM carrier_connections
WHERE organization_id = sqlc.arg(organization_id)
  AND status = 'active'
ORDER BY created_at DESC;

-- name: GetCarrierConnectionCredentials :one
SELECT
    outbound_auth_method,
    auth_username,
    auth_secret_ciphertext,
    inbound_auth_method,
    inbound_username,
    inbound_secret_ciphertext
FROM carrier_connections
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'active'
LIMIT 1;

-- name: UpdateCarrierConnection :one
UPDATE carrier_connections
SET
    name = COALESCE(sqlc.narg(name), name),
    status = COALESCE(sqlc.narg(status), status),
    inbound_enabled = COALESCE(sqlc.narg(inbound_enabled), inbound_enabled),
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

-- name: SetCarrierConnectionOutboundDigestAuth :exec
UPDATE carrier_connections
SET
    outbound_auth_method = 'digest',
    auth_username = sqlc.arg(auth_username),
    auth_secret_ciphertext = sqlc.arg(auth_secret_ciphertext),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

-- name: ClearCarrierConnectionOutboundAuth :exec
UPDATE carrier_connections
SET
    outbound_auth_method = 'none',
    auth_username = NULL,
    auth_secret_ciphertext = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

-- name: SetCarrierConnectionInboundDigestAuth :exec
UPDATE carrier_connections
SET
    inbound_auth_method = 'digest',
    inbound_username = sqlc.arg(inbound_username),
    inbound_secret_ciphertext = sqlc.arg(inbound_secret_ciphertext),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

-- name: SetCarrierConnectionInboundIPAuth :exec
UPDATE carrier_connections
SET
    inbound_auth_method = 'ip',
    inbound_username = NULL,
    inbound_secret_ciphertext = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

-- name: SetCarrierConnectionInboundNoAuth :exec
UPDATE carrier_connections
SET
    inbound_auth_method = 'none',
    inbound_username = NULL,
    inbound_secret_ciphertext = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id);

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
    sqlc.arg(organization_id) AS organization_id,
    sqlc.arg(carrier_connection_id) AS carrier_connection_id,
    sqlc.arg(cidr) AS cidr
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
