-- Keep the FreeSWITCH channel and background-job identities durable before
-- issuing a remote originate command. A missing ACK must not cause a new INVITE.
CREATE TABLE IF NOT EXISTS call_origination_attempts (
    channel_id UUID PRIMARY KEY,
    call_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    trunk_id UUID NOT NULL,
    trunk_endpoint_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'prepared',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_call_originations_call
        FOREIGN KEY (call_id, organization_id)
        REFERENCES calls(id, organization_id)
        ON DELETE CASCADE,
    CONSTRAINT chk_call_originations_status CHECK (
        status IN ('prepared', 'submitted', 'uncertain', 'observed', 'failed', 'finished')
    )
);

CREATE INDEX IF NOT EXISTS idx_call_originations_unresolved
    ON call_origination_attempts (created_at, channel_id)
    WHERE status IN ('prepared', 'submitted', 'uncertain');

CREATE INDEX IF NOT EXISTS idx_call_originations_call
    ON call_origination_attempts (call_id);
