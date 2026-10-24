-- name: CreateCall :one
INSERT INTO calls (
    organization_id,
    voice_agent_id,
    direction,
    state,
    from_uri,
    to_uri,
    sip_call_id
) VALUES (
    sqlc.arg(organization_id),
    sqlc.narg(voice_agent_id),
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
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetCallBySIPCallID :one
SELECT *
FROM calls
WHERE organization_id = sqlc.arg(organization_id)
  AND sip_call_id = sqlc.arg(sip_call_id)
LIMIT 1;

-- name: GetCallBySIPCallIDGlobal :one
SELECT *
FROM calls
WHERE sip_call_id = sqlc.arg(sip_call_id)
LIMIT 1;

-- name: GetCallLifecycleSnapshot :one
SELECT
    c.organization_id,
    c.trunk_id,
    c.direction,
    c.state,
    c.media_state,
    c.answered_at
FROM calls AS c
WHERE c.id = sqlc.arg(id)
LIMIT 1;

-- name: ListCalls :many
SELECT *
FROM calls
WHERE organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(state)::text IS NULL OR state = sqlc.narg(state)::text)
  AND (sqlc.narg(direction)::text IS NULL OR direction = sqlc.narg(direction)::text)
  AND (sqlc.narg(trunk_id)::uuid IS NULL OR trunk_id = sqlc.narg(trunk_id)::uuid)
  AND (sqlc.narg(voice_agent_id)::uuid IS NULL OR voice_agent_id = sqlc.narg(voice_agent_id)::uuid)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_before)::timestamptz IS NULL OR created_at < sqlc.narg(created_before)::timestamptz)
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: ListCallsForReconciliation :many
SELECT *
FROM calls
WHERE state IN ('initiating', 'ringing', 'answered', 'active')
  AND sip_call_id IS NOT NULL
  AND updated_at <= sqlc.arg(updated_before)
ORDER BY updated_at ASC
LIMIT sqlc.arg(batch_size);

-- name: ListActiveCallsForAdmissionReconciliation :many
SELECT
    id,
    organization_id,
    trunk_id
FROM calls
WHERE state IN ('initiating', 'ringing', 'answered', 'active')
  AND trunk_id IS NOT NULL
ORDER BY created_at ASC;

-- Revalidate the DID-derived tenant and route tuple before the call service
-- persists or admits an inbound call.
-- name: GetInboundCallContext :one
SELECT
    t.max_cps,
    t.max_concurrent_calls
FROM phone_numbers AS pn
JOIN trunks AS t
  ON t.id = pn.trunk_id
JOIN voice_agent_bindings AS binding
  ON binding.phone_number_id = pn.id
 AND binding.organization_id = pn.organization_id
JOIN voice_agents AS agent
  ON agent.id = binding.voice_agent_id
 AND agent.organization_id = binding.organization_id
JOIN organizations AS o
  ON o.id = pn.organization_id
WHERE pn.id = sqlc.arg(phone_number_id)
  AND pn.organization_id = sqlc.arg(organization_id)
  AND pn.number = sqlc.arg(called_number)
  AND pn.trunk_id = sqlc.arg(trunk_id)
  AND pn.status = 'active'
  AND pn.voice_enabled = true
  AND t.status = 'active'
  AND t.inbound_enabled = true
  AND t.organization_id = pn.organization_id
  AND binding.id = sqlc.arg(voice_agent_binding_id)
  AND agent.id = sqlc.arg(voice_agent_id)
  AND agent.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: SetCallRouteAttribution :one
UPDATE calls
SET
    trunk_id = sqlc.arg(trunk_id),
    trunk_endpoint_id = sqlc.arg(trunk_endpoint_id),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;

-- name: SetOutboundCallSIPCallID :one
-- Bind the actual FreeSWITCH SIP dialog identity after asynchronous originate.
-- Replayed events with the same identity are idempotent; a different Call-ID
-- must never overwrite a previously attributed dialog.
UPDATE calls
SET
    sip_call_id = sqlc.arg(sip_call_id),
    updated_at = CASE WHEN sip_call_id IS NULL THEN NOW() ELSE updated_at END
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND direction = 'outbound'
  AND (sip_call_id IS NULL OR sip_call_id = sqlc.arg(sip_call_id))
