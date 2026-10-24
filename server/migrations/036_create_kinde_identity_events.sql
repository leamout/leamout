CREATE TABLE kinde_identity_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer TEXT NOT NULL,
    event_id TEXT NOT NULL,
    subject TEXT NOT NULL,
    event_type TEXT NOT NULL
        CHECK (event_type IN ('user.updated', 'user.deleted')),
    occurred_at TIMESTAMPTZ NOT NULL,
    disable_user BOOLEAN NOT NULL DEFAULT FALSE,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (issuer, event_id)
);

CREATE INDEX kinde_identity_events_subject_idx
    ON kinde_identity_events (issuer, subject, occurred_at DESC);
