CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    plan_id UUID NOT NULL REFERENCES plans(id),
    status TEXT NOT NULL DEFAULT 'pending',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    renews_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    billing_provider TEXT,
    provider_subscription_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_subscriptions_id_organization UNIQUE (id, organization_id),
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
    ),
    CONSTRAINT chk_subscriptions_billing_provider CHECK (
        billing_provider IS NULL OR (length(trim(billing_provider)) > 0 AND billing_provider !~ '[[:space:]]')
    ),
    CONSTRAINT chk_subscriptions_provider_subscription_id CHECK (
        provider_subscription_id IS NULL OR length(trim(provider_subscription_id)) > 0
    ),
    CONSTRAINT chk_subscriptions_billing_pair CHECK (
        (billing_provider IS NULL AND provider_subscription_id IS NULL)
        OR (billing_provider IS NOT NULL AND provider_subscription_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_subscriptions_provider_subscription
    ON subscriptions (billing_provider, provider_subscription_id)
    WHERE billing_provider IS NOT NULL AND provider_subscription_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_subscriptions_organization_status
    ON subscriptions (organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_status
    ON subscriptions (plan_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscriptions_renews_at
    ON subscriptions (renews_at)
    WHERE status = 'active' AND renews_at IS NOT NULL;

CREATE TRIGGER set_subscriptions_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
