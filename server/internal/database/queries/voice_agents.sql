-- name: CreateVoiceAgent :one
INSERT INTO voice_agents (
    organization_id,
    name,
    engine,
    instructions,
    voice,
    language,
    engine_config,
    interruption_policy,
    recording_policy
)
SELECT
    sqlc.arg(organization_id),
    sqlc.arg(name),
    sqlc.arg(engine),
    sqlc.arg(instructions),
    sqlc.narg(voice),
    sqlc.narg(language),
    COALESCE(sqlc.narg(engine_config)::jsonb, '{}'::jsonb),
    sqlc.arg(interruption_policy),
    sqlc.arg(recording_policy)
FROM organizations AS o
WHERE o.id = sqlc.arg(organization_id)
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: GetVoiceAgentByID :one
SELECT va.*
FROM voice_agents AS va
JOIN organizations AS o ON o.id = va.organization_id
WHERE va.id = sqlc.arg(id)
  AND va.organization_id = sqlc.arg(organization_id)
  AND va.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: ListVoiceAgentsByOrganizationID :many
SELECT va.*
FROM voice_agents AS va
JOIN organizations AS o ON o.id = va.organization_id
WHERE va.organization_id = sqlc.arg(organization_id)
  AND va.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
  AND (sqlc.narg(engine)::text IS NULL OR va.engine = sqlc.narg(engine)::text)
  AND (sqlc.narg(language)::text IS NULL OR va.language = sqlc.narg(language)::text)
ORDER BY va.created_at DESC, va.id DESC;

-- name: UpdateVoiceAgent :one
UPDATE voice_agents AS va
SET
    name = COALESCE(sqlc.narg(name), va.name),
    engine = COALESCE(sqlc.narg(engine), va.engine),
    instructions = COALESCE(sqlc.narg(instructions), va.instructions),
    voice = COALESCE(sqlc.narg(voice), va.voice),
    language = COALESCE(sqlc.narg(language), va.language),
    engine_config = COALESCE(sqlc.narg(engine_config)::jsonb, va.engine_config),
    interruption_policy = COALESCE(sqlc.narg(interruption_policy), va.interruption_policy),
    recording_policy = COALESCE(sqlc.narg(recording_policy), va.recording_policy),
    configuration_revision = va.configuration_revision + 1,
    updated_at = NOW()
FROM organizations AS o
WHERE va.id = sqlc.arg(id)
  AND va.organization_id = sqlc.arg(organization_id)
  AND va.status = 'active'
  AND o.id = va.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING va.*;

-- name: DisableVoiceAgent :exec
UPDATE voice_agents AS va
SET
    status = 'disabled',
    updated_at = NOW()
FROM organizations AS o
WHERE va.id = sqlc.arg(id)
  AND va.organization_id = sqlc.arg(organization_id)
  AND va.status = 'active'
  AND o.id = va.organization_id
  AND o.status = 'active'
  AND o.deleted_at IS NULL;

-- name: CountOrganizationVoiceAgents :one
SELECT COUNT(*)::bigint AS count FROM voice_agents
WHERE organization_id = sqlc.arg(organization_id);

-- Backoffice projections deliberately include disabled agents and organizations.
-- name: ListBackofficeAgents :many
SELECT
    va.id,
    va.name,
    va.organization_id,
    o.name AS organization_name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::text AS organization_status,
    va.engine,
    va.status,
    va.language,
    va.created_at
FROM voice_agents AS va
JOIN organizations AS o ON o.id = va.organization_id
WHERE (
    sqlc.arg(search)::text = ''
    OR strpos(lower(va.name), lower(sqlc.arg(search)::text)) > 0
    OR va.id::text = sqlc.arg(search)::text
    OR strpos(lower(o.name), lower(sqlc.arg(search)::text)) > 0
)
  AND (sqlc.arg(status)::text = '' OR va.status = sqlc.arg(status)::text)
  AND (sqlc.arg(engine)::text = '' OR va.engine = sqlc.arg(engine)::text)
  AND (sqlc.narg(organization_id)::uuid IS NULL OR va.organization_id = sqlc.narg(organization_id)::uuid)
ORDER BY va.created_at DESC, va.id DESC
LIMIT sqlc.arg(page_limit)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: GetBackofficeAgent :one
SELECT
    va.id,
    va.name,
    va.organization_id,
    o.name AS organization_name,
    CASE WHEN o.deleted_at IS NOT NULL THEN 'deleted' ELSE o.status END::text AS organization_status,
    va.engine,
    va.status,
    va.language,
    va.created_at,
    va.updated_at,
    va.instructions,
    COALESCE(va.voice, '')::text AS voice,
    va.interruption_policy,
    va.recording_policy,
    va.configuration_revision
FROM voice_agents AS va
JOIN organizations AS o ON o.id = va.organization_id
WHERE va.id = sqlc.arg(id);