RETURNING id;

-- name: UpdateCallState :one
UPDATE calls
SET
    state = sqlc.arg(state),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;

-- name: MarkCallRinging :one
UPDATE calls
SET state = 'ringing', updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND state = 'initiating'
RETURNING *;

-- name: MarkCallAnswered :one
UPDATE calls
SET
    state = 'answered',
    answered_at = COALESCE(answered_at, NOW()),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND state IN ('initiating', 'ringing')
RETURNING *;

-- name: MarkCallActive :one
UPDATE calls
SET state = 'active', updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND state IN ('answered', 'ringing')
RETURNING *;

-- name: MarkCallHeld :one
UPDATE calls
SET
    media_state = 'held',
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND state IN ('answered', 'active')
  AND media_state = 'active'
RETURNING *;

-- name: MarkCallResumed :one
UPDATE calls
SET
    media_state = 'active',
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND state IN ('answered', 'active')
  AND media_state = 'held'
RETURNING *;

-- name: MarkCallCompleted :one
UPDATE calls
SET
    state = 'completed',
    ended_at = COALESCE(ended_at, NOW()),
    hangup_reason = COALESCE(sqlc.narg(hangup_reason), hangup_reason),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
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
WHERE organization_id = sqlc.arg(organization_id)
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
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND state IN ('initiating', 'ringing')
RETURNING *;

-- Operator projections intentionally include disabled and deleted organizations.
-- name: ListBackofficeCalls :many
SELECT
    c.id,
    c.organization_id,
    o.name AS organization_name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::text AS organization_status,
    c.voice_agent_id, COALESCE(va.name, '')::text AS agent_name,
    c.direction,
    c.state,
    c.from_uri,
    c.to_uri,
    c.created_at
FROM calls AS c
JOIN organizations AS o ON o.id = c.organization_id
LEFT JOIN voice_agents AS va ON va.id = c.voice_agent_id AND va.organization_id = c.organization_id
WHERE (
    sqlc.arg(search)::text = ''
    OR c.id::text = sqlc.arg(search)::text
    OR c.sip_call_id = sqlc.arg(search)::text
    OR strpos(lower(c.from_uri), lower(sqlc.arg(search)::text)) > 0
    OR strpos(lower(c.to_uri), lower(sqlc.arg(search)::text)) > 0
)
  AND (sqlc.arg(state)::text = '' OR c.state = sqlc.arg(state)::text)
  AND (sqlc.arg(direction)::text = '' OR c.direction = sqlc.arg(direction)::text)
  AND (sqlc.narg(organization_id)::uuid IS NULL OR c.organization_id = sqlc.narg(organization_id)::uuid)
  AND (sqlc.narg(agent_id)::uuid IS NULL OR c.voice_agent_id = sqlc.narg(agent_id)::uuid)
  AND (sqlc.narg(created_from)::timestamptz IS NULL OR c.created_at >= sqlc.narg(created_from)::timestamptz)
  AND (sqlc.narg(created_before)::timestamptz IS NULL OR c.created_at < sqlc.narg(created_before)::timestamptz)
ORDER BY c.created_at DESC,
    c.id DESC
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: GetBackofficeCall :one
SELECT
    c.id,
    c.organization_id,
    o.name AS organization_name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::text AS organization_status,
    c.voice_agent_id, COALESCE(va.name, '')::text AS agent_name,
    c.direction,
    c.state,
    c.from_uri,
    c.to_uri,
    c.created_at,
    c.media_state,
    c.started_at,
    c.answered_at,
    c.ended_at,
    c.updated_at,
    COALESCE(c.sip_call_id, '')::text AS sip_call_id,
    COALESCE(c.hangup_reason, '')::text AS hangup_reason,
    c.trunk_id,
    c.trunk_endpoint_id
FROM calls AS c
JOIN organizations AS o ON o.id = c.organization_id
LEFT JOIN voice_agents AS va ON va.id = c.voice_agent_id AND va.organization_id = c.organization_id
WHERE c.id = sqlc.arg(id);
