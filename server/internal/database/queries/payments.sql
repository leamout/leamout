-- name: CreatePayment :one
INSERT INTO payments (
    checkout_order_id,
    organization_id,
    provider,
    provider_payment_id,
    amount,
    currency,
    status,
    paid_at,
    metadata
)
SELECT
    co.id AS checkout_order_id,
    co.organization_id,
    co.provider,
    sqlc.narg(provider_payment_id) AS provider_payment_id,
    sqlc.arg(amount) AS amount,
    sqlc.arg(currency) AS currency,
    COALESCE(sqlc.narg(status), 'pending') AS status,
    sqlc.narg(paid_at) AS paid_at,
    COALESCE(sqlc.narg(metadata), '{}'::jsonb) AS metadata
FROM checkout_orders AS co
JOIN organizations AS o ON o.id = co.organization_id
WHERE co.id = sqlc.arg(checkout_order_id)
  AND co.organization_id = sqlc.arg(organization_id)
  AND co.provider = sqlc.arg(provider)
  AND co.amount = sqlc.arg(amount)
  AND co.currency = sqlc.arg(currency)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: GetPayment :one
SELECT p.*
FROM payments AS p
JOIN organizations AS o ON o.id = p.organization_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.id = sqlc.arg(id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: GetPaymentByProviderID :one
SELECT p.*
FROM payments AS p
JOIN organizations AS o ON o.id = p.organization_id
WHERE p.provider = sqlc.arg(provider)
  AND p.provider_payment_id = sqlc.arg(provider_payment_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListPaymentsByOrganization :many
SELECT p.*
FROM payments AS p
JOIN organizations AS o ON o.id = p.organization_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY p.created_at DESC;

-- name: UpdatePaymentStatus :one
UPDATE payments AS p
SET
    status = sqlc.arg(status),
    paid_at = COALESCE(sqlc.narg(paid_at), p.paid_at),
    metadata = COALESCE(sqlc.narg(metadata), p.metadata),
    updated_at = NOW()
FROM organizations AS o
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.id = sqlc.arg(id)
  AND o.id = p.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;
