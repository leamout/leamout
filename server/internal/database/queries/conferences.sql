-- name: CreateConference :one
INSERT INTO conferences (
    tenant_id,
    application_id,
    name,
    state,
    started_at
) VALUES (
    sqlc.arg(tenant_id),
    sqlc.narg(application_id),
    sqlc.arg(name),
    COALESCE(sqlc.narg(state), 'active'),
    sqlc.narg(started_at)
)
RETURNING *;

-- name: GetConference :one
SELECT *
FROM conferences
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetConferenceByName :one
SELECT *
FROM conferences
WHERE tenant_id = sqlc.arg(tenant_id)
  AND name = sqlc.arg(name)
LIMIT 1;

-- name: ListConferences :many
SELECT *
FROM conferences
WHERE tenant_id = sqlc.arg(tenant_id)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit)
OFFSET sqlc.arg(offset);

-- name: UpdateConferenceState :one
UPDATE conferences
SET
    state = sqlc.arg(state),
    ended_at = CASE
        WHEN sqlc.arg(state)::text = 'ended' THEN COALESCE(ended_at, NOW())
        ELSE ended_at
    END,
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
RETURNING *;
