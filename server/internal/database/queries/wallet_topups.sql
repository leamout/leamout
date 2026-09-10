-- name: LockWalletTopupByReference :one
SELECT
    c.id AS checkout_order_id,
    c.organization_id,
    c.wallet_id,
    c.provider,
    c.reference,
    c.amount_minor,
    c.currency,
    c.status AS checkout_status,
    p.id AS payment_id,
    p.provider_payment_id,
    p.status AS payment_status
FROM checkouts AS c
JOIN payments AS p ON p.checkout_id = c.id
WHERE c.reference = sqlc.arg(reference)
  AND c.checkout_type = 'wallet_topup'
FOR UPDATE OF c, p;

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
