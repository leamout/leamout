-- name: CreatePhoneNumber :one
INSERT INTO phone_numbers (
    organization_id,
    number,
    country_code,
    carrier_connection_id,
    provider_resource_id,
    voice_enabled,
    sms_enabled
)
SELECT
    sqlc.arg(organization_id) AS organization_id,
    sqlc.arg(number) AS number,
    sqlc.arg(country_code) AS country_code,
    sqlc.narg(carrier_connection_id) AS carrier_connection_id,
    sqlc.narg(provider_resource_id) AS provider_resource_id,
    COALESCE(sqlc.narg(voice_enabled), true) AS voice_enabled,
    COALESCE(sqlc.narg(sms_enabled), false) AS sms_enabled
FROM organizations AS o
LEFT JOIN carrier_connections AS cc
  ON cc.id = sqlc.narg(carrier_connection_id)::UUID
 AND cc.organization_id = sqlc.arg(organization_id)
 AND cc.status = 'active'
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (
      sqlc.narg(carrier_connection_id)::UUID IS NULL
      OR cc.id IS NOT NULL
  )
RETURNING *;

-- name: GetPhoneNumberByID :one
SELECT pn.*
FROM phone_numbers AS pn
JOIN organizations AS o ON o.id = pn.organization_id
WHERE pn.id = sqlc.arg(id)
  AND pn.organization_id = sqlc.arg(organization_id)
  AND pn.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: GetPhoneNumberByNumber :one
SELECT pn.*
FROM phone_numbers AS pn
JOIN organizations AS o ON o.id = pn.organization_id
WHERE pn.number = sqlc.arg(number)
  AND pn.organization_id = sqlc.arg(organization_id)
  AND pn.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListPhoneNumbersByOrganizationID :many
SELECT pn.*
FROM phone_numbers AS pn
WHERE pn.organization_id = sqlc.arg(organization_id)
  AND pn.status = 'active'
ORDER BY pn.created_at DESC;

-- name: ListPhoneNumbersByCountry :many
SELECT pn.*
FROM phone_numbers AS pn
WHERE pn.organization_id = sqlc.arg(organization_id)
  AND pn.country_code = sqlc.arg(country_code)
  AND pn.status = 'active'
ORDER BY pn.number ASC;

-- name: UpdatePhoneNumber :one
UPDATE phone_numbers AS pn
SET
    country_code = COALESCE(sqlc.narg(country_code), pn.country_code),
    carrier_connection_id = COALESCE(sqlc.narg(carrier_connection_id), pn.carrier_connection_id),
    provider_resource_id = COALESCE(sqlc.narg(provider_resource_id), pn.provider_resource_id),
    voice_enabled = COALESCE(sqlc.narg(voice_enabled), pn.voice_enabled),
    sms_enabled = COALESCE(sqlc.narg(sms_enabled), pn.sms_enabled),
    updated_at = NOW()
WHERE pn.id = sqlc.arg(id)
  AND pn.organization_id = sqlc.arg(organization_id)
  AND pn.status = 'active'
  AND (
      sqlc.narg(carrier_connection_id)::UUID IS NULL
      OR EXISTS (
          SELECT 1
          FROM carrier_connections AS cc
          WHERE cc.id = sqlc.narg(carrier_connection_id)::UUID
            AND cc.organization_id = pn.organization_id
            AND cc.status = 'active'
      )
  )
RETURNING pn.*;

-- name: DisablePhoneNumber :exec
UPDATE phone_numbers
SET
    status = 'disabled',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'active';

-- name: EnablePhoneNumber :exec
UPDATE phone_numbers
SET
    status = 'active',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'disabled';

-- name: ReleasePhoneNumber :exec
UPDATE phone_numbers
SET
    status = 'released',
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
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
    pn.organization_id
FROM phone_numbers AS pn
JOIN voice_bindings AS vb ON vb.phone_number_id = pn.id
JOIN voice_applications AS va ON va.id = vb.voice_application_id
JOIN organizations AS o ON o.id = pn.organization_id
WHERE pn.number = sqlc.arg(number)
  AND pn.status = 'active'
  AND pn.voice_enabled = true
  AND va.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;
