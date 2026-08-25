-- name: CreateRecording :one
INSERT INTO recordings (
    tenant_id,
    call_id,
    status,
    storage_key,
    format,
    started_at
) VALUES (
    sqlc.arg(tenant_id),
    sqlc.arg(call_id),
    COALESCE(sqlc.narg(status), 'recording'),
    sqlc.narg(storage_key),
    sqlc.narg(format),
    sqlc.narg(started_at)
)
RETURNING *;

-- name: GetRecording :one
SELECT *
FROM recordings
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: ListCallRecordings :many
SELECT *
FROM recordings
WHERE tenant_id = sqlc.arg(tenant_id)
  AND call_id = sqlc.arg(call_id)
ORDER BY created_at DESC;

-- name: CompleteRecording :one
UPDATE recordings
SET
    status = 'completed',
    storage_key = COALESCE(sqlc.narg(storage_key), storage_key),
    format = COALESCE(sqlc.narg(format), format),
    duration_seconds = sqlc.narg(duration_seconds),
    completed_at = COALESCE(completed_at, NOW()),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND status = 'recording'
RETURNING *;

-- name: FailRecording :one
UPDATE recordings
SET
    status = 'failed',
    completed_at = COALESCE(completed_at, NOW()),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
  AND status = 'recording'
RETURNING *;
