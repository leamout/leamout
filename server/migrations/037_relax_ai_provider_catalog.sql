ALTER TABLE ai_provider_credentials
    DROP CONSTRAINT IF EXISTS chk_ai_provider_credentials_provider,
    ADD CONSTRAINT chk_ai_provider_credentials_provider
        CHECK (length(btrim(provider)) BETWEEN 1 AND 64);

ALTER TABLE voice_agent_provider_bindings
    DROP CONSTRAINT IF EXISTS chk_voice_agent_provider_bindings_provider,
    ADD CONSTRAINT chk_voice_agent_provider_bindings_provider
        CHECK (length(btrim(provider)) BETWEEN 1 AND 64);
