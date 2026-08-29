CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_subscription_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_billing_subscriptions_subscription_provider UNIQUE (subscription_id, provider),
    CONSTRAINT uq_billing_subscriptions_provider_subscription UNIQUE (provider, provider_subscription_id),
    CONSTRAINT chk_billing_subscriptions_provider CHECK (
        length(trim(provider)) > 0 AND provider !~ '[[:space:]]'
    ),
    CONSTRAINT chk_billing_subscriptions_provider_subscription_id CHECK (
        length(trim(provider_subscription_id)) > 0
    )
);

CREATE INDEX IF NOT EXISTS idx_billing_subscriptions_subscription_id
    ON billing_subscriptions (subscription_id);
