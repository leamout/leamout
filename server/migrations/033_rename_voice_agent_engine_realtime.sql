ALTER TABLE voice_agents
    DROP CONSTRAINT chk_voice_agents_engine;

UPDATE voice_agents
SET engine = 'realtime'
WHERE engine = 'integrated';

UPDATE voice_agent_sessions
SET configuration_snapshot = jsonb_set(
    configuration_snapshot,
    '{engine}',
    '"realtime"'::jsonb,
    false
)
WHERE configuration_snapshot ->> 'engine' = 'integrated';

ALTER TABLE voice_agents
    ADD CONSTRAINT chk_voice_agents_engine
    CHECK (engine IN ('composable', 'realtime'));
