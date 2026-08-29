-- name: CreatePayment :one
INSERT INTO payments (
    organization_id,
    invoice_id,
    provider,
    provider_payment_id,
    amount,
    currency,
    status,
    paid_at,
    metadata
)
SELECT
    o.id,
    sqlc.narg(invoice_id)::uuid,
    sqlc.arg(provider),
    sqlc.narg(provider_payment_id),
    sqlc.arg(amount),
    sqlc.arg(currency),
    COALESCE(sqlc.narg(status), 'pending'),
    sqlc.narg(paid_at),
    COALESCE(sqlc.narg(metadata), '{}'::jsonb)
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (
      sqlc.narg(invoice_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM invoices AS i
          WHERE i.id = sqlc.narg(invoice_id)::uuid
            AND i.organization_id = o.id
            AND i.currency = sqlc.arg(currency)
      )
  )
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

-- name: ListPaymentsByInvoice :many
SELECT p.*
FROM payments AS p
JOIN invoices AS i
  ON i.id = p.invoice_id
 AND i.organization_id = p.organization_id
JOIN organizations AS o ON o.id = p.organization_id
WHERE p.organization_id = sqlc.arg(organization_id)
  AND p.invoice_id = sqlc.arg(invoice_id)
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
RETURNING p.*;
