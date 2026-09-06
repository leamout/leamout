-- name: CreateNumberOrder :one
INSERT INTO number_orders (
    organization_id,
    selection_id,
    number,
    country_code
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(selection_id),
    sqlc.arg(number),
    sqlc.arg(country_code)
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING number_orders.*;

-- name: GetNumberOrderByID :one
SELECT no.*
FROM number_orders AS no
JOIN organizations AS o ON o.id = no.organization_id
WHERE no.id = sqlc.arg(id)
  AND no.organization_id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListNumberOrdersByOrganizationID :many
SELECT no.*
FROM number_orders AS no
WHERE no.organization_id = sqlc.arg(organization_id)
ORDER BY no.created_at DESC;

-- name: LockNumberOrderForProviderOperation :one
SELECT *
FROM number_orders
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND selection_id = sqlc.arg(selection_id)
  AND number = sqlc.arg(number)
  AND country_code = sqlc.arg(country_code)
FOR UPDATE;

-- name: MarkNumberOrderProcessing :one
UPDATE number_orders
SET status = 'processing'
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'pending'
RETURNING *;

-- name: MarkNumberOrderCompleted :one
UPDATE number_orders
SET
    status = 'completed',
    phone_number_id = sqlc.arg(phone_number_id),
    error_code = NULL,
    error_message = NULL
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'processing'
RETURNING *;

-- name: MarkNumberOrderFailed :exec
UPDATE number_orders
SET
    status = 'failed',
    error_code = 'provider_order_failed',
    error_message = sqlc.arg(error_message)
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status IN ('pending', 'processing');
