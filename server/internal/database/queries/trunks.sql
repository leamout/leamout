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

-- name: DisableTrunk :one
UPDATE trunks
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'active'
RETURNING *;

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
    health_status = CASE
        WHEN sqlc.narg(host)::TEXT IS NOT NULL
          OR sqlc.narg(port)::INTEGER IS NOT NULL
          OR sqlc.narg(transport)::TEXT IS NOT NULL THEN 'unknown'
        ELSE health_status
    END,
    consecutive_failures = CASE
        WHEN sqlc.narg(host)::TEXT IS NOT NULL
          OR sqlc.narg(port)::INTEGER IS NOT NULL
          OR sqlc.narg(transport)::TEXT IS NOT NULL THEN 0
        ELSE consecutive_failures
    END,
    last_checked_at = CASE WHEN sqlc.narg(host)::TEXT IS NOT NULL OR sqlc.narg(port)::INTEGER IS NOT NULL OR sqlc.narg(transport)::TEXT IS NOT NULL THEN NULL ELSE last_checked_at END,
    last_response_code = CASE WHEN sqlc.narg(host)::TEXT IS NOT NULL OR sqlc.narg(port)::INTEGER IS NOT NULL OR sqlc.narg(transport)::TEXT IS NOT NULL THEN NULL ELSE last_response_code END,
    last_latency_ms = CASE WHEN sqlc.narg(host)::TEXT IS NOT NULL OR sqlc.narg(port)::INTEGER IS NOT NULL OR sqlc.narg(transport)::TEXT IS NOT NULL THEN NULL ELSE last_latency_ms END,
    last_error = CASE WHEN sqlc.narg(host)::TEXT IS NOT NULL OR sqlc.narg(port)::INTEGER IS NOT NULL OR sqlc.narg(transport)::TEXT IS NOT NULL THEN NULL ELSE last_error END,
    cooldown_until = CASE WHEN sqlc.narg(host)::TEXT IS NOT NULL OR sqlc.narg(port)::INTEGER IS NOT NULL OR sqlc.narg(transport)::TEXT IS NOT NULL THEN NULL ELSE cooldown_until END,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND trunk_id = sqlc.arg(trunk_id)
  AND organization_id = sqlc.arg(organization_id)
RETURNING *;

-- name: DeleteTrunkEndpoint :one
DELETE FROM trunk_endpoints
WHERE id = sqlc.arg(id)
  AND trunk_id = sqlc.arg(trunk_id)
  AND organization_id = sqlc.arg(organization_id)
RETURNING *;

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

-- name: ListTrunkEndpointsForHealthCheck :many
WITH due AS (
    SELECT te.id
    FROM trunk_endpoints AS te
    JOIN trunks AS t
      ON t.id = te.trunk_id
     AND t.organization_id = te.organization_id
    JOIN carrier_connections AS cc
      ON cc.id = t.carrier_connection_id
     AND cc.organization_id = t.organization_id
    WHERE te.enabled = true
      AND t.status = 'active'
      AND cc.status = 'active'
      AND (te.cooldown_until IS NULL OR te.cooldown_until <= sqlc.arg(checked_at))
      AND (te.last_checked_at IS NULL OR te.last_checked_at <= sqlc.arg(due_before))
    ORDER BY te.last_checked_at ASC NULLS FIRST
    LIMIT sqlc.arg(batch_size)
    FOR UPDATE OF te SKIP LOCKED
)
UPDATE trunk_endpoints AS te
SET last_checked_at = sqlc.arg(checked_at)
FROM due
WHERE te.id = due.id
RETURNING te.*;

-- name: MarkTrunkEndpointHealthy :one
UPDATE trunk_endpoints
SET health_status = 'healthy',
    consecutive_failures = 0,
    last_checked_at = sqlc.arg(checked_at),
    last_response_code = sqlc.arg(response_code),
    last_latency_ms = sqlc.arg(latency_ms),
    last_error = NULL,
    cooldown_until = NULL
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: MarkTrunkEndpointProbeFailed :one
UPDATE trunk_endpoints
SET consecutive_failures = consecutive_failures + 1,
    health_status = CASE
        WHEN consecutive_failures + 1 >= sqlc.arg(failure_threshold) THEN 'unhealthy'
        ELSE health_status
    END,
    last_checked_at = sqlc.arg(checked_at),
    last_response_code = NULL,
    last_latency_ms = sqlc.arg(latency_ms),
    last_error = sqlc.arg(last_error),
    cooldown_until = CASE
        WHEN consecutive_failures + 1 >= sqlc.arg(failure_threshold) THEN sqlc.arg(cooldown_until)::TIMESTAMPTZ
        ELSE NULL
    END
WHERE id = sqlc.arg(id)
RETURNING *;
