-- Existing NULL bindings remain unready until an organization integration is assigned.
-- NOT VALID allows an existing deployment to migrate without deleting user data,
-- while PostgreSQL immediately enforces these rules for all new and updated rows.

ALTER TABLE ai_provider_credentials
    ADD CONSTRAINT uq_ai_provider_credentials_scope_provider
    UNIQUE (id, organization_id, provider);

ALTER TABLE voice_agent_provider_bindings
    ADD CONSTRAINT chk_voice_agent_provider_bindings_byoa
    CHECK (credential_id IS NOT NULL) NOT VALID;

ALTER TABLE voice_agent_provider_bindings
    ADD CONSTRAINT fk_voice_agent_provider_bindings_provider_identity
    FOREIGN KEY (credential_id, organization_id, provider)
    REFERENCES ai_provider_credentials (id, organization_id, provider)
    ON DELETE RESTRICT
    NOT VALID;

-- After legacy rows have been re-bound:
-- ALTER TABLE voice_agent_provider_bindings
--     VALIDATE CONSTRAINT chk_voice_agent_provider_bindings_byoa;
-- ALTER TABLE voice_agent_provider_bindings
--     VALIDATE CONSTRAINT fk_voice_agent_provider_bindings_provider_identity;
