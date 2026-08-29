CREATE TABLE IF NOT EXISTS call_events (
    id UUID PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    call_id UUID NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_call_events_type CHECK (length(trim(event_type)) > 0),
    CONSTRAINT chk_call_events_payload_object CHECK (jsonb_typeof(payload) = 'object')
);

CREATE INDEX IF NOT EXISTS idx_call_events_call_occurred
    ON call_events (organization_id, call_id, occurred_at, created_at);

CREATE INDEX IF NOT EXISTS idx_call_events_type_occurred
    ON call_events (organization_id, event_type, occurred_at DESC);
