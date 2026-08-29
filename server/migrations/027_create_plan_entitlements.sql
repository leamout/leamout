CREATE TABLE IF NOT EXISTS plan_entitlements (
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    entitlement_key TEXT NOT NULL,
    kind TEXT NOT NULL,
    enabled BOOLEAN,
    limit_value BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (plan_id, entitlement_key),
    CONSTRAINT chk_plan_entitlements_key CHECK (
        length(trim(entitlement_key)) > 0 AND entitlement_key !~ '[[:space:]]'
    ),
    CONSTRAINT chk_plan_entitlements_kind CHECK (kind IN ('feature', 'limit')),
    CONSTRAINT chk_plan_entitlements_value CHECK (
        (kind = 'feature' AND enabled IS NOT NULL AND limit_value IS NULL)
        OR
        (kind = 'limit' AND enabled IS NULL AND limit_value IS NOT NULL AND limit_value >= 0)
    )
);

CREATE INDEX IF NOT EXISTS idx_plan_entitlements_kind
    ON plan_entitlements (kind, entitlement_key);

CREATE TRIGGER set_plan_entitlements_updated_at
BEFORE UPDATE ON plan_entitlements
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
