CREATE TABLE email_deliveries (
    id UUID PRIMARY KEY,
    recipient TEXT NOT NULL,
    template TEXT NOT NULL CHECK (template IN ('otp', 'invitation')),
    encrypted_data TEXT,
    challenge_id UUID REFERENCES auth_challenges(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sending', 'sent', 'failed', 'expired')),
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
CREATE INDEX email_deliveries_pending_idx ON email_deliveries (available_at) WHERE status IN ('pending', 'sending');
CREATE INDEX email_deliveries_challenge_idx ON email_deliveries (challenge_id);
CREATE INDEX email_deliveries_recipient_idx ON email_deliveries (recipient, created_at) WHERE template='otp';
