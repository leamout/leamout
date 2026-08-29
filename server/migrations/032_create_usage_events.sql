CREATE TABLE IF NOT EXISTS usage_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    subscription_id UUID,
    meter_id UUID NOT NULL REFERENCES meters(id),
    quantity BIGINT NOT NULL,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    dimensions JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_usage_events_subscription_organization
        FOREIGN KEY (subscription_id, organization_id)
        REFERENCES subscriptions (id, organization_id)
        ON DELETE SET NULL (subscription_id),
    CONSTRAINT chk_usage_events_quantity CHECK (quantity > 0),
    CONSTRAINT chk_usage_events_source_type CHECK (
        length(trim(source_type)) > 0 AND source_type !~ '[[:space:]]'
    ),
    CONSTRAINT chk_usage_events_source_id CHECK (length(trim(source_id)) > 0),
    CONSTRAINT chk_usage_events_idempotency_key CHECK (
        length(trim(idempotency_key)) > 0
    ),
    CONSTRAINT chk_usage_events_dimensions_object CHECK (
        jsonb_typeof(dimensions) = 'object'
    )
);

CREATE INDEX IF NOT EXISTS idx_usage_events_organization_meter_occurred
    ON usage_events (organization_id, meter_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_usage_events_subscription_meter_occurred
    ON usage_events (subscription_id, meter_id, occurred_at)
    WHERE subscription_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_usage_events_source
    ON usage_events (source_type, source_id);
