CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,

    status TEXT NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    trial_ends_at TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN NOT NULL DEFAULT FALSE,
    canceled_at TIMESTAMPTZ,
    provider TEXT,
    provider_customer_id TEXT,
    provider_subscription_id TEXT,
    provider_event_created_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_subscriptions_organization UNIQUE (organization_id),
    CONSTRAINT chk_subscriptions_status
        CHECK (status IN (
            'trialing',
            'active',
            'past_due',
            'canceled',
            'incomplete'
        )),
    CONSTRAINT chk_subscriptions_period
        CHECK (
            current_period_end IS NULL
            OR current_period_start IS NULL
            OR current_period_end > current_period_start
        ),
    CONSTRAINT chk_subscriptions_provider
        CHECK (provider IS NULL OR provider = 'polar')
);

CREATE INDEX IF NOT EXISTS idx_subscriptions_plan
    ON subscriptions (plan_id, status);

CREATE INDEX IF NOT EXISTS idx_subscriptions_status
    ON subscriptions (status, current_period_end);

CREATE TRIGGER set_subscriptions_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE UNIQUE INDEX uq_subscriptions_provider_subscription
    ON subscriptions (provider, provider_subscription_id)
    WHERE provider IS NOT NULL
      AND provider_subscription_id IS NOT NULL;

CREATE INDEX idx_subscriptions_provider_customer
    ON subscriptions (provider, provider_customer_id)
    WHERE provider IS NOT NULL
      AND provider_customer_id IS NOT NULL;
