-- name: CreatePhoneNumber :one
INSERT INTO phone_numbers (
    tenant_id,
    number,
    country_code,
    provider_id,
    voice_enabled,
    sms_enabled,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetPhoneNumber :one
SELECT *
FROM phone_numbers
WHERE id = $1;

-- name: GetPhoneNumberByNumber :one
SELECT *
FROM phone_numbers
WHERE number = $1;

-- name: ListPhoneNumbersByTenant :many
SELECT *
FROM phone_numbers
WHERE tenant_id = $1
ORDER BY number;

-- name: ListPhoneNumbersByTenantAndStatus :many
SELECT *
FROM phone_numbers
WHERE tenant_id = $1
  AND status = $2
ORDER BY number;

-- name: UpdatePhoneNumber :one
UPDATE phone_numbers
SET
    number = $2,
    country_code = $3,
    provider_id = $4,
    voice_enabled = $5,
    sms_enabled = $6,
    status = $7,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeletePhoneNumber :exec
DELETE FROM phone_numbers
WHERE id = $1;
