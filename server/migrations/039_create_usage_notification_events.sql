CREATE TABLE usage_notification_events (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    period_start TIMESTAMPTZ NOT NULL,
    threshold_percent INTEGER NOT NULL CHECK (threshold_percent IN (80, 100)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (organization_id, period_start, threshold_percent)
);

CREATE INDEX usage_notification_events_period_idx
    ON usage_notification_events (period_start, organization_id);
