CREATE TABLE IF NOT EXISTS license_entitlements (
    license_id UUID NOT NULL REFERENCES licenses(id) ON DELETE CASCADE,
    entitlement_key TEXT NOT NULL,
    kind TEXT NOT NULL,
    enabled BOOLEAN,
    limit_value BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (license_id, entitlement_key),
    CONSTRAINT chk_license_entitlements_key CHECK (
        length(trim(entitlement_key)) > 0 AND entitlement_key !~ '[[:space:]]'
    ),
    CONSTRAINT chk_license_entitlements_kind CHECK (kind IN ('feature', 'limit')),
    CONSTRAINT chk_license_entitlements_value CHECK (
        (kind = 'feature' AND enabled IS NOT NULL AND limit_value IS NULL)
        OR
        (kind = 'limit' AND enabled IS NULL AND limit_value IS NOT NULL AND limit_value >= 0)
    )
);

CREATE INDEX IF NOT EXISTS idx_license_entitlements_kind
    ON license_entitlements (kind, entitlement_key);
