CREATE TABLE IF NOT EXISTS tenant_members (
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'member',
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (tenant_id, user_id),
    CONSTRAINT chk_tenant_members_role CHECK (role IN ('admin', 'member')),
    CONSTRAINT chk_tenant_members_status CHECK (status IN ('active', 'disabled'))
);

-- Indexes for efficient querying of tenant memberships
CREATE INDEX IF NOT EXISTS idx_tenant_members_user_id
    ON tenant_members (user_id);

CREATE INDEX IF NOT EXISTS idx_tenant_members_tenant_id_status
    ON tenant_members (tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_tenant_members_user_id_status
    ON tenant_members (user_id, status);

-- Trigger to auto-update updated_at
CREATE TRIGGER set_tenant_members_updated_at
BEFORE UPDATE ON tenant_members
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
