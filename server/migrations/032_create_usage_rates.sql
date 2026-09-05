CREATE TABLE IF NOT EXISTS usage_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    meter_id UUID NOT NULL REFERENCES meters(id) ON DELETE CASCADE,
    carrier_provider_id UUID REFERENCES carrier_providers(id) ON DELETE SET NULL,
    direction TEXT,
    country_code TEXT,
    network TEXT,
    currency TEXT NOT NULL,
    unit_amount_micros BIGINT NOT NULL,
    unit_size BIGINT NOT NULL DEFAULT 1,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
    effective_until TIMESTAMPTZ,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_usage_rates_direction CHECK (
        direction IS NULL OR direction IN ('inbound', 'outbound')
    ),
    CONSTRAINT chk_usage_rates_country_code CHECK (
        country_code IS NULL OR country_code ~ '^[A-Z]{2}$'
    ),
    CONSTRAINT chk_usage_rates_network CHECK (
        network IS NULL OR length(trim(network)) > 0
    ),
    CONSTRAINT chk_usage_rates_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_usage_rates_unit_amount CHECK (unit_amount_micros >= 0),
    CONSTRAINT chk_usage_rates_unit_size CHECK (unit_size > 0),
    CONSTRAINT chk_usage_rates_period CHECK (
        effective_until IS NULL OR effective_until > effective_from
    )
);

COMMENT ON TABLE usage_rates IS
    'Customer-facing usage pricing rules used by Leamout rating; actual upstream carrier costs are stored separately in wholesale_charges.';

CREATE INDEX IF NOT EXISTS idx_usage_rates_lookup
    ON usage_rates (
        plan_id,
        meter_id,
        country_code,
        network,
        direction,
        effective_from DESC
    )
    WHERE active;

CREATE INDEX IF NOT EXISTS idx_usage_rates_provider
    ON usage_rates (carrier_provider_id, effective_from DESC)
    WHERE carrier_provider_id IS NOT NULL;

CREATE TRIGGER set_usage_rates_updated_at
BEFORE UPDATE ON usage_rates
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
