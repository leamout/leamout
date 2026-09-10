-- name: LockWalletTopupByReference :one
SELECT
    co.id AS checkout_order_id,
    co.organization_id,
    co.wallet_id,
    co.provider,
    co.reference,
    co.amount_minor,
    co.currency,
    co.status AS checkout_status,
    p.id AS payment_id,
    p.provider_payment_id,
    p.status AS payment_status
FROM checkout_orders AS co
JOIN payments AS p ON p.checkout_order_id = co.id
WHERE co.reference = sqlc.arg(reference)
  AND co.order_type = 'wallet_topup'
FOR UPDATE OF co, p;

-- name: InsertPaymentProviderEvent :one
INSERT INTO payment_provider_events (
    payment_id, organization_id, provider, provider_event_id,
    event_type, payload_sha256, payload
)
VALUES (
    sqlc.arg(payment_id), sqlc.arg(organization_id), sqlc.arg(provider),
    sqlc.arg(provider_event_id), sqlc.arg(event_type),
    sqlc.arg(payload_sha256), sqlc.arg(payload)
)
ON CONFLICT (provider, provider_event_id) DO NOTHING
RETURNING *;

-- name: MarkPaymentProviderEventProcessed :exec
UPDATE payment_provider_events
SET processed_at = NOW()
WHERE id = sqlc.arg(id)
  AND processed_at IS NULL;
