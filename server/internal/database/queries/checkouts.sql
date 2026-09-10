-- name: CreateCheckoutOrder :one
INSERT INTO checkouts (
    organization_id, wallet_id, price_id, checkout_type,
    provider, payment_method, reference, amount_minor, currency, expires_at, metadata
)
SELECT
    sqlc.arg(organization_id) AS organization_id,
    sqlc.narg(wallet_id)::UUID AS wallet_id,
    sqlc.narg(price_id)::UUID AS price_id,
    sqlc.arg(checkout_type) AS checkout_type,
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
SELECT c.*
FROM checkouts AS c
JOIN organizations AS o ON o.id = c.organization_id
WHERE c.organization_id = sqlc.arg(organization_id)
  AND c.id = sqlc.arg(id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: GetCheckoutOrderByReference :one
SELECT c.*
FROM checkouts AS c
JOIN organizations AS o ON o.id = c.organization_id
WHERE c.reference = sqlc.arg(reference)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: CompareAndSetCheckoutOrderState :one
WITH updated AS (
    UPDATE checkouts AS c
    SET status = sqlc.arg(status),
        next_action = sqlc.arg(next_action),
        provider_message = sqlc.narg(provider_message),
        completed_at = sqlc.narg(completed_at),
        updated_at = NOW()
    WHERE c.organization_id = sqlc.arg(organization_id)
      AND c.id = sqlc.arg(id)
      AND c.status = sqlc.arg(expected_status)
    RETURNING c.*
), created_order AS (
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
    FROM updated AS c
    JOIN payments AS p
      ON p.checkout_id = c.id
     AND p.organization_id = c.organization_id
    WHERE c.status = 'succeeded'
      AND c.completed_at IS NOT NULL
      AND p.status = 'succeeded'
    ON CONFLICT (checkout_id) DO NOTHING
    RETURNING id
)
SELECT c.*
FROM checkouts AS c
JOIN updated AS u
  ON u.id = c.id
 AND u.organization_id = c.organization_id
LEFT JOIN created_order AS o ON TRUE;

-- name: ExpireCheckoutOrders :many
UPDATE checkouts
SET status = 'expired', next_action = 'none', completed_at = NOW(), updated_at = NOW()
WHERE status IN ('pending', 'processing')
  AND expires_at <= NOW()
RETURNING *;

-- name: ClaimCheckoutOrderRefresh :one
UPDATE checkouts
SET updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND provider = 'paystack'
  AND status = 'processing'
  AND updated_at <= sqlc.arg(refresh_before)
RETURNING *;
