-- name: CreateWallet :one
INSERT INTO wallets (organization_id, currency)
SELECT sqlc.arg(organization_id), sqlc.arg(currency)
FROM organizations
WHERE id = sqlc.arg(organization_id)
  AND status = 'active'
  AND deleted_at IS NULL
RETURNING *;

-- name: GetWallet :one
SELECT w.*
FROM wallets AS w
JOIN organizations AS o ON o.id = w.organization_id
WHERE w.organization_id = sqlc.arg(organization_id)
  AND w.id = sqlc.arg(id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListWallets :many
SELECT w.*
FROM wallets AS w
JOIN organizations AS o ON o.id = w.organization_id
WHERE w.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ORDER BY w.created_at ASC, w.id ASC;

-- name: GetWalletByCurrency :one
SELECT w.*
FROM wallets AS w
JOIN organizations AS o ON o.id = w.organization_id
WHERE w.organization_id = sqlc.arg(organization_id)
  AND w.currency = sqlc.arg(currency)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: GetWalletBalance :one
SELECT
    COALESCE((SELECT SUM(amount_minor) FROM wallet_ledger_entries WHERE wallet_id = w.id), 0)::BIGINT AS posted_minor,
    COALESCE((SELECT SUM(amount_minor) FROM wallet_reservations WHERE wallet_id = w.id AND status = 'active'), 0)::BIGINT AS reserved_minor,
    (
        COALESCE((SELECT SUM(amount_minor) FROM wallet_ledger_entries WHERE wallet_id = w.id), 0)
        - COALESCE((SELECT SUM(amount_minor) FROM wallet_reservations WHERE wallet_id = w.id AND status = 'active'), 0)
    )::BIGINT AS available_minor
FROM wallets AS w
JOIN organizations AS o ON o.id = w.organization_id
WHERE w.organization_id = sqlc.arg(organization_id)
  AND w.id = sqlc.arg(wallet_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: CreateWalletLedgerEntry :one
INSERT INTO wallet_ledger_entries (
    wallet_id, organization_id, entry_type, amount_minor, source_type,
    source_id, idempotency_key, metadata, occurred_at
)
SELECT
    w.id AS wallet_id,
    w.organization_id,
    sqlc.arg(entry_type) AS entry_type,
    sqlc.arg(amount_minor) AS amount_minor,
    sqlc.arg(source_type) AS source_type,
    sqlc.arg(source_id) AS source_id,
    sqlc.arg(idempotency_key) AS idempotency_key,
    COALESCE(sqlc.narg(metadata), '{}'::jsonb) AS metadata,
    COALESCE(sqlc.narg(occurred_at), NOW()) AS occurred_at
FROM wallets AS w
JOIN organizations AS o ON o.id = w.organization_id
WHERE w.id = sqlc.arg(wallet_id)
  AND w.organization_id = sqlc.arg(organization_id)
  AND w.status != 'closed'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: ListWalletLedgerEntries :many
SELECT *
FROM wallet_ledger_entries
WHERE wallet_id = sqlc.arg(wallet_id)
  AND organization_id = sqlc.arg(organization_id)
ORDER BY occurred_at DESC, id DESC;

-- name: LockActiveWallet :one
SELECT w.*
FROM wallets AS w
JOIN organizations AS o ON o.id = w.organization_id
WHERE w.id = sqlc.arg(id)
  AND w.organization_id = sqlc.arg(organization_id)
  AND w.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
FOR UPDATE OF w;

-- name: InsertWalletReservation :one
INSERT INTO wallet_reservations (
    wallet_id, organization_id, amount_minor, operation_type, operation_id, expires_at
)
VALUES (
    sqlc.arg(wallet_id), sqlc.arg(organization_id), sqlc.arg(amount_minor),
    sqlc.arg(operation_type), sqlc.arg(operation_id), sqlc.arg(expires_at)
)
ON CONFLICT (wallet_id, operation_type, operation_id) DO NOTHING
RETURNING *;

-- name: GetWalletReservationByOperation :one
SELECT *
FROM wallet_reservations
WHERE wallet_id = sqlc.arg(wallet_id)
  AND organization_id = sqlc.arg(organization_id)
  AND operation_type = sqlc.arg(operation_type)
  AND operation_id = sqlc.arg(operation_id)
LIMIT 1;

-- name: GetWalletReservation :one
SELECT *
FROM wallet_reservations
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: CaptureWalletReservation :one
UPDATE wallet_reservations AS wr
SET status = 'captured',
    captured_amount_minor = sqlc.arg(captured_amount_minor),
    captured_at = NOW(),
    updated_at = NOW()
WHERE wr.organization_id = sqlc.arg(organization_id)
  AND wr.id = sqlc.arg(id)
  AND wr.status = 'active'
  AND wr.expires_at > NOW()
  AND sqlc.arg(captured_amount_minor) <= wr.amount_minor
RETURNING wr.*;

-- name: ReleaseWalletReservation :one
UPDATE wallet_reservations
SET status = 'released', released_at = NOW(), updated_at = NOW()
WHERE organization_id = sqlc.arg(organization_id)
  AND id = sqlc.arg(id)
  AND status = 'active'
RETURNING *;

-- name: ExpireWalletReservations :many
UPDATE wallet_reservations
SET status = 'expired', expired_at = NOW(), updated_at = NOW()
WHERE status = 'active'
  AND expires_at <= NOW()
RETURNING *;
