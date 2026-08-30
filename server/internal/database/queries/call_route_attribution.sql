-- name: SetCallRouteAttribution :one
UPDATE calls
SET
    carrier_connection_id = sqlc.arg(carrier_connection_id),
    trunk_id = sqlc.arg(trunk_id),
    trunk_endpoint_id = sqlc.arg(trunk_endpoint_id),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;
