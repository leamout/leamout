-- name: GetProductByID :one
SELECT *
FROM products
WHERE id = sqlc.arg(id)
LIMIT 1;

-- name: GetProductByCode :one
SELECT *
FROM products
WHERE code = sqlc.arg(code)
LIMIT 1;

-- name: ListProducts :many
SELECT *
FROM products
ORDER BY created_at DESC;

-- name: ListActiveProducts :many
SELECT *
FROM products
WHERE active = true
ORDER BY created_at DESC;
