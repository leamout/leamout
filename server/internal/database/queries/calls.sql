-- name: CreateCall :one
INSERT INTO calls (
    tenant_id,
    application_id,
    direction,
    state,
    from_uri,
    to_uri,
    sip_call_id
) VALUES (
    sqlc.arg(tenant_id),
    sqlc.narg(application_id),
    sqlc.arg(direction),
    COALESCE(sqlc.narg(state), 'initiating'),
    sqlc.arg(from_uri),
    sqlc.arg(to_uri),
    sqlc.narg(sip_call_id)
)
RETURNING *;

-- name: GetCall :one
SELECT *
FROM calls
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetCallBySIPCallID :one
SELECT *
FROM calls
WHERE tenant_id = sqlc.arg(tenant_id)
  AND sip_call_id = sqlc.arg(sip_call_id)
LIMIT 1;

-- name: ListCalls :many
SELECT *
FROM calls
WHERE tenant_id = sqlc.arg(tenant_id)
  AND (sqlc.narg(state)::text IS NULL OR state = sqlc.narg(state)::text)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit)
OFFSET sqlc.arg(offset);

-- name: UpdateCallState :one
UPDATE calls
SET
    state = sqlc.arg(state),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
RETURNING *;

-- name: MarkCallRinging :one
UPDATE calls
SET state = 'ringing', updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND state = 'initiating'
RETURNING *;

-- name: MarkCallAnswered :one
UPDATE calls
SET
    state = 'answered',
    answered_at = COALESCE(answered_at, NOW()),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND state IN ('initiating', 'ringing')
RETURNING *;

-- name: MarkCallActive :one
UPDATE calls
SET state = 'active', updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND state IN ('answered', 'ringing')
RETURNING *;

-- name: MarkCallCompleted :one
UPDATE calls
SET
    state = 'completed',
    ended_at = COALESCE(ended_at, NOW()),
    hangup_reason = COALESCE(sqlc.narg(hangup_reason), hangup_reason),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND state NOT IN ('completed', 'failed', 'cancelled')
RETURNING *;

-- name: MarkCallFailed :one
UPDATE calls
SET
    state = 'failed',
    ended_at = COALESCE(ended_at, NOW()),
    hangup_reason = COALESCE(sqlc.narg(hangup_reason), hangup_reason),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND state NOT IN ('completed', 'failed', 'cancelled')
RETURNING *;

-- name: MarkCallCancelled :one
UPDATE calls
SET
    state = 'cancelled',
    ended_at = COALESCE(ended_at, NOW()),
    hangup_reason = COALESCE(sqlc.narg(hangup_reason), hangup_reason),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND state IN ('initiating', 'ringing')
RETURNING *;
