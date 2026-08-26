-- name: CreatePhoneNumber :one
INSERT INTO phone_numbers (
    tenant_id,
    number,
    country_code,
    voice_enabled,
    sms_enabled
)
SELECT
    sqlc.arg(tenant_id) as tenant_id,
    sqlc.arg(number) as number,
    sqlc.arg(country_code) as country_code,
    COALESCE(sqlc.narg(voice_enabled), true) as voice_enabled,
    COALESCE(sqlc.narg(sms_enabled), false) as sms_enabled
FROM tenants AS t
WHERE t.id = sqlc.arg(tenant_id)
  AND t.status = 'active'
  AND t.deleted_at IS NULL
RETURNING *;

-- name: GetPhoneNumberByID :one
SELECT pn.*
FROM phone_numbers AS pn
JOIN tenants AS t ON t.id = pn.tenant_id
WHERE pn.id = sqlc.arg(id)
  AND pn.tenant_id = sqlc.arg(tenant_id)
  AND pn.status = 'active'
  AND t.status = 'active'
  AND t.deleted_at IS NULL
LIMIT 1;

-- name: GetPhoneNumberByNumber :one
SELECT pn.*
FROM phone_numbers AS pn
JOIN tenants AS t ON t.id = pn.tenant_id
WHERE pn.number = sqlc.arg(number)
  AND pn.tenant_id = sqlc.arg(tenant_id)
  AND pn.status = 'active'
  AND t.status = 'active'
  AND t.deleted_at IS NULL
LIMIT 1;


-- name: ListPhoneNumbersByTenantID :many
SELECT pn.*
FROM phone_numbers AS pn
WHERE pn.tenant_id = sqlc.arg(tenant_id)
  AND pn.status = 'active'
ORDER BY pn.created_at DESC;

-- name: ListPhoneNumbersByCountry :many
SELECT pn.*
FROM phone_numbers AS pn
WHERE pn.tenant_id = sqlc.arg(tenant_id)
  AND pn.country_code = sqlc.arg(country_code)
  AND pn.status = 'active'
ORDER BY pn.number ASC;

-- name: UpdatePhoneNumber :one
UPDATE phone_numbers
SET
    country_code = COALESCE(sqlc.narg(country_code), country_code),
    voice_enabled = COALESCE(sqlc.narg(voice_enabled), voice_enabled),
    sms_enabled = COALESCE(sqlc.narg(sms_enabled), sms_enabled),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND tenant_id = sqlc.arg(tenant_id)
  AND status = 'active'
RETURNING *;

-- name: DisablePhoneNumber :exec
UPDATE phone_numbers
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND tenant_id = sqlc.arg(tenant_id)
  AND status = 'active';

-- name: EnablePhoneNumber :exec
UPDATE phone_numbers
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND tenant_id = sqlc.arg(tenant_id)
  AND status = 'disabled';

-- name: ReleasePhoneNumber :exec
UPDATE phone_numbers
SET
    status = 'released',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND tenant_id = sqlc.arg(tenant_id)
  AND status IN ('active', 'disabled');

-- name: GetVoiceBindingByNumber :one
SELECT
    vb.id AS binding_id,
    vb.voice_application_id,
    va.name AS application_name,
    va.ring_timeout_seconds,
    va.caller_id AS application_caller_id,
    pn.id AS phone_number_id,
    pn.number,
    pn.tenant_id
FROM phone_numbers AS pn
JOIN voice_bindings AS vb ON vb.phone_number_id = pn.id
JOIN voice_applications AS va ON va.id = vb.voice_application_id
JOIN tenants AS t ON t.id = pn.tenant_id
WHERE pn.number = sqlc.arg(number)
  AND pn.status = 'active'
  AND pn.voice_enabled = true
  AND va.status = 'active'
  AND t.status = 'active'
  AND t.deleted_at IS NULL
LIMIT 1;
