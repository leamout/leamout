-- name: CreatePayment :one
INSERT INTO payments (
    checkout_id,
    organization_id,
    provider,
    provider_payment_id,
    amount_minor,
    currency,
    status,
    paid_at,
    metadata
)
SELECT
    c.id AS checkout_id,
    c.organization_id,
    c.provider,
    sqlc.narg(provider_payment_id) AS provider_payment_id,
    sqlc.arg(amount_minor) AS amount_minor,
    sqlc.arg(currency) AS currency,
    COALESCE(sqlc.narg(status), 'pending') AS status,
    sqlc.narg(paid_at) AS paid_at,
    COALESCE(sqlc.narg(metadata), '{}'::jsonb) AS metadata
FROM checkouts AS c
JOIN organizations AS o ON o.id = c.organization_id
WHERE c.id = sqlc.arg(checkout_id)
  AND c.organization_id = sqlc.arg(organization_id)
  AND c.provider = sqlc.arg(provider)
  AND c.amount_minor = sqlc.arg(amount_minor)
  AND c.currency = sqlc.arg(currency)
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

-- name: GetPaymentByCheckoutOrder :one
SELECT p.*
FROM payments AS p
JOIN organizations AS o ON o.id = p.organization_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.checkout_id = sqlc.arg(checkout_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

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
RETURNING p.*;

-- name: SetPaymentProviderID :one
UPDATE payments AS p
SET provider_payment_id = sqlc.arg(provider_payment_id),
    status = sqlc.arg(status),
    updated_at = NOW()
FROM organizations AS o
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.id = sqlc.arg(id)
  AND p.provider_payment_id IS NULL
  AND o.id = p.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING p.*;
