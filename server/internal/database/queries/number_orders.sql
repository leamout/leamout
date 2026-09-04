-- name: CreateNumberOrder :one
INSERT INTO number_orders (
    organization_id,
    provider_id,
    provider_inventory_id,
    provider_product_id,
    number,
    country_code
)
SELECT
    sqlc.arg(organization_id),
    cp.id,
    sqlc.arg(provider_inventory_id),
    sqlc.arg(provider_product_id),
    sqlc.arg(number),
    sqlc.arg(country_code)
FROM organizations AS o
JOIN carrier_providers AS cp
  ON cp.id = sqlc.arg(provider_id)
 AND cp.status = 'active'
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

-- name: MarkNumberOrderPurchasing :one
UPDATE number_orders
SET
    status = 'purchasing',
    failed_stage = NULL,
    error_code = NULL,
    error_message = NULL
WHERE id = sqlc.arg(id)
  AND (
      status = 'pending'
      OR (status = 'failed' AND failed_stage = 'purchasing')
  )
RETURNING *;

-- name: MarkNumberOrderPurchased :one
UPDATE number_orders
SET
    status = 'purchased',
    provider_order_id = sqlc.arg(provider_order_id),
    failed_stage = NULL,
    error_code = NULL,
    error_message = NULL
WHERE id = sqlc.arg(id)
  AND status = 'purchasing'
RETURNING *;

-- name: MarkNumberOrderPersisting :one
UPDATE number_orders
SET
    status = 'persisting',
    provider_resource_id = COALESCE(sqlc.narg(provider_resource_id), provider_resource_id),
    failed_stage = NULL,
    error_code = NULL,
    error_message = NULL
WHERE id = sqlc.arg(id)
  AND (
      status = 'purchased'
      OR (status = 'failed' AND failed_stage = 'persisting')
  )
RETURNING *;

-- name: MarkNumberOrderConfiguring :one
UPDATE number_orders
SET
    status = 'configuring',
    phone_number_id = COALESCE(sqlc.narg(phone_number_id), phone_number_id),
    failed_stage = NULL,
    error_code = NULL,
    error_message = NULL
WHERE id = sqlc.arg(id)
  AND (
      status = 'persisting'
      OR (status = 'failed' AND failed_stage = 'configuring')
  )
RETURNING *;

-- name: MarkNumberOrderCompleted :one
UPDATE number_orders
SET
    status = 'completed',
    failed_stage = NULL,
    error_code = NULL,
    error_message = NULL
WHERE id = sqlc.arg(id)
  AND status = 'configuring'
  AND provider_order_id IS NOT NULL
  AND provider_resource_id IS NOT NULL
  AND phone_number_id IS NOT NULL
RETURNING *;

-- name: MarkNumberOrderFailed :one
UPDATE number_orders
SET
    status = 'failed',
    failed_stage = sqlc.arg(failed_stage),
    error_code = sqlc.narg(error_code),
    error_message = sqlc.arg(error_message)
WHERE id = sqlc.arg(id)
  AND status = sqlc.arg(expected_status)
  AND status IN ('purchasing', 'persisting', 'configuring')
RETURNING *;

-- name: ListRecoverableNumberOrders :many
SELECT *
FROM number_orders
WHERE status IN ('pending', 'purchasing', 'purchased', 'persisting', 'configuring', 'failed')
ORDER BY updated_at ASC
LIMIT sqlc.arg(limit_count);
