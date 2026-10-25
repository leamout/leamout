-- name: PrepareCallOrigination :exec
INSERT INTO call_origination_attempts (
    channel_id, call_id, organization_id, trunk_id, trunk_endpoint_id
) VALUES (
    sqlc.arg(channel_id), sqlc.arg(call_id),
    sqlc.arg(organization_id), sqlc.arg(trunk_id), sqlc.arg(trunk_endpoint_id)
);

-- name: MarkCallOriginationSubmitted :exec
UPDATE call_origination_attempts
SET status = 'submitted', updated_at = NOW()
WHERE channel_id = sqlc.arg(channel_id)
  AND status = 'prepared';

-- name: MarkCallOriginationUncertain :exec
UPDATE call_origination_attempts
SET status = 'uncertain', updated_at = NOW()
WHERE channel_id = sqlc.arg(channel_id)
  AND status IN ('prepared', 'submitted');

-- name: MarkCallOriginationObserved :exec
UPDATE call_origination_attempts
SET status = 'observed', updated_at = NOW()
WHERE call_id = sqlc.arg(call_id)
  AND channel_id = sqlc.arg(channel_id)
  AND status IN ('prepared', 'submitted', 'uncertain');

-- name: MarkCallOriginationFinished :exec
UPDATE call_origination_attempts
SET status = 'finished', updated_at = NOW()
WHERE call_id = sqlc.arg(call_id)
  AND channel_id = sqlc.arg(channel_id)
  AND status <> 'finished';

-- name: ListUnresolvedCallOriginations :many
SELECT
    a.channel_id, a.call_id, a.organization_id, a.trunk_id,
    a.status, a.created_at
FROM call_origination_attempts AS a
JOIN calls AS c ON c.id = a.call_id AND c.organization_id = a.organization_id
WHERE a.status IN ('prepared', 'submitted', 'uncertain')
  AND c.state = 'initiating'
ORDER BY a.created_at ASC, a.channel_id ASC
LIMIT 100;

-- name: MarkCallOriginationObservedForChannel :exec
UPDATE call_origination_attempts
SET status = 'observed', updated_at = NOW()
WHERE channel_id = sqlc.arg(channel_id)
  AND status IN ('prepared', 'submitted', 'uncertain');

-- name: MarkCallOriginationFinishedForCall :exec
UPDATE call_origination_attempts
SET status = 'finished', updated_at = NOW()
WHERE call_id = sqlc.arg(call_id)
  AND status IN ('prepared', 'submitted', 'uncertain', 'observed');

-- name: GetCallOriginationByChannel :one
SELECT call_id, organization_id, trunk_id, status
FROM call_origination_attempts
WHERE channel_id = sqlc.arg(channel_id)
LIMIT 1;
