-- name: CreateVoiceApplication :one
INSERT INTO voice_applications (
    tenant_id,
    name,
    ring_timeout_seconds,
    caller_id,
    status
) VALUES (
    sqlc.arg(tenant_id),
    sqlc.arg(name),
    COALESCE(sqlc.narg(ring_timeout_seconds), 30),
    sqlc.narg(caller_id),
    COALESCE(sqlc.narg(status), 'active')
)
RETURNING *;

-- name: GetVoiceApplication :one
SELECT *
FROM voice_applications
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
LIMIT 1;

-- name: GetVoiceApplicationByName :one
SELECT *
FROM voice_applications
WHERE tenant_id = sqlc.arg(tenant_id)
  AND name = sqlc.arg(name)
LIMIT 1;

-- name: ListVoiceApplications :many
SELECT *
FROM voice_applications
WHERE tenant_id = sqlc.arg(tenant_id)
ORDER BY created_at DESC;

-- name: UpdateVoiceApplication :one
UPDATE voice_applications
SET
    name = COALESCE(sqlc.narg(name), name),
    ring_timeout_seconds = COALESCE(sqlc.narg(ring_timeout_seconds), ring_timeout_seconds),
    caller_id = COALESCE(sqlc.narg(caller_id), caller_id),
    status = COALESCE(sqlc.narg(status), status),
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id)
RETURNING *;

-- name: DisableVoiceApplication :exec
UPDATE voice_applications
SET
    status = 'disabled',
    updated_at = NOW()
WHERE tenant_id = sqlc.arg(tenant_id)
  AND id = sqlc.arg(id);

-- name: CreateVoiceBinding :one
INSERT INTO voice_bindings (
    voice_application_id,
    phone_number_id,
    sip_domain_id,
    subscriber_id
) VALUES (
    sqlc.arg(voice_application_id),
    sqlc.narg(phone_number_id),
    sqlc.narg(sip_domain_id),
    sqlc.narg(subscriber_id)
)
RETURNING *;

-- name: GetVoiceBinding :one
SELECT vb.*
FROM voice_bindings AS vb
JOIN voice_applications AS va ON va.id = vb.voice_application_id
WHERE va.tenant_id = sqlc.arg(tenant_id)
  AND vb.id = sqlc.arg(id)
LIMIT 1;

-- name: ListVoiceBindings :many
SELECT vb.*
FROM voice_bindings AS vb
JOIN voice_applications AS va ON va.id = vb.voice_application_id
WHERE va.tenant_id = sqlc.arg(tenant_id)
  AND vb.voice_application_id = sqlc.arg(voice_application_id)
ORDER BY vb.created_at ASC;

-- name: DeleteVoiceBinding :exec
DELETE FROM voice_bindings AS vb
USING voice_applications AS va
WHERE vb.id = sqlc.arg(id)
  AND vb.voice_application_id = va.id
  AND va.tenant_id = sqlc.arg(tenant_id);
