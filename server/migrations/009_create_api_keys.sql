CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    token_hash TEXT NOT NULL,
    token_prefix TEXT NOT NULL,
    scopes JSONB NOT NULL DEFAULT '["*"]'::jsonb,
    last_used_at TIMESTAMPTZ,
    last_used_ip INET,
    expires_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_api_keys_scopes_is_array
        CHECK (jsonb_typeof(scopes) = 'array')
);

-- One active key per tenant with same name (prevent duplicates)
CREATE UNIQUE INDEX idx_api_keys_tenant_name_active
    ON api_keys (tenant_id, name)
    WHERE disabled_at IS NULL;

CREATE INDEX idx_api_keys_tenant_id
    ON api_keys (tenant_id);

CREATE INDEX idx_api_keys_token_hash
    ON api_keys (token_hash);

CREATE INDEX idx_api_keys_token_prefix
    ON api_keys (token_prefix);

CREATE TRIGGER set_api_keys_updated_at
BEFORE UPDATE ON api_keys
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
