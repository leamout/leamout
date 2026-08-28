-- name: CreateTrunk :one
INSERT INTO trunks (
    organization_id,
    carrier_connection_id,
    name,
    direction,
    status
)
SELECT
    sqlc.arg(organization_id) AS organization_id,
    sqlc.arg(carrier_connection_id) AS carrier_connection_id,
    sqlc.arg(name) AS name,
    COALESCE(sqlc.narg(direction), 'bidirectional') AS direction,
    COALESCE(sqlc.narg(status), 'active') AS status
FROM carrier_connections AS cc
WHERE cc.id = sqlc.arg(carrier_connection_id)
  AND cc.organization_id = sqlc.arg(organization_id)
  AND cc.status = 'active'
RETURNING *;

-- name: GetTrunkByID :one
SELECT *
FROM trunks
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
LIMIT 1;

-- name: ListTrunksByOrganizationID :many
SELECT *
FROM trunks
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: ListTrunksByCarrierConnectionID :many
SELECT *
FROM trunks
WHERE carrier_connection_id = sqlc.arg(carrier_connection_id)
  AND organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: UpdateTrunk :one
UPDATE trunks
SET
    name = COALESCE(sqlc.narg(name), name),
    direction = COALESCE(sqlc.narg(direction), direction),
    status = COALESCE(sqlc.narg(status), status),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: DisableTrunk :exec
UPDATE trunks
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'active';

-- name: EnableTrunk :exec
UPDATE trunks
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'disabled';

-- name: CreateTrunkEndpoint :one
INSERT INTO trunk_endpoints (
    organization_id,
    trunk_id,
    host,
    port,
    transport,
    direction,
    priority,
    weight,
    enabled
)
SELECT
    sqlc.arg(organization_id) AS organization_id,
    sqlc.arg(trunk_id) AS trunk_id,
    sqlc.arg(host) AS host,
    COALESCE(sqlc.narg(port), 5060) AS port,
    COALESCE(sqlc.narg(transport), 'udp') AS transport,
    COALESCE(sqlc.narg(direction), 'bidirectional') AS direction,
    COALESCE(sqlc.narg(priority), 10) AS priority,
    COALESCE(sqlc.narg(weight), 100) AS weight,
    COALESCE(sqlc.narg(enabled), true) AS enabled
FROM trunks AS t
WHERE t.id = sqlc.arg(trunk_id)
  AND t.organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: GetTrunkEndpointByID :one
SELECT *
FROM trunk_endpoints
WHERE id = sqlc.arg(id)
  AND trunk_id = sqlc.arg(trunk_id)
  AND organization_id = sqlc.arg(organization_id)
LIMIT 1;

-- name: ListTrunkEndpoints :many
SELECT *
FROM trunk_endpoints
WHERE trunk_id = sqlc.arg(trunk_id)
  AND organization_id = sqlc.arg(organization_id)
ORDER BY priority ASC, weight DESC, created_at ASC;

-- name: UpdateTrunkEndpoint :one
UPDATE trunk_endpoints
SET
    host = COALESCE(sqlc.narg(host), host),
    port = COALESCE(sqlc.narg(port), port),
    transport = COALESCE(sqlc.narg(transport), transport),
    direction = COALESCE(sqlc.narg(direction), direction),
    priority = COALESCE(sqlc.narg(priority), priority),
    weight = COALESCE(sqlc.narg(weight), weight),
    enabled = COALESCE(sqlc.narg(enabled), enabled),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND trunk_id = sqlc.arg(trunk_id)
  AND organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: DeleteTrunkEndpoint :exec
DELETE FROM trunk_endpoints
WHERE id = sqlc.arg(id)
  AND trunk_id = sqlc.arg(trunk_id)
  AND organization_id = sqlc.arg(organization_id);

-- name: ListActiveOutboundTrunkEndpoints :many
SELECT te.*
FROM trunk_endpoints AS te
JOIN trunks AS t
  ON t.id = te.trunk_id
 AND t.organization_id = te.organization_id
JOIN carrier_connections AS cc
  ON cc.id = t.carrier_connection_id
 AND cc.organization_id = t.organization_id
WHERE t.id = sqlc.arg(trunk_id)
  AND t.organization_id = sqlc.arg(organization_id)
  AND t.status = 'active'
  AND t.direction IN ('outbound', 'bidirectional')
  AND cc.status = 'active'
  AND te.enabled = true
  AND te.direction IN ('outbound', 'bidirectional')
ORDER BY te.priority ASC, te.weight DESC, te.created_at ASC;
