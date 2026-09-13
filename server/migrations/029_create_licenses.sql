CREATE TABLE IF NOT EXISTS licenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    status TEXT NOT NULL DEFAULT 'pending',
    signing_key_id TEXT,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_licenses_status CHECK (
        status IN ('pending', 'active', 'suspended', 'expired', 'revoked')
    ),
    CONSTRAINT chk_licenses_signing_key_id CHECK (
        signing_key_id IS NULL OR length(trim(signing_key_id)) > 0
    ),
    CONSTRAINT chk_licenses_expires_at CHECK (
        expires_at IS NULL OR expires_at >= issued_at
    )
);

CREATE INDEX IF NOT EXISTS idx_licenses_organization_status
    ON licenses (organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_licenses_expires_at
    ON licenses (expires_at)
    WHERE status = 'active' AND expires_at IS NOT NULL;

CREATE TRIGGER set_licenses_updated_at
BEFORE UPDATE ON licenses
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
