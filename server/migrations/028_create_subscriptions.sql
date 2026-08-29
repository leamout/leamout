CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    plan_id UUID NOT NULL REFERENCES plans(id),
    status TEXT NOT NULL DEFAULT 'pending',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    renews_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_subscriptions_status CHECK (
        status IN ('pending', 'active', 'past_due', 'cancelled', 'expired')
    ),
    CONSTRAINT chk_subscriptions_renews_at CHECK (
        renews_at IS NULL OR renews_at >= starts_at
    ),
    CONSTRAINT chk_subscriptions_ends_at CHECK (
        ends_at IS NULL OR ends_at >= starts_at
    ),
    CONSTRAINT chk_subscriptions_period CHECK (
        renews_at IS NULL OR ends_at IS NULL OR renews_at <= ends_at
    )
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_customer_status
    ON subscriptions (customer_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_status
    ON subscriptions (plan_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscriptions_renews_at
    ON subscriptions (renews_at)
    WHERE status = 'active' AND renews_at IS NOT NULL;

CREATE TRIGGER set_subscriptions_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
