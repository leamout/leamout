-- name: ListProviderRoutingTargets :many
SELECT
    cc.id AS carrier_connection_id,
    resource.provider_resource_id
FROM carrier_connections AS cc
JOIN carrier_connection_provider_resources AS resource
  ON resource.carrier_connection_id = cc.id
 AND resource.provider_id = cc.provider_id
 AND resource.resource_type = 'voice_in_trunk'
WHERE cc.provider_id = sqlc.arg(provider_id)
  AND cc.scope = 'platform'
  AND cc.organization_id IS NULL
  AND cc.status = 'active'
  AND cc.inbound_enabled = true
ORDER BY cc.created_at ASC
LIMIT 2;
