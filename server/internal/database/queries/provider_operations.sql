-- name: CreateNumberOrderProviderOperation :one
INSERT INTO provider_operations (
    organization_id,
    carrier_provider_id,
    operation_type,
    number_order_id,
    idempotency_key,
    request
)
VALUES (
    sqlc.arg(organization_id),
    sqlc.arg(carrier_provider_id),
    'number_order',
    sqlc.arg(number_order_id),
    sqlc.arg(idempotency_key),
    sqlc.arg(request)
)
ON CONFLICT (organization_id, carrier_provider_id, idempotency_key)
DO UPDATE SET idempotency_key = provider_operations.idempotency_key
WHERE provider_operations.operation_type = EXCLUDED.operation_type
  AND provider_operations.number_order_id = EXCLUDED.number_order_id
  AND provider_operations.request = EXCLUDED.request
RETURNING *;

-- name: CreateNumberReleaseProviderOperation :one
INSERT INTO provider_operations (
    organization_id,
    carrier_provider_id,
    operation_type,
    idempotency_key,
    request,
    phone_number_id,
    provider_resource_id
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(carrier_provider_id),
    'number_release',
    sqlc.arg(idempotency_key),
    sqlc.arg(request),
    pn.id,
    sqlc.arg(provider_resource_id)
FROM phone_numbers AS pn
WHERE pn.id = sqlc.arg(phone_number_id)
  AND pn.organization_id = sqlc.arg(organization_id)
  AND pn.provider_id = sqlc.arg(carrier_provider_id)
  AND pn.provider_resource_id = sqlc.arg(provider_resource_id)
  AND pn.provisioning_mode = 'managed'
  AND pn.status IN ('active', 'disabled')
ON CONFLICT (organization_id, carrier_provider_id, idempotency_key)
DO UPDATE SET idempotency_key = provider_operations.idempotency_key
WHERE provider_operations.operation_type = EXCLUDED.operation_type
  AND provider_operations.phone_number_id = EXCLUDED.phone_number_id
  AND provider_operations.provider_resource_id = EXCLUDED.provider_resource_id
  AND provider_operations.request = EXCLUDED.request
RETURNING *;

-- name: MarkProviderOperationAccepted :one
UPDATE provider_operations
SET
    state = 'provider_accepted',
    provider_operation_id = sqlc.arg(provider_operation_id),
    response = sqlc.arg(response),
    attempts = attempts + 1,
    last_error = NULL,
    next_attempt_at = now()
WHERE id = sqlc.arg(id)
  AND operation_type = 'number_order'
  AND state = 'pending'
RETURNING *;

-- name: RecordProviderOperationAttemptFailure :one
UPDATE provider_operations
SET
    attempts = attempts + 1,
    last_error = sqlc.arg(last_error),
    next_attempt_at = now() + interval '5 minutes'
WHERE id = sqlc.arg(id)
  AND state IN ('pending', 'provider_accepted')
RETURNING *;

-- name: MarkProviderOperationFailed :exec
UPDATE provider_operations
SET
    state = 'failed',
    attempts = attempts + 1,
    last_error = sqlc.arg(last_error),
    next_attempt_at = NULL,
    completed_at = now()
WHERE id = sqlc.arg(id)
  AND state IN ('pending', 'provider_accepted');

-- name: MarkNumberOrderProviderOperationSucceeded :one
UPDATE provider_operations
SET
    state = 'succeeded',
    provider_resource_id = sqlc.arg(provider_resource_id),
    phone_number_id = sqlc.arg(phone_number_id),
    response = sqlc.arg(response),
    completed_at = now(),
    last_error = NULL,
    next_attempt_at = NULL
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND carrier_provider_id = sqlc.arg(carrier_provider_id)
  AND operation_type = 'number_order'
  AND number_order_id = sqlc.arg(number_order_id)
  AND state = 'provider_accepted'
RETURNING *;

-- name: MarkNumberReleaseProviderOperationSucceeded :one
UPDATE provider_operations
SET
    state = 'succeeded',
    completed_at = now(),
    last_error = NULL,
    next_attempt_at = NULL
WHERE id = sqlc.arg(id)
  AND phone_number_id = sqlc.arg(phone_number_id)
  AND operation_type = 'number_release'
  AND state IN ('pending', 'provider_accepted')
RETURNING *;

-- name: ListProviderOperationsReadyForRetry :many
SELECT *
FROM provider_operations
WHERE state IN ('pending', 'provider_accepted')
  AND next_attempt_at IS NOT NULL
  AND next_attempt_at <= now()
ORDER BY next_attempt_at ASC, created_at ASC
LIMIT sqlc.arg(limit_count);

-- name: TryProviderOperationAdvisoryLock :one
SELECT pg_try_advisory_lock(hashtextextended(sqlc.arg(operation_id), 0));

-- name: ReleaseProviderOperationAdvisoryLock :exec
SELECT pg_advisory_unlock(hashtextextended(sqlc.arg(operation_id), 0));
