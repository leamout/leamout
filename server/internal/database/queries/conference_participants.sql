-- name: CreateConferenceParticipant :one
INSERT INTO conference_participants (
    tenant_id,
    conference_id,
    call_participant_id,
    state,
    muted,
    deaf,
    speaking,
    joined_at
) VALUES (
    sqlc.arg(tenant_id),
    sqlc.arg(conference_id),
    sqlc.narg(call_participant_id),
    COALESCE(sqlc.narg(state), 'joining'),
    COALESCE(sqlc.narg(muted), false),
    COALESCE(sqlc.narg(deaf), false),
    COALESCE(sqlc.narg(speaking), false),
    sqlc.narg(joined_at)
)
RETURNING *;

-- name: GetConferenceParticipant :one
SELECT *
FROM conference_participants
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: ListConferenceParticipants :many
SELECT *
FROM conference_participants
WHERE tenant_id = sqlc.arg(tenant_id)
  AND conference_id = sqlc.arg(conference_id)
ORDER BY created_at ASC;

-- name: UpdateConferenceParticipantState :one
UPDATE conference_participants
SET
    state = sqlc.arg(state),
    muted = COALESCE(sqlc.narg(muted), muted),
    deaf = COALESCE(sqlc.narg(deaf), deaf),
    speaking = COALESCE(sqlc.narg(speaking), speaking),
    joined_at = CASE
        WHEN sqlc.arg(state)::text = 'joined' THEN COALESCE(joined_at, NOW())
        ELSE joined_at
    END,
    left_at = CASE
        WHEN sqlc.arg(state)::text IN ('left', 'failed') THEN COALESCE(left_at, NOW())
        ELSE left_at
    END,
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
RETURNING *;
