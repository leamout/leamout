CREATE TABLE IF NOT EXISTS usage_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    meter_id UUID NOT NULL REFERENCES meters(id),
    quantity BIGINT NOT NULL,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    dimensions JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_usage_events_organization_idempotency
        UNIQUE (organization_id, idempotency_key),
    CONSTRAINT chk_usage_events_quantity CHECK (quantity > 0),
    CONSTRAINT chk_usage_events_source_type CHECK (
        source_type ~ '^[a-z0-9]+(?:_[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_usage_events_source_id CHECK (length(btrim(source_id)) > 0),
    CONSTRAINT chk_usage_events_idempotency_key CHECK (
        length(btrim(idempotency_key)) > 0
    ),
    CONSTRAINT chk_usage_events_dimensions_object CHECK (
        jsonb_typeof(dimensions) = 'object'
    )
);

COMMENT ON TABLE usage_events IS
    'Immutable usage observations. Recording usage does not by itself make that usage billable.';

CREATE INDEX IF NOT EXISTS idx_usage_events_organization_meter_occurred
    ON usage_events (organization_id, meter_id, occurred_at);

CREATE INDEX IF NOT EXISTS idx_usage_events_source
    ON usage_events (source_type, source_id);
