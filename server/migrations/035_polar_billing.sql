-- Preserve historical provider identities; all new billing writes use Polar.
ALTER TABLE subscriptions DROP CONSTRAINT chk_subscriptions_provider;
ALTER TABLE subscriptions ADD CONSTRAINT chk_subscriptions_provider
    CHECK (provider IS NULL OR provider IN ('stripe', 'polar'));
