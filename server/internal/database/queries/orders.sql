-- name: CreateOrderFromCheckoutPayment :one
INSERT INTO orders (
    organization_id,
    checkout_id,
    payment_id,
    wallet_id,
    price_id,
    order_type,
    amount_minor,
    currency,
    completed_at,
    metadata
)
SELECT
    c.organization_id,
    c.id,
    p.id,
    c.wallet_id,
    c.price_id,
    c.checkout_type,
    c.amount_minor,
    c.currency,
    c.completed_at,
    c.metadata
FROM checkouts AS c
JOIN payments AS p
  ON p.checkout_id = c.id
 AND p.organization_id = c.organization_id
WHERE c.id = sqlc.arg(checkout_id)
  AND p.id = sqlc.arg(payment_id)
  AND c.organization_id = sqlc.arg(organization_id)
  AND c.status = 'succeeded'
  AND c.completed_at IS NOT NULL
  AND p.status = 'succeeded'
ON CONFLICT (checkout_id) DO NOTHING
RETURNING *;

-- name: GetOrder :one
SELECT o.*
FROM orders AS o
JOIN organizations AS org ON org.id = o.organization_id
WHERE o.organization_id = sqlc.arg(organization_id)
  AND o.id = sqlc.arg(id)
  AND org.status = 'active'
  AND org.deleted_at IS NULL
LIMIT 1;

-- name: ListOrdersByOrganization :many
SELECT o.*
FROM orders AS o
JOIN organizations AS org ON org.id = o.organization_id
WHERE o.organization_id = sqlc.arg(organization_id)
  AND org.status = 'active'
  AND org.deleted_at IS NULL
ORDER BY o.created_at DESC;
