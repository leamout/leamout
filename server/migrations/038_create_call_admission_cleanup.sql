-- Terminal call state and deferred Redis cleanup are committed together.
-- A failed or interrupted release can be retried without reviving the call.
CREATE TABLE IF NOT EXISTS call_admission_cleanup (
    call_id UUID PRIMARY KEY REFERENCES calls(id) ON DELETE CASCADE,
    organization_id UUID NOT NULL,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_call_admission_cleanup_due
    ON call_admission_cleanup (next_attempt_at, call_id);

-- Keep pending origination checks fair when the first 100 attempts lack
-- conclusive FreeSWITCH evidence.
CREATE INDEX IF NOT EXISTS idx_call_origination_attempts_check
    ON call_origination_attempts (updated_at, channel_id)
    WHERE status IN ('prepared', 'submitted', 'uncertain');
