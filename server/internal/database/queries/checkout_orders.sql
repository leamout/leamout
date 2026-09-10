-- name: CreateCheckoutOrder :one
INSERT INTO checkout_orders (
    organization_id, wallet_id, price_id, order_type,
    provider, payment_method, reference, amount_minor, currency, expires_at, metadata
)
SELECT
    sqlc.arg(organization_id) AS organization_id,
    sqlc.narg(wallet_id)::UUID AS wallet_id,
    sqlc.narg(price_id)::UUID AS price_id,
    sqlc.arg(order_type) AS order_type,
    sqlc.arg(provider) AS provider,
    sqlc.arg(payment_method) AS payment_method,
    sqlc.arg(reference) AS reference,
    sqlc.arg(amount_minor) AS amount_minor,
    sqlc.arg(currency) AS currency,
    sqlc.arg(expires_at) AS expires_at,
    COALESCE(sqlc.narg(metadata), '{}'::jsonb) AS metadata
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: GetCheckoutOrder :one
SELECT co.*
FROM checkout_orders AS co
JOIN organizations AS o ON o.id = co.organization_id
WHERE co.organization_id = sqlc.arg(organization_id)
  AND co.id = sqlc.arg(id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: GetCheckoutOrderByReference :one
SELECT co.*
FROM checkout_orders AS co
JOIN organizations AS o ON o.id = co.organization_id
WHERE co.reference = sqlc.arg(reference)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: CompareAndSetCheckoutOrderState :one
UPDATE checkout_orders
SET status = sqlc.arg(status),
    next_action = sqlc.arg(next_action),
    provider_message = sqlc.narg(provider_message),
    completed_at = sqlc.narg(completed_at),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = sqlc.arg(expected_status)
RETURNING *;

-- name: ExpireCheckoutOrders :many
UPDATE checkout_orders
SET status = 'expired', next_action = 'none', completed_at = NOW(), updated_at = NOW()
WHERE status IN ('pending', 'processing')
  AND expires_at <= NOW()
RETURNING *;

-- name: ClaimCheckoutOrderRefresh :one
UPDATE checkout_orders
SET updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND provider = 'paystack'
  AND status = 'processing'
  AND updated_at <= sqlc.arg(refresh_before)
RETURNING *;
