-- name: CreateConference :one
INSERT INTO conferences (
    organization_id,
    application_id,
    name,
    state,
    started_at
) VALUES (
    sqlc.arg(organization_id),
    sqlc.narg(application_id),
    sqlc.arg(name),
    COALESCE(sqlc.narg(state), 'active'),
    COALESCE(sqlc.narg(started_at), NOW())
)
RETURNING *;

-- name: GetConference :one
SELECT *
FROM conferences
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetConferenceByName :one
SELECT *
FROM conferences
WHERE organization_id = sqlc.arg(organization_id)
  AND name = sqlc.arg(name)
LIMIT 1;

-- name: ListConferences :many
SELECT *
FROM conferences
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: EndConference :one
UPDATE conferences
SET
    state = 'ended',
    ended_at = COALESCE(ended_at, NOW()),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND state = 'active'
RETURNING *;

-- name: EndConferenceParticipants :many
UPDATE conference_participants
SET
    state = 'left',
    left_at = COALESCE(left_at, NOW()),
    muted = false,
    deaf = false,
    speaking = false,
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND conference_id = sqlc.arg(conference_id)
  AND state IN ('joining', 'joined')
RETURNING *;
