-- name: GetCarrierProviderByID :one
SELECT *
FROM carrier_providers
WHERE id = sqlc.arg(id)
  AND status = 'active'
LIMIT 1;

-- name: GetCarrierProviderBySlug :one
SELECT *
FROM carrier_providers
WHERE slug = sqlc.arg(slug)
  AND status = 'active'
LIMIT 1;

-- name: ListCarrierProviders :many
SELECT *
FROM carrier_providers
WHERE status = 'active'
ORDER BY name ASC, slug ASC;

-- name: ListBackofficeProviders :many
SELECT
    cp.id::TEXT AS id,
    cp.slug,
    cp.name,
    cp.adapter,
    cp.status,
    COUNT(DISTINCT cc.id)::BIGINT AS connection_count,
    COUNT(DISTINCT po.id) FILTER (
        WHERE po.state IN ('pending', 'provider_accepted')
    )::BIGINT AS pending_operation_count,
    COUNT(DISTINCT po.id) FILTER (
        WHERE po.state = 'failed'
    )::BIGINT AS failed_operation_count
FROM carrier_providers AS cp
LEFT JOIN carrier_connections AS cc ON cc.provider_id = cp.id
LEFT JOIN provider_operations AS po ON po.carrier_provider_id = cp.id
GROUP BY cp.id, cp.slug, cp.name, cp.adapter, cp.status, cp.created_at
ORDER BY cp.created_at DESC;

-- name: GetBackofficeProvider :one
SELECT
    cp.id::TEXT AS id,
    cp.slug,
    cp.name,
    cp.adapter,
    cp.status,
    COUNT(DISTINCT cc.id)::BIGINT AS connection_count,
    COUNT(DISTINCT pn.id)::BIGINT AS phone_number_count,
    COUNT(DISTINCT po.id) FILTER (WHERE po.state IN ('pending', 'provider_accepted'))::BIGINT AS pending_operation_count,
    COUNT(DISTINCT po.id) FILTER (WHERE po.state = 'failed')::BIGINT AS failed_operation_count,
    to_char(cp.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at,
    to_char(cp.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS updated_at
FROM carrier_providers AS cp
LEFT JOIN carrier_connections AS cc ON cc.provider_id = cp.id
LEFT JOIN phone_numbers AS pn ON pn.provider_id = cp.id
LEFT JOIN provider_operations AS po ON po.carrier_provider_id = cp.id
WHERE cp.id = sqlc.arg(id)
GROUP BY cp.id
LIMIT 1;

-- name: ListBackofficeProviderOperations :many
SELECT po.id::TEXT AS id, o.name AS organization_name, pn.number,
       po.operation_type, po.state, po.attempts,
       COALESCE(po.provider_operation_id, '—') AS provider_operation_id,
       COALESCE(po.last_error, '—') AS last_error,
       to_char(po.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS created_at,
       to_char(po.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI') AS updated_at
FROM provider_operations AS po
JOIN organizations AS o ON o.id = po.organization_id
JOIN phone_numbers AS pn ON pn.id = po.phone_number_id
WHERE po.carrier_provider_id = sqlc.arg(carrier_provider_id)
ORDER BY po.created_at DESC
LIMIT 50;
