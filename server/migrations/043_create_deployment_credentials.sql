CREATE TABLE IF NOT EXISTS deployment_credentials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id),
    purpose TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    token_prefix TEXT NOT NULL,
    scopes JSONB NOT NULL DEFAULT '[]'::JSONB,
    expires_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_deployment_credentials_token_hash UNIQUE (token_hash),
    CONSTRAINT chk_deployment_credentials_purpose CHECK (purpose IN ('managed_carrier')),
    CONSTRAINT chk_deployment_credentials_token_hash CHECK (length(trim(token_hash)) > 0),
    CONSTRAINT chk_deployment_credentials_token_prefix CHECK (length(trim(token_prefix)) > 0),
    CONSTRAINT chk_deployment_credentials_scopes CHECK (jsonb_typeof(scopes) = 'array'),
    CONSTRAINT chk_deployment_credentials_expires_at CHECK (
        expires_at IS NULL OR expires_at > created_at
    ),
    CONSTRAINT chk_deployment_credentials_last_used_at CHECK (
        last_used_at IS NULL OR last_used_at >= created_at
    ),
    CONSTRAINT chk_deployment_credentials_disabled_at CHECK (
        disabled_at IS NULL OR disabled_at >= created_at
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_deployment_credentials_active_purpose
    ON deployment_credentials (deployment_id, purpose)
    WHERE disabled_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_deployment_credentials_token_prefix
    ON deployment_credentials (token_prefix);

CREATE INDEX IF NOT EXISTS idx_deployment_credentials_deployment
    ON deployment_credentials (deployment_id, created_at DESC);

CREATE TRIGGER set_deployment_credentials_updated_at
BEFORE UPDATE ON deployment_credentials
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
