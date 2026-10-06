CREATE TABLE IF NOT EXISTS voice_agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    name TEXT NOT NULL,

    engine TEXT NOT NULL DEFAULT 'composable',
    instructions TEXT NOT NULL,
    voice TEXT,
    language TEXT NOT NULL DEFAULT 'en',
    engine_config JSONB NOT NULL DEFAULT '{}'::jsonb,

    interruption_policy TEXT NOT NULL DEFAULT 'allow',
    recording_policy TEXT NOT NULL DEFAULT 'none',
    configuration_revision INTEGER NOT NULL DEFAULT 1,

    status TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_voice_agents_id_organization UNIQUE (id, organization_id),
    CONSTRAINT chk_voice_agents_name
        CHECK (length(btrim(name)) BETWEEN 1 AND 128),
    CONSTRAINT chk_voice_agents_engine
        CHECK (engine IN ('composable', 'integrated')),
    CONSTRAINT chk_voice_agents_instructions
        CHECK (length(btrim(instructions)) BETWEEN 1 AND 20000),
    CONSTRAINT chk_voice_agents_voice
        CHECK (voice IS NULL OR length(btrim(voice)) BETWEEN 1 AND 255),
    CONSTRAINT chk_voice_agents_language
        CHECK (length(btrim(language)) BETWEEN 1 AND 64),
    CONSTRAINT chk_voice_agents_engine_config
        CHECK (jsonb_typeof(engine_config) = 'object'),
    CONSTRAINT chk_voice_agents_interruption_policy
        CHECK (interruption_policy IN ('allow', 'disabled')),
    CONSTRAINT chk_voice_agents_recording_policy
        CHECK (recording_policy IN ('none', 'all')),
    CONSTRAINT chk_voice_agents_configuration_revision
        CHECK (configuration_revision > 0),
    CONSTRAINT chk_voice_agents_status
        CHECK (status IN ('active', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_voice_agents_organization
    ON voice_agents (organization_id, created_at DESC);

CREATE TABLE IF NOT EXISTS voice_agent_bindings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    voice_agent_id UUID NOT NULL,
    phone_number_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_voice_agent_bindings_agent_scope
        FOREIGN KEY (voice_agent_id, organization_id)
        REFERENCES voice_agents(id, organization_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_voice_agent_bindings_phone_number_scope
        FOREIGN KEY (phone_number_id, organization_id)
        REFERENCES phone_numbers(id, organization_id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_voice_agent_bindings_phone_number
    ON voice_agent_bindings (phone_number_id);

CREATE INDEX IF NOT EXISTS idx_voice_agent_bindings_agent
    ON voice_agent_bindings (organization_id, voice_agent_id, created_at DESC);

CREATE TRIGGER set_voice_agents_updated_at
BEFORE UPDATE ON voice_agents
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE FUNCTION bump_voice_agent_configuration_revision()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE voice_agents
    SET configuration_revision = configuration_revision + 1,
        updated_at = now()
    WHERE id = COALESCE(NEW.voice_agent_id, OLD.voice_agent_id)
      AND organization_id = COALESCE(NEW.organization_id, OLD.organization_id);
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
