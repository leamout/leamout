-- name: InsertProviderCDR :one
INSERT INTO provider_cdrs (
    carrier_provider_id,
    carrier_connection_id,
    provider_record_id,
    direction,
    sip_call_id,
    started_at,
    duration_seconds,
    currency,
    cost_micros,
    raw
)
VALUES (
    sqlc.arg(carrier_provider_id),
    sqlc.arg(carrier_connection_id),
    sqlc.arg(provider_record_id),
    sqlc.arg(direction),
    sqlc.arg(sip_call_id),
    sqlc.arg(started_at),
    sqlc.arg(duration_seconds),
    sqlc.arg(currency),
    sqlc.arg(cost_micros),
    sqlc.arg(raw)
)
ON CONFLICT (carrier_provider_id, direction, provider_record_id)
DO NOTHING
RETURNING *;

-- name: GetProviderCDRForUpdate :one
SELECT *
FROM provider_cdrs
WHERE carrier_provider_id = sqlc.arg(carrier_provider_id)
  AND direction = sqlc.arg(direction)
  AND provider_record_id = sqlc.arg(provider_record_id)
FOR UPDATE;

-- name: FindManagedCallForProviderCDR :one
SELECT
    c.id,
    c.organization_id
FROM calls AS c
JOIN carrier_connections AS cc
  ON cc.id = c.carrier_connection_id
WHERE c.sip_call_id = sqlc.arg(sip_call_id)
  AND c.direction = 'outbound'
  AND c.carrier_connection_id = sqlc.arg(carrier_connection_id)
  AND cc.scope = 'platform'
  AND cc.provider_id = sqlc.arg(carrier_provider_id)
LIMIT 1;

-- name: MarkProviderCDRReconciled :one
UPDATE provider_cdrs
SET
    call_id = sqlc.arg(call_id),
    organization_id = sqlc.arg(organization_id),
    reconciled_at = now()
WHERE id = sqlc.arg(id)
  AND reconciled_at IS NULL
RETURNING *;

-- name: CreateWholesaleCharge :one
INSERT INTO wholesale_charges (
    provider_cdr_id,
    organization_id,
    call_id,
    amount_micros,
    currency,
    occurred_at
)
VALUES (
    sqlc.arg(provider_cdr_id),
    sqlc.arg(organization_id),
    sqlc.arg(call_id),
    sqlc.arg(amount_micros),
    sqlc.arg(currency),
    sqlc.arg(occurred_at)
)
RETURNING *;

-- name: SumWholesaleChargesForDay :one
SELECT COALESCE(SUM(amount_micros), 0)::BIGINT
FROM wholesale_charges
WHERE organization_id = sqlc.arg(organization_id)
  AND occurred_at >= sqlc.arg(day)::TIMESTAMPTZ
  AND occurred_at < sqlc.arg(day)::TIMESTAMPTZ + interval '1 day';
