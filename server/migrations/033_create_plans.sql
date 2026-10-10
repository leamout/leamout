CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    currency TEXT NOT NULL DEFAULT 'USD',
    amount_minor BIGINT NOT NULL DEFAULT 0,
    billing_interval TEXT NOT NULL DEFAULT 'month',
    status TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_subscription_plans_code UNIQUE (code),
    CONSTRAINT chk_subscription_plans_code
        CHECK (length(btrim(code)) BETWEEN 1 AND 64),
    CONSTRAINT chk_subscription_plans_name
        CHECK (length(btrim(name)) BETWEEN 1 AND 128),
    CONSTRAINT chk_subscription_plans_currency
        CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_subscription_plans_amount
        CHECK (amount_minor >= 0),
    CONSTRAINT chk_subscription_plans_interval
        CHECK (billing_interval IN ('month', 'year')),
    CONSTRAINT chk_subscription_plans_status
        CHECK (status IN ('active', 'inactive'))
);

CREATE TABLE IF NOT EXISTS subscription_plan_entitlements (
    plan_id UUID NOT NULL REFERENCES subscription_plans(id) ON DELETE CASCADE,
    capability TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,

    PRIMARY KEY (plan_id, capability),
    CONSTRAINT chk_subscription_plan_entitlements_capability
        CHECK (capability IN (
            'sso',
            'scim',
            'advanced_rbac',
            'retention_policies',
            'private_networking'
        ))
);

CREATE INDEX IF NOT EXISTS idx_subscription_plans_active
    ON subscription_plans (status, amount_minor, created_at);

CREATE TRIGGER set_subscription_plans_updated_at
BEFORE UPDATE ON subscription_plans
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
