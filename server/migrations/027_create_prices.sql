CREATE TABLE IF NOT EXISTS prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
    meter_id UUID REFERENCES meters(id) ON DELETE RESTRICT,
    pricing_type TEXT NOT NULL,
    currency TEXT NOT NULL,
    amount_minor BIGINT,
    billing_interval TEXT,
    unit_amount_micros BIGINT,
    unit_size BIGINT,
    dimensions JSONB NOT NULL DEFAULT '{}'::jsonb,
    active BOOLEAN NOT NULL DEFAULT true,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
    effective_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_prices_id_plan UNIQUE (id, plan_id),
    CONSTRAINT uq_prices_id_currency UNIQUE (id, currency),
    CONSTRAINT chk_prices_pricing_type CHECK (
        pricing_type IN ('one_time', 'recurring', 'metered')
    ),
    CONSTRAINT chk_prices_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_prices_billing_interval CHECK (
        billing_interval IS NULL OR billing_interval IN ('month', 'year')
    ),
    CONSTRAINT chk_prices_dimensions CHECK (jsonb_typeof(dimensions) = 'object'),
    CONSTRAINT chk_prices_terms CHECK (
        (
            pricing_type = 'one_time'
            AND amount_minor IS NOT NULL
            AND amount_minor >= 0
            AND billing_interval IS NULL
            AND meter_id IS NULL
            AND unit_amount_micros IS NULL
            AND unit_size IS NULL
            AND dimensions = '{}'::jsonb
        )
        OR (
            pricing_type = 'recurring'
            AND amount_minor IS NOT NULL
            AND amount_minor >= 0
            AND billing_interval IS NOT NULL
            AND meter_id IS NULL
            AND unit_amount_micros IS NULL
            AND unit_size IS NULL
            AND dimensions = '{}'::jsonb
        )
        OR (
            pricing_type = 'metered'
            AND amount_minor IS NULL
            AND billing_interval IS NULL
            AND meter_id IS NOT NULL
            AND unit_amount_micros IS NOT NULL
            AND unit_amount_micros >= 0
            AND unit_size IS NOT NULL
            AND unit_size > 0
        )
    ),
    CONSTRAINT chk_prices_effective_period CHECK (
        effective_until IS NULL OR effective_until > effective_from
    )
);

COMMENT ON TABLE prices IS
    'Customer-facing catalog prices. Metered prices rate Leamout usage; upstream provider cost is tracked separately as COGS.';

CREATE INDEX IF NOT EXISTS idx_prices_plan_effective
    ON prices (plan_id, effective_from DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_prices_meter_effective
    ON prices (meter_id, effective_from DESC, created_at DESC)
    WHERE pricing_type = 'metered';

CREATE UNIQUE INDEX IF NOT EXISTS uq_prices_open_recurring_terms
    ON prices (plan_id, currency, billing_interval)
    WHERE active AND effective_until IS NULL AND pricing_type = 'recurring';

CREATE UNIQUE INDEX IF NOT EXISTS uq_prices_open_one_time_terms
    ON prices (plan_id, currency)
    WHERE active AND effective_until IS NULL AND pricing_type = 'one_time';

CREATE UNIQUE INDEX IF NOT EXISTS uq_prices_open_metered_terms
    ON prices (plan_id, meter_id, currency, dimensions)
    WHERE active AND effective_until IS NULL AND pricing_type = 'metered';

CREATE OR REPLACE FUNCTION enforce_price_terms_immutable()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.plan_id IS DISTINCT FROM OLD.plan_id
        OR NEW.meter_id IS DISTINCT FROM OLD.meter_id
        OR NEW.pricing_type IS DISTINCT FROM OLD.pricing_type
        OR NEW.currency IS DISTINCT FROM OLD.currency
        OR NEW.amount_minor IS DISTINCT FROM OLD.amount_minor
        OR NEW.billing_interval IS DISTINCT FROM OLD.billing_interval
        OR NEW.unit_amount_micros IS DISTINCT FROM OLD.unit_amount_micros
        OR NEW.unit_size IS DISTINCT FROM OLD.unit_size
        OR NEW.dimensions IS DISTINCT FROM OLD.dimensions
        OR NEW.effective_from IS DISTINCT FROM OLD.effective_from THEN
        RAISE EXCEPTION 'price commercial terms are immutable';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_prices_terms_immutable
BEFORE UPDATE ON prices
FOR EACH ROW
EXECUTE FUNCTION enforce_price_terms_immutable();
