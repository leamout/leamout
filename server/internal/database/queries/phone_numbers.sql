-- name: CreateBYOCPhoneNumber :one
INSERT INTO phone_numbers (
    organization_id,
    number,
    country_code,
    trunk_id,
    voice_enabled
)
SELECT
    sqlc.arg(organization_id) AS organization_id,
    sqlc.arg(number) AS number,
    sqlc.arg(country_code) AS country_code,
    sqlc.narg(trunk_id) AS trunk_id,
    COALESCE(sqlc.narg(voice_enabled), true) AS voice_enabled
FROM organizations AS o
LEFT JOIN trunks AS t
  ON t.id = sqlc.narg(trunk_id)::UUID
 AND t.organization_id = sqlc.arg(organization_id)
 AND t.status = 'active'
 AND t.direction IN ('inbound', 'bidirectional')
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (sqlc.narg(trunk_id)::UUID IS NULL OR t.id IS NOT NULL)
RETURNING *;

-- name: GetPhoneNumberByID :one
SELECT pn.*
FROM phone_numbers AS pn
JOIN organizations AS o ON o.id = pn.organization_id
WHERE pn.id = sqlc.arg(id)
  AND pn.organization_id = sqlc.arg(organization_id)
  AND pn.status <> 'released'
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
SELECT *
FROM phone_numbers
WHERE organization_id = sqlc.arg(organization_id)
  AND status <> 'released'
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
  AND (sqlc.narg(country_code)::text IS NULL OR country_code = sqlc.narg(country_code)::text)
  AND (sqlc.narg(trunk_id)::uuid IS NULL OR trunk_id = sqlc.narg(trunk_id)::uuid)
  AND (sqlc.narg(voice_enabled)::boolean IS NULL OR voice_enabled = sqlc.narg(voice_enabled)::boolean)
ORDER BY created_at DESC, id DESC;

-- name: ListPhoneNumbersByCountry :many
SELECT *
FROM phone_numbers
WHERE organization_id = sqlc.arg(organization_id)
  AND country_code = sqlc.arg(country_code)
  AND status <> 'released'
ORDER BY number ASC;

-- name: UpdatePhoneNumber :one
UPDATE phone_numbers
SET
    voice_enabled = COALESCE(sqlc.narg(voice_enabled), voice_enabled),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status = 'active'
RETURNING *;

-- name: SetBYOCPhoneNumberTrunk :one
UPDATE phone_numbers AS pn
SET
    trunk_id = sqlc.arg(trunk_id),
    updated_at = NOW()
FROM trunks AS t
WHERE pn.id = sqlc.arg(id)
  AND pn.organization_id = sqlc.arg(organization_id)
  AND pn.status = 'active'
  AND t.id = sqlc.arg(trunk_id)
  AND t.organization_id = pn.organization_id
  AND t.status = 'active'
  AND t.direction IN ('inbound', 'bidirectional')
RETURNING pn.*;

-- name: ReleaseBYOCPhoneNumber :one
UPDATE phone_numbers
SET
    status = 'released',
    trunk_id = NULL,
    voice_enabled = false,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
  AND organization_id = sqlc.arg(organization_id)
  AND status IN ('active', 'disabled')
RETURNING *;

-- name: GetVoiceAgentBindingByNumber :one
SELECT
    binding.id AS binding_id,
    binding.voice_agent_id,
    pn.id AS phone_number_id,
    pn.number,
    pn.organization_id
FROM phone_numbers AS pn
JOIN voice_agent_bindings AS binding
  ON binding.phone_number_id = pn.id
 AND binding.organization_id = pn.organization_id
JOIN voice_agents AS agent
  ON agent.id = binding.voice_agent_id
 AND agent.organization_id = binding.organization_id
JOIN organizations AS o ON o.id = pn.organization_id
WHERE pn.number = sqlc.arg(number)
  AND pn.status = 'active'
  AND pn.voice_enabled = true
  AND agent.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- Operator projections include all number and organization lifecycle states.
-- name: ListBackofficePhoneNumbers :many
SELECT
    pn.id::text AS id,
    pn.organization_id::text AS organization_id,
    o.name AS organization_name,
    CASE WHEN o.deleted_at IS NULL THEN o.status ELSE 'deleted' END::text AS organization_status,
    pn.number,
    pn.country_code::text AS country_code,
    COALESCE(pn.trunk_id::text, '')::text AS trunk_id,
    COALESCE(t.name, '')::text AS trunk_name,
    COALESCE(t.status, '')::text AS trunk_status,
    COALESCE(b.voice_agent_id::text, '')::text AS agent_id,
    COALESCE(a.name, '')::text AS agent_name,
    COALESCE(a.status, '')::text AS agent_status,
    pn.voice_enabled,
    pn.status,
    to_char(pn.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')::text AS created_at
FROM phone_numbers AS pn
JOIN organizations AS o ON o.id = pn.organization_id
LEFT JOIN trunks AS t ON t.id = pn.trunk_id AND t.organization_id = pn.organization_id
LEFT JOIN voice_agent_bindings AS b ON b.phone_number_id = pn.id AND b.organization_id = pn.organization_id
LEFT JOIN voice_agents AS a ON a.id = b.voice_agent_id AND a.organization_id = pn.organization_id
WHERE (
    sqlc.arg(search)::text = ''
    OR pn.id::text = sqlc.arg(search)::text
    OR strpos(pn.number, sqlc.arg(search)::text) > 0
    OR strpos(lower(o.name), lower(sqlc.arg(search)::text)) > 0
)
  AND (sqlc.arg(status)::text = '' OR pn.status = sqlc.arg(status)::text)
  AND (sqlc.arg(country)::text = '' OR pn.country_code = sqlc.arg(country)::text)
  AND (sqlc.arg(voice)::text = '' OR pn.voice_enabled = (sqlc.arg(voice)::text = 'yes'))
  AND (sqlc.narg(organization_id)::uuid IS NULL OR pn.organization_id = sqlc.narg(organization_id)::uuid)
  AND (sqlc.narg(trunk_id)::uuid IS NULL OR pn.trunk_id = sqlc.narg(trunk_id)::uuid)
ORDER BY pn.created_at DESC, pn.id DESC
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: GetBackofficePhoneNumber :one
SELECT
    pn.id::text AS id,
    pn.organization_id::text AS organization_id,
    o.name AS organization_name,
    CASE WHEN o.deleted_at IS NULL THEN o.status ELSE 'deleted' END::text AS organization_status,
    pn.number,
    pn.country_code::text AS country_code,
    COALESCE(pn.trunk_id::text, '')::text AS trunk_id,
    COALESCE(t.name, '')::text AS trunk_name,
    COALESCE(t.status, '')::text AS trunk_status,
    COALESCE(b.voice_agent_id::text, '')::text AS agent_id,
    COALESCE(a.name, '')::text AS agent_name,
    COALESCE(a.status, '')::text AS agent_status,
    pn.voice_enabled,
    pn.status,
    to_char(pn.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')::text AS created_at,
    to_char(pn.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS')::text AS updated_at
FROM phone_numbers AS pn
JOIN organizations AS o ON o.id = pn.organization_id
LEFT JOIN trunks AS t ON t.id = pn.trunk_id AND t.organization_id = pn.organization_id
LEFT JOIN voice_agent_bindings AS b ON b.phone_number_id = pn.id AND b.organization_id = pn.organization_id
LEFT JOIN voice_agents AS a ON a.id = b.voice_agent_id AND a.organization_id = pn.organization_id
WHERE pn.id = sqlc.arg(id);
