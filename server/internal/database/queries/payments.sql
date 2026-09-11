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
    c.organization_id AS organization_id,
    sqlc.arg(provider) AS provider,
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
  AND c.status = 'processing'
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

-- name: GetPaymentByCheckout :one
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
SET status = sqlc.arg(status),
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

-- name: LockPaymentByReference :one
SELECT
    c.id AS checkout_id,
    c.organization_id AS organization_id,
    c.reference AS reference,
    c.amount_minor AS amount_minor,
    c.currency AS currency,
    c.status AS checkout_status,
    p.id AS payment_id,
    p.provider AS provider,
    p.provider_payment_id AS provider_payment_id,
    p.status AS payment_status
FROM checkouts AS c
JOIN payments AS p
  ON p.checkout_id = c.id
 AND p.organization_id = c.organization_id
WHERE c.reference = sqlc.arg(reference)
FOR UPDATE OF c, p;

-- name: InsertPaymentProviderEvent :one
INSERT INTO payment_provider_events (
    payment_id,
    organization_id,
    provider,
    provider_event_id,
    event_type,
    payload_sha256,
    payload
)
VALUES (
    sqlc.arg(payment_id),
    sqlc.arg(organization_id),
    sqlc.arg(provider),
    sqlc.arg(provider_event_id),
    sqlc.arg(event_type),
    sqlc.arg(payload_sha256),
    sqlc.arg(payload)
)
ON CONFLICT (provider, provider_event_id) DO NOTHING
RETURNING *;

-- name: MarkPaymentProviderEventProcessed :exec
UPDATE payment_provider_events
SET processed_at = NOW()
WHERE id = sqlc.arg(id)
  AND processed_at IS NULL;

-- name: GetPaymentProviderEvent :one
SELECT *
FROM payment_provider_events
WHERE provider = sqlc.arg(provider)
  AND provider_event_id = sqlc.arg(provider_event_id)
LIMIT 1;
