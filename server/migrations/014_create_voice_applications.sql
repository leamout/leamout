CREATE TABLE IF NOT EXISTS voice_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_voice_applications_tenant_name UNIQUE (tenant_id, name),
    CONSTRAINT chk_voice_applications_name CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_voice_applications_status CHECK (status IN ('active', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_voice_applications_tenant_id
    ON voice_applications (tenant_id);

CREATE INDEX IF NOT EXISTS idx_voice_applications_tenant_status
    ON voice_applications (tenant_id, status);

CREATE TRIGGER set_voice_applications_updated_at
BEFORE UPDATE ON voice_applications
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
