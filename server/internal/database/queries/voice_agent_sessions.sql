-- name: CreateVoiceAgentSession :one
INSERT INTO voice_agent_sessions (
    organization_id,
    call_id,
    voice_agent_id,
    configuration_revision,
    configuration_snapshot
)
SELECT
    sqlc.arg(organization_id),
    c.id,
    agent.id,
    agent.configuration_revision,
    jsonb_build_object(
        'engine', agent.engine,
        'instructions', agent.instructions,
        'voice', agent.voice,
        'language', agent.language,
        'engine_config', agent.engine_config,
        'interruption_policy', agent.interruption_policy,
        'recording_policy', agent.recording_policy,
        'providers', COALESCE((
            SELECT jsonb_agg(
                jsonb_build_object(
                    'role', binding.role,
                    'provider', binding.provider,
                    'credential_id', binding.credential_id,
                    'config', binding.config
                )
                ORDER BY binding.role
            )
            FROM voice_agent_provider_bindings AS binding
            WHERE binding.organization_id = agent.organization_id
              AND binding.voice_agent_id = agent.id
        ), '[]'::jsonb),
        'tools', COALESCE((
            SELECT jsonb_agg(
                jsonb_build_object(
                    'id', tool.id,
                    'name', tool.name,
                    'description', tool.description,
                    'parameters', tool.parameters
                )
                ORDER BY tool.name
            )
            FROM voice_agent_tools AS tool
            WHERE tool.organization_id = agent.organization_id
              AND tool.voice_agent_id = agent.id
              AND tool.enabled = TRUE
        ), '[]'::jsonb)
    )
FROM calls AS c
JOIN organizations AS o
  ON o.id = c.organization_id
JOIN voice_agents AS agent
  ON agent.id = c.voice_agent_id
 AND agent.organization_id = c.organization_id
WHERE c.id = sqlc.arg(call_id)
  AND c.organization_id = sqlc.arg(organization_id)
  AND agent.id = sqlc.arg(voice_agent_id)
  AND c.state IN ('answered', 'active')
  AND c.ended_at IS NULL
  AND agent.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
RETURNING *;

-- name: GetVoiceAgentSessionByID :one
SELECT session.*
FROM voice_agent_sessions AS session
WHERE session.id = sqlc.arg(id)
  AND session.organization_id = sqlc.arg(organization_id)
LIMIT 1;

-- name: GetActiveVoiceAgentSessionByCallID :one
SELECT session.*
FROM voice_agent_sessions AS session
JOIN organizations AS o ON o.id = session.organization_id
WHERE session.organization_id = sqlc.arg(organization_id)
  AND session.call_id = sqlc.arg(call_id)
  AND session.state = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL
LIMIT 1;

-- name: CompleteVoiceAgentSession :one
UPDATE voice_agent_sessions AS session
SET
    state = sqlc.arg(state),
    turn_count = sqlc.arg(turn_count),
    interruption_count = sqlc.arg(interruption_count),
    first_response_latency_ms = sqlc.narg(first_response_latency_ms),
    avg_turn_latency_ms = sqlc.narg(avg_turn_latency_ms),
    ended_at = sqlc.arg(ended_at),
    updated_at = NOW()
WHERE session.id = sqlc.arg(id)
  AND session.organization_id = sqlc.arg(organization_id)
  AND session.state = 'active'
RETURNING session.*;
