CREATE TABLE IF NOT EXISTS prices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES plans(id),
    currency TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    billing_interval TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT now(),
    effective_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_prices_id_plan UNIQUE (id, plan_id),
    CONSTRAINT chk_prices_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_prices_amount_minor CHECK (amount_minor >= 0),
    CONSTRAINT chk_prices_billing_interval CHECK (
        billing_interval IN ('month', 'year')
    ),
    CONSTRAINT chk_prices_effective_period CHECK (
        effective_until IS NULL OR effective_until > effective_from
    )
);

CREATE INDEX IF NOT EXISTS idx_prices_plan_effective
    ON prices (plan_id, effective_from DESC, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_prices_open_active_terms
    ON prices (plan_id, currency, billing_interval)
    WHERE active AND effective_until IS NULL;

CREATE OR REPLACE FUNCTION enforce_price_terms_immutable()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.plan_id IS DISTINCT FROM OLD.plan_id
        OR NEW.currency IS DISTINCT FROM OLD.currency
        OR NEW.amount_minor IS DISTINCT FROM OLD.amount_minor
        OR NEW.billing_interval IS DISTINCT FROM OLD.billing_interval
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
