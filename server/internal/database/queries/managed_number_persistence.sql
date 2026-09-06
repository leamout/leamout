-- name: EnsureTransitManagedPhoneNumberForProviderOperation :one
INSERT INTO phone_numbers (
    organization_id,
    number,
    country_code,
    provisioning_mode,
    carrier_connection_id,
    provider_id,
    provider_resource_id,
    voice_enabled,
    sms_enabled,
    status
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(number),
    sqlc.arg(country_code),
    'managed',
    NULL,
    NULL,
    NULL,
    true,
    false,
    'active'
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
ON CONFLICT (number)
    WHERE status <> 'released'
DO UPDATE SET updated_at = phone_numbers.updated_at
WHERE phone_numbers.organization_id = EXCLUDED.organization_id
  AND phone_numbers.provisioning_mode = 'managed'
  AND phone_numbers.provider_id IS NULL
  AND phone_numbers.provider_resource_id IS NULL
  AND phone_numbers.status = 'active'
RETURNING *;
