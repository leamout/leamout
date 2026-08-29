CREATE TABLE IF NOT EXISTS entitlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID REFERENCES plans(id) ON DELETE CASCADE,
    organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE,
    license_id UUID REFERENCES licenses(id) ON DELETE CASCADE,
    entitlement_key TEXT NOT NULL,
    kind TEXT NOT NULL,
    enabled BOOLEAN,
    limit_value BIGINT,
    starts_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_entitlements_owner CHECK (
        (plan_id IS NOT NULL)::int
        + (organization_id IS NOT NULL)::int
        + (license_id IS NOT NULL)::int = 1
    ),
    CONSTRAINT chk_entitlements_key CHECK (
        length(trim(entitlement_key)) > 0 AND entitlement_key !~ '[[:space:]]'
    ),
    CONSTRAINT chk_entitlements_kind CHECK (kind IN ('feature', 'limit')),
    CONSTRAINT chk_entitlements_value CHECK (
        (kind = 'feature' AND enabled IS NOT NULL AND limit_value IS NULL)
        OR
        (kind = 'limit' AND enabled IS NULL AND limit_value IS NOT NULL AND limit_value >= 0)
    ),
    CONSTRAINT chk_entitlements_period CHECK (
        starts_at IS NULL OR expires_at IS NULL OR expires_at >= starts_at
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_entitlements_plan_key
    ON entitlements (plan_id, entitlement_key)
    WHERE plan_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_entitlements_organization_key
    ON entitlements (organization_id, entitlement_key)
    WHERE organization_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_entitlements_license_key
    ON entitlements (license_id, entitlement_key)
    WHERE license_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_entitlements_expires_at
    ON entitlements (expires_at)
    WHERE expires_at IS NOT NULL;

CREATE TRIGGER set_entitlements_updated_at
BEFORE UPDATE ON entitlements
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
