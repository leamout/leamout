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
