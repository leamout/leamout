-- name: CreateProduct :one
INSERT INTO products (
    code,
    name,
    description,
    active
) VALUES (
    sqlc.arg(code),
    sqlc.arg(name),
    sqlc.narg(description),
    COALESCE(sqlc.narg(active), true)
)
RETURNING *;

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

-- name: UpdateProduct :one
UPDATE products
SET
    code = COALESCE(sqlc.narg(code), code),
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    active = COALESCE(sqlc.narg(active), active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;
