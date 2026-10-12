CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    pricing_type TEXT NOT NULL DEFAULT 'fixed',
    currency TEXT NOT NULL DEFAULT 'USD',
    amount_minor BIGINT,
    billing_interval TEXT NOT NULL DEFAULT 'month',
    entitlements JSONB NOT NULL DEFAULT '{}'::jsonb,
    limits JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_plans_code UNIQUE (code),
    CONSTRAINT chk_plans_code
        CHECK (length(btrim(code)) BETWEEN 1 AND 64),
    CONSTRAINT chk_plans_name
        CHECK (length(btrim(name)) BETWEEN 1 AND 128),
    CONSTRAINT chk_plans_pricing_type
        CHECK (pricing_type IN ('fixed', 'custom')),
    CONSTRAINT chk_plans_currency
        CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_plans_pricing
        CHECK (
            (pricing_type = 'fixed' AND amount_minor IS NOT NULL AND amount_minor >= 0)
            OR
            (pricing_type = 'custom' AND amount_minor IS NULL)
        ),
    CONSTRAINT chk_plans_interval
        CHECK (billing_interval IN ('month', 'year')),
    CONSTRAINT chk_plans_entitlements
        CHECK (jsonb_typeof(entitlements) = 'object'),
    CONSTRAINT chk_plans_limits
        CHECK (jsonb_typeof(limits) = 'object'),
    CONSTRAINT chk_plans_status
        CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX IF NOT EXISTS idx_plans_active
    ON plans (status, amount_minor, created_at);

CREATE TRIGGER set_plans_updated_at
BEFORE UPDATE ON plans
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

INSERT INTO plans (
    code,
    name,
    description,
    pricing_type,
    currency,
    amount_minor,
    billing_interval,
    entitlements,
    limits,
    status
)
VALUES
    (
        'free',
        'Free',
        'Build and test voice agents before moving to production.',
        'fixed',
        'USD',
        0,
        'month',
        '{
            "sso": false,
            "scim": false,
            "advanced_rbac": false,
            "retention_policies": false,
            "private_networking": false
        }'::jsonb,
        '{
            "max_concurrent_calls": 1,
            "retention_days": 1,
            "monthly_runtime_minutes": 100,
            "max_voice_agents": 1,
            "max_sip_trunks": 1
        }'::jsonb,
        'active'
    ),
    (
        'developer',
        'Developer',
        'For developers and small production voice-agent deployments.',
        'fixed',
        'USD',
        7900,
        'month',
        '{
            "sso": false,
            "scim": false,
            "advanced_rbac": false,
            "retention_policies": false,
            "private_networking": false
        }'::jsonb,
        '{
            "max_concurrent_calls": 3,
            "retention_days": 14,
            "monthly_runtime_minutes": 2000,
            "max_voice_agents": 5,
            "max_sip_trunks": 3
        }'::jsonb,
        'active'
    ),
    (
        'pro',
        'Pro',
        'For production teams running higher-volume voice-agent workloads.',
        'fixed',
        'USD',
        29900,
        'month',
        '{
            "sso": false,
            "scim": false,
            "advanced_rbac": true,
            "retention_policies": true,
            "private_networking": false
        }'::jsonb,
        '{
            "max_concurrent_calls": 15,
            "retention_days": 30,
            "monthly_runtime_minutes": 10000,
            "max_voice_agents": 25,
            "max_sip_trunks": 10
        }'::jsonb,
        'active'
    ),
    (
        'scale',
        'Scale',
        'For high-volume production voice agents and expanding teams.',
        'fixed',
        'USD',
        79900,
        'month',
        '{
            "sso": false,
            "scim": false,
            "advanced_rbac": true,
            "retention_policies": true,
            "private_networking": false
        }'::jsonb,
        '{
            "max_concurrent_calls": 40,
            "retention_days": 90,
            "monthly_runtime_minutes": 30000,
            "max_voice_agents": 100,
            "max_sip_trunks": 30
        }'::jsonb,
        'active'
    ),
    (
        'enterprise',
        'Enterprise',
        'For organizations that need custom scale, security, networking, and support.',
        'custom',
        'USD',
        NULL,
        'year',
        '{
            "sso": true,
            "scim": true,
            "advanced_rbac": true,
            "retention_policies": true,
            "private_networking": true
        }'::jsonb,
        '{}'::jsonb,
        'active'
    )
ON CONFLICT (code) DO NOTHING;
