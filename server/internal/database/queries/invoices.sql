-- name: CreateInvoice :one
INSERT INTO invoices (
    organization_id,
    subscription_id,
    invoice_number,
    currency,
    subtotal,
    tax,
    total,
    status,
    issued_at,
    due_at,
    paid_at,
    metadata
) VALUES (
    sqlc.arg(organization_id),
    sqlc.narg(subscription_id),
    sqlc.arg(invoice_number),
    sqlc.arg(currency),
    sqlc.arg(subtotal),
    COALESCE(sqlc.narg(tax), 0),
    sqlc.arg(total),
    COALESCE(sqlc.narg(status), 'draft'),
    sqlc.narg(issued_at),
    sqlc.narg(due_at),
    sqlc.narg(paid_at),
    COALESCE(sqlc.narg(metadata), '{}'::jsonb)
)
RETURNING *;

-- name: GetInvoice :one
SELECT *
FROM invoices
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetInvoiceByNumber :one
SELECT *
FROM invoices
WHERE organization_id = sqlc.arg(organization_id)
  AND invoice_number = sqlc.arg(invoice_number)
LIMIT 1;

-- name: ListInvoicesByOrganization :many
SELECT *
FROM invoices
WHERE organization_id = sqlc.arg(organization_id)
ORDER BY created_at DESC;

-- name: UpdateInvoiceStatus :one
UPDATE invoices
SET
    status = sqlc.arg(status),
    paid_at = COALESCE(sqlc.narg(paid_at), paid_at),
    updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
RETURNING *;
