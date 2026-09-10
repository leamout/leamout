CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    plan_id UUID NOT NULL REFERENCES plans(id),
    price_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    renews_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_subscriptions_id_organization UNIQUE (id, organization_id),
    CONSTRAINT fk_subscriptions_price_plan
        FOREIGN KEY (price_id, plan_id)
        REFERENCES prices(id, plan_id)
        ON DELETE RESTRICT,
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

COMMENT ON TABLE subscriptions IS
    'Leamout-owned recurring software access. Payment providers are settlement adapters and do not own subscription identity or lifecycle.';

CREATE UNIQUE INDEX IF NOT EXISTS uq_subscriptions_current_organization
    ON subscriptions (organization_id)
    WHERE status IN ('active', 'past_due');

CREATE INDEX IF NOT EXISTS idx_subscriptions_organization_status
    ON subscriptions (organization_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_status
    ON subscriptions (plan_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_subscriptions_price
    ON subscriptions (price_id);

CREATE INDEX IF NOT EXISTS idx_subscriptions_renews_at
    ON subscriptions (renews_at)
    WHERE status = 'active' AND renews_at IS NOT NULL;

CREATE TRIGGER set_subscriptions_updated_at
BEFORE UPDATE ON subscriptions
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
