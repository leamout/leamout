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
