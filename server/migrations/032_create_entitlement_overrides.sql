CREATE TABLE IF NOT EXISTS customer_entitlement_overrides (
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    entitlement_key TEXT NOT NULL,
    kind TEXT NOT NULL,
    enabled BOOLEAN,
    limit_value BIGINT,
    starts_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (customer_id, entitlement_key),
    CONSTRAINT chk_customer_entitlement_overrides_key CHECK (
        length(trim(entitlement_key)) > 0 AND entitlement_key !~ '[[:space:]]'
    ),
    CONSTRAINT chk_customer_entitlement_overrides_kind CHECK (kind IN ('feature', 'limit')),
    CONSTRAINT chk_customer_entitlement_overrides_value CHECK (
        (kind = 'feature' AND enabled IS NOT NULL AND limit_value IS NULL)
        OR
        (kind = 'limit' AND enabled IS NULL AND limit_value IS NOT NULL AND limit_value >= 0)
    ),
    CONSTRAINT chk_customer_entitlement_overrides_period CHECK (
        starts_at IS NULL OR expires_at IS NULL OR expires_at >= starts_at
    )
);

CREATE INDEX IF NOT EXISTS idx_customer_entitlement_overrides_expires_at
    ON customer_entitlement_overrides (expires_at)
    WHERE expires_at IS NOT NULL;

CREATE TRIGGER set_customer_entitlement_overrides_updated_at
BEFORE UPDATE ON customer_entitlement_overrides
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
