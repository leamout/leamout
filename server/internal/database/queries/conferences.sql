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
    sqlc.narg(started_at)
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

-- name: UpdateConferenceState :one
UPDATE conferences
SET
    state = sqlc.arg(state),
    ended_at = CASE
        WHEN sqlc.arg(state)::text = 'ended' THEN COALESCE(ended_at, NOW())
        ELSE ended_at
    END,
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;
