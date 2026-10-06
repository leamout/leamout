CREATE TABLE email_deliveries (
    id UUID PRIMARY KEY,
    recipient TEXT NOT NULL,
    template TEXT NOT NULL,
    encrypted_data TEXT,
    cancellation_key TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'expired', 'cancelled')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    locked_at TIMESTAMPTZ,
    lock_token UUID,
    expires_at TIMESTAMPTZ NOT NULL,
    provider_message_id TEXT,
    last_error_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at TIMESTAMPTZ
);

CREATE INDEX email_deliveries_pending_idx
    ON email_deliveries (available_at)
    WHERE status IN ('pending', 'sending');

CREATE INDEX email_deliveries_cancellation_idx
    ON email_deliveries (cancellation_key)
    WHERE status IN ('pending', 'sending');

CREATE INDEX email_deliveries_recipient_idx
    ON email_deliveries (recipient, template, created_at);
