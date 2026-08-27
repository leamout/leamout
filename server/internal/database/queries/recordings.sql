-- name: CreateRecording :one
INSERT INTO recordings (
    organization_id,
    call_id,
    status,
    storage_key,
    storage_provider,
    storage_bucket,
    storage_url,
    file_size_bytes,
    format,
    started_at
) VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(call_id),
    COALESCE(sqlc.narg(status), 'recording'),
    sqlc.narg(storage_key),
    sqlc.narg(storage_provider),
    sqlc.narg(storage_bucket),
    sqlc.narg(storage_url),
    sqlc.narg(file_size_bytes),
    sqlc.narg(format),
    sqlc.narg(started_at)
)
RETURNING *;

-- name: GetRecording :one
SELECT *
FROM recordings
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: ListCallRecordings :many
SELECT *
FROM recordings
WHERE organization_id = sqlc.arg(organization_id)
  AND call_id = sqlc.arg(call_id)
ORDER BY created_at DESC;

-- name: CompleteRecording :one
UPDATE recordings
SET
    status = 'completed',
    storage_key = COALESCE(sqlc.narg(storage_key), storage_key),
    storage_provider = COALESCE(sqlc.narg(storage_provider), storage_provider),
    storage_bucket = COALESCE(sqlc.narg(storage_bucket), storage_bucket),
    storage_url = COALESCE(sqlc.narg(storage_url), storage_url),
    file_size_bytes = COALESCE(sqlc.narg(file_size_bytes), file_size_bytes),
    format = COALESCE(sqlc.narg(format), format),
    duration_seconds = COALESCE(sqlc.narg(duration_seconds), duration_seconds),
    completed_at = COALESCE(completed_at, NOW()),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = 'recording'
RETURNING *;

-- name: FailRecording :one
UPDATE recordings
SET
    status = 'failed',
    completed_at = COALESCE(completed_at, NOW()),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = 'recording'
RETURNING *;
