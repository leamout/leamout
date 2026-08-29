-- name: CreateInvoiceItem :one
INSERT INTO invoice_items (
    invoice_id,
    meter_id,
    carrier_rate_id,
    type,
    description,
    quantity,
    unit_amount_micros,
    amount,
    period_start,
    period_end,
    metadata
) VALUES (
    sqlc.arg(invoice_id),
    sqlc.narg(meter_id),
    sqlc.narg(carrier_rate_id),
    sqlc.arg(type),
    sqlc.arg(description),
    COALESCE(sqlc.narg(quantity), 1),
    sqlc.narg(unit_amount_micros),
    sqlc.arg(amount),
    sqlc.narg(period_start),
    sqlc.narg(period_end),
    COALESCE(sqlc.narg(metadata), '{}'::jsonb)
)
RETURNING *;

-- name: ListInvoiceItems :many
SELECT *
FROM invoice_items
WHERE invoice_id = sqlc.arg(invoice_id)
ORDER BY created_at ASC;

-- name: DeleteInvoiceItems :exec
DELETE FROM invoice_items
WHERE invoice_id = sqlc.arg(invoice_id);
