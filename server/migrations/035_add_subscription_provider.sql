ALTER TABLE subscriptions
    ADD COLUMN provider TEXT,
    ADD COLUMN provider_customer_id TEXT,
    ADD COLUMN provider_subscription_id TEXT,
    ADD COLUMN provider_event_created_at TIMESTAMPTZ,
    ADD CONSTRAINT chk_subscriptions_provider
        CHECK (provider IS NULL OR provider = 'stripe');

CREATE UNIQUE INDEX uq_subscriptions_provider_subscription
    ON subscriptions (provider, provider_subscription_id)
    WHERE provider IS NOT NULL
      AND provider_subscription_id IS NOT NULL;

CREATE INDEX idx_subscriptions_provider_customer
    ON subscriptions (provider, provider_customer_id)
    WHERE provider IS NOT NULL
      AND provider_customer_id IS NOT NULL;
