CREATE TABLE IF NOT EXISTS tenant_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email CITEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    token_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    declined_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- One pending invitation per email per tenant
    CONSTRAINT chk_tenant_invitations_status
        CHECK (status IN ('pending', 'accepted', 'declined', 'expired', 'revoked')),
    CONSTRAINT chk_tenant_invitations_role
        CHECK (role IN ('admin', 'member'))
);

-- Prevent duplicate pending invitations for same email + tenant
CREATE UNIQUE INDEX idx_tenant_invitations_pending_email
    ON tenant_invitations (tenant_id, email)
    WHERE status = 'pending';

CREATE INDEX idx_tenant_invitations_tenant_id
    ON tenant_invitations (tenant_id);

CREATE INDEX idx_tenant_invitations_token_hash
    ON tenant_invitations (token_hash);

CREATE INDEX idx_tenant_invitations_email
    ON tenant_invitations (email);

CREATE INDEX idx_tenant_invitations_expires_at
    ON tenant_invitations (expires_at)
    WHERE status = 'pending';

CREATE TRIGGER set_tenant_invitations_updated_at
BEFORE UPDATE ON tenant_invitations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
