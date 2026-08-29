-- name: CreatePlan :one
INSERT INTO plans (
    product_id,
    code,
    name,
    description,
    active
)
SELECT
    p.id,
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.narg(description),
    COALESCE(sqlc.narg(active), true)
FROM products AS p
WHERE p.id = sqlc.arg(product_id)
  AND p.active = true
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
SELECT pl.*
FROM plans AS pl
JOIN products AS p ON p.id = pl.product_id
WHERE pl.product_id = sqlc.arg(product_id)
  AND pl.active = true
  AND p.active = true
ORDER BY pl.created_at DESC;

-- name: UpdatePlan :one
UPDATE plans AS pl
SET
    code = COALESCE(sqlc.narg(code), pl.code),
    name = COALESCE(sqlc.narg(name), pl.name),
    description = COALESCE(sqlc.narg(description), pl.description),
    active = COALESCE(sqlc.narg(active), pl.active),
    updated_at = NOW()
FROM products AS p
WHERE pl.id = sqlc.arg(id)
  AND p.id = pl.product_id
  AND (
      COALESCE(sqlc.narg(active), pl.active) = false
      OR p.active = true
  )
RETURNING pl.*;
