-- name: CreateInvoiceItem :one
INSERT INTO invoice_items (
    invoice_id,
    meter_id,
    usage_rate_id,
    type,
    description,
    quantity,
    unit_amount_micros,
    amount,
    period_start,
    period_end,
    metadata
)
SELECT
    i.id AS invoice_id,
    sqlc.narg(meter_id)::uuid AS meter_id,
    sqlc.narg(usage_rate_id)::uuid AS usage_rate_id,
    sqlc.arg(type) AS type,
    sqlc.arg(description) AS description,
    COALESCE(sqlc.narg(quantity), 1) AS quantity,
    sqlc.narg(unit_amount_micros) AS unit_amount_micros,
    sqlc.arg(amount) AS amount,
    sqlc.narg(period_start) AS period_start,
    sqlc.narg(period_end) AS period_end,
    COALESCE(sqlc.narg(metadata), '{}'::jsonb) AS metadata
FROM invoices AS i
JOIN organizations AS o ON o.id = i.organization_id
WHERE i.id = sqlc.arg(invoice_id)
  AND i.organization_id = sqlc.arg(organization_id)
  AND i.status = 'draft'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (
      sqlc.narg(meter_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM meters AS m
          WHERE m.id = sqlc.narg(meter_id)::uuid
      )
  )
  AND (
      sqlc.narg(usage_rate_id)::uuid IS NULL
      OR EXISTS (
          SELECT 1
          FROM usage_rates AS ur
          LEFT JOIN subscriptions AS s
            ON s.id = i.subscription_id
           AND s.organization_id = i.organization_id
          WHERE ur.id = sqlc.narg(usage_rate_id)::uuid
            AND (
                sqlc.narg(meter_id)::uuid IS NULL
                OR ur.meter_id = sqlc.narg(meter_id)::uuid
            )
            AND (
                i.subscription_id IS NULL
                OR ur.plan_id = s.plan_id
            )
      )
  )
RETURNING *;

-- name: ListInvoiceItems :many
SELECT ii.*
FROM invoice_items AS ii
JOIN invoices AS i ON i.id = ii.invoice_id
JOIN organizations AS o ON o.id = i.organization_id
WHERE ii.invoice_id = sqlc.arg(invoice_id)
  AND i.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY ii.created_at ASC;

-- name: DeleteInvoiceItems :exec
DELETE FROM invoice_items AS ii
USING invoices AS i, organizations AS o
WHERE ii.invoice_id = sqlc.arg(invoice_id)
  AND i.id = ii.invoice_id
  AND i.organization_id = sqlc.arg(organization_id)
  AND i.status = 'draft'
  AND o.id = i.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL;
