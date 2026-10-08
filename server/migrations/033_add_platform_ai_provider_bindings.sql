ALTER TABLE voice_agent_provider_bindings
    DROP CONSTRAINT fk_voice_agent_provider_bindings_credential_scope;

ALTER TABLE voice_agent_provider_bindings
    ALTER COLUMN credential_id DROP NOT NULL;

ALTER TABLE voice_agent_provider_bindings
    ADD CONSTRAINT fk_voice_agent_provider_bindings_credential_scope
        FOREIGN KEY (credential_id, organization_id)
        REFERENCES ai_provider_credentials(id, organization_id)
        ON DELETE RESTRICT;
