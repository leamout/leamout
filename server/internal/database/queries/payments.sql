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
) VALUES (
    sqlc.arg(organization_id),
    sqlc.narg(invoice_id),
    sqlc.arg(provider),
    sqlc.narg(provider_payment_id),
    sqlc.arg(amount),
    sqlc.arg(currency),
    COALESCE(sqlc.narg(status), 'pending'),
    sqlc.narg(paid_at),
    COALESCE(sqlc.narg(metadata), '{}'::jsonb)
)
RETURNING *;

-- name: GetPayment :one
SELECT *
FROM payments
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetPaymentByProviderID :one
SELECT *
FROM payments
WHERE provider = sqlc.arg(provider)
  AND provider_payment_id = sqlc.arg(provider_payment_id)
LIMIT 1;

-- name: ListPaymentsByOrganization :many
SELECT *
FROM payments
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: ListPaymentsByInvoice :many
SELECT *
FROM payments
WHERE organization_id = sqlc.arg(organization_id)
  AND invoice_id = sqlc.arg(invoice_id)
ORDER BY created_at DESC;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET
    status = sqlc.arg(status),
    paid_at = COALESCE(sqlc.narg(paid_at), paid_at),
    metadata = COALESCE(sqlc.narg(metadata), metadata),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;
