-- name: CreatePlan :one
INSERT INTO plans (
    product_id,
    code,
    name,
    description,
    active
) VALUES (
    sqlc.arg(product_id),
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.narg(description),
    COALESCE(sqlc.narg(active), true)
)
RETURNING *;

-- name: GetPlanByID :one
SELECT *
FROM plans
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetPlanByCode :one
SELECT *
FROM plans
WHERE code = sqlc.arg(code)
LIMIT 1;

-- name: ListPlansByProduct :many
SELECT *
FROM plans
WHERE product_id = sqlc.arg(product_id)
ORDER BY created_at DESC;

-- name: ListActivePlansByProduct :many
SELECT *
FROM plans
WHERE product_id = sqlc.arg(product_id)
  AND active = true
ORDER BY created_at DESC;

-- name: UpdatePlan :one
UPDATE plans
SET
    code = COALESCE(sqlc.narg(code), code),
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    active = COALESCE(sqlc.narg(active), active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;
