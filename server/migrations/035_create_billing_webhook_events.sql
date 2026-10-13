CREATE TABLE billing_webhook_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL CHECK (provider = 'polar'),
    environment TEXT NOT NULL CHECK (environment IN ('production', 'sandbox')),
    event_id TEXT NOT NULL CHECK (length(event_id) BETWEEN 1 AND 256),
    event_type TEXT NOT NULL CHECK (length(event_type) BETWEEN 1 AND 128),
    event_created_at TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    payload_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'processed', 'ignored', 'dead_letter')),
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    locked_at TIMESTAMPTZ,
    lease_token UUID,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    last_error TEXT,
    UNIQUE (provider, environment, event_id)
);

CREATE INDEX idx_billing_webhook_events_pending
    ON billing_webhook_events (environment, available_at, received_at)
    WHERE status IN ('pending', 'processing');
