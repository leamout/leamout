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
)
SELECT
    o.id AS organization_id,
    sqlc.narg(subscription_id)::uuid AS subscription_id,
    sqlc.arg(invoice_number) AS invoice_number,
    sqlc.arg(currency) AS currency,
    sqlc.arg(subtotal) AS subtotal,
    COALESCE(sqlc.narg(tax), 0) AS tax,
    sqlc.arg(total) AS total,
    COALESCE(sqlc.narg(status), 'draft') AS status,
    sqlc.narg(issued_at) AS issued_at,
    sqlc.narg(due_at) AS due_at,
    sqlc.narg(paid_at) AS paid_at,
    COALESCE(sqlc.narg(metadata), '{}'::jsonb) AS metadata
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (
      sqlc.narg(subscription_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM subscriptions AS s
          WHERE s.id = sqlc.narg(subscription_id)::uuid
            AND s.organization_id = o.id
      )
  )
RETURNING *;

-- name: GetInvoice :one
SELECT i.*
FROM invoices AS i
JOIN organizations AS o ON o.id = i.organization_id
WHERE i.organization_id = sqlc.arg(organization_id)
  AND i.id = sqlc.arg(id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: GetInvoiceByNumber :one
SELECT i.*
FROM invoices AS i
JOIN organizations AS o ON o.id = i.organization_id
WHERE i.organization_id = sqlc.arg(organization_id)
  AND i.invoice_number = sqlc.arg(invoice_number)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListInvoicesByOrganization :many
SELECT i.*
FROM invoices AS i
JOIN organizations AS o ON o.id = i.organization_id
WHERE i.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY i.created_at DESC;

-- name: UpdateInvoiceStatus :one
UPDATE invoices AS i
SET
    status = sqlc.arg(status),
    paid_at = COALESCE(sqlc.narg(paid_at), i.paid_at),
    updated_at = NOW()
FROM organizations AS o
WHERE i.organization_id = sqlc.arg(organization_id)
  AND i.id = sqlc.arg(id)
  AND o.id = i.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;
