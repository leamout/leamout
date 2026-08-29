-- name: CreateMeter :one
INSERT INTO meters (
    key,
    name,
    unit,
    active
) VALUES (
    sqlc.arg(key),
    sqlc.arg(name),
    sqlc.arg(unit),
    COALESCE(sqlc.narg(active), true)
)
RETURNING *;

-- name: GetMeterByID :one
SELECT *
FROM meters
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetMeterByKey :one
SELECT *
FROM meters
WHERE key = sqlc.arg(key)
LIMIT 1;

-- name: ListMeters :many
SELECT *
FROM meters
ORDER BY created_at DESC;

-- name: ListActiveMeters :many
SELECT *
FROM meters
WHERE active = true
ORDER BY created_at DESC;

-- name: UpdateMeter :one
UPDATE meters
SET
    name = COALESCE(sqlc.narg(name), name),
    unit = COALESCE(sqlc.narg(unit), unit),
    active = COALESCE(sqlc.narg(active), active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;
