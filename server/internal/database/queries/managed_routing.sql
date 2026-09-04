-- name: CreatePlatformCarrierConnectionSourceIP :one
INSERT INTO carrier_connection_source_ips (
    organization_id,
    carrier_connection_id,
    cidr
)
SELECT
    NULL::UUID AS organization_id,
    cc.id AS carrier_connection_id,
    sqlc.arg(cidr) AS cidr
FROM carrier_connections AS cc
WHERE cc.id = sqlc.arg(carrier_connection_id)
  AND cc.scope = 'platform'
  AND cc.organization_id IS NULL
RETURNING *;

-- name: ListPlatformCarrierConnectionSourceIPs :many
SELECT src.*
FROM carrier_connection_source_ips AS src
JOIN carrier_connections AS cc ON cc.id = src.carrier_connection_id
WHERE src.carrier_connection_id = sqlc.arg(carrier_connection_id)
  AND src.organization_id IS NULL
  AND cc.scope = 'platform'
  AND cc.organization_id IS NULL
ORDER BY src.cidr ASC;

-- name: ResolveCarrierConnectionBySourceIPAnyScope :one
WITH matches AS (
    SELECT
        cc.id,
        masklen(src.cidr) AS specificity
    FROM carrier_connection_source_ips AS src
    JOIN carrier_connections AS cc
      ON cc.id = src.carrier_connection_id
     AND cc.organization_id IS NOT DISTINCT FROM src.organization_id
    WHERE sqlc.arg(source_ip)::INET <<= src.cidr
      AND cc.status = 'active'
      AND cc.inbound_enabled = true
      AND cc.inbound_auth_method = 'ip'
),
best AS (
    SELECT max(specificity) AS specificity
    FROM matches
),
resolved AS (
    SELECT matches.id
    FROM matches
    JOIN best USING (specificity)
)
SELECT cc.*
FROM resolved
JOIN carrier_connections AS cc ON cc.id = resolved.id
WHERE (SELECT count(*) FROM resolved) = 1
LIMIT 1;

-- name: ResolveInboundPhoneNumber :one
SELECT pn.*
FROM phone_numbers AS pn
JOIN carrier_connections AS cc ON cc.id = pn.carrier_connection_id
JOIN organizations AS o ON o.id = pn.organization_id
WHERE pn.number = sqlc.arg(number)
  AND pn.carrier_connection_id = sqlc.arg(carrier_connection_id)
  AND pn.status = 'active'
  AND pn.voice_enabled = true
  AND cc.status = 'active'
  AND cc.inbound_enabled = true
  AND (
      (
          cc.scope = 'organization'
          AND cc.organization_id = pn.organization_id
          AND pn.provisioning_mode = 'byoc'
      )
      OR (
          cc.scope = 'platform'
          AND cc.organization_id IS NULL
          AND pn.provisioning_mode = 'managed'
      )
  )
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;
