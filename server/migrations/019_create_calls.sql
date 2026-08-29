CREATE TABLE IF NOT EXISTS calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    application_id UUID REFERENCES voice_applications(id) ON DELETE SET NULL,
    direction TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'initiating',
    media_state TEXT NOT NULL DEFAULT 'active',
    from_uri TEXT NOT NULL,
    to_uri TEXT NOT NULL,
    sip_call_id TEXT,
    provider_id UUID,
    started_at TIMESTAMPTZ,
    answered_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    hangup_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_calls_direction CHECK (direction IN ('inbound', 'outbound')),
    CONSTRAINT chk_calls_state CHECK (
        state IN ('initiating', 'ringing', 'answered', 'active', 'completed', 'failed', 'cancelled')
    ),
    CONSTRAINT chk_calls_media_state CHECK (media_state IN ('active', 'held')),
    CONSTRAINT chk_calls_from_uri CHECK (length(trim(from_uri)) > 0),
    CONSTRAINT chk_calls_to_uri CHECK (length(trim(to_uri)) > 0),
    CONSTRAINT chk_calls_timestamps CHECK (
        answered_at IS NULL OR started_at IS NULL OR answered_at >= started_at
    ),
    CONSTRAINT chk_calls_ended_at CHECK (
        ended_at IS NULL OR started_at IS NULL OR ended_at >= started_at
    )
);

CREATE INDEX IF NOT EXISTS idx_calls_organization_created
    ON calls (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_calls_organization_state
    ON calls (organization_id, state, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_calls_application_created
    ON calls (application_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_calls_organization_sip_call_id
    ON calls (organization_id, sip_call_id)
    WHERE sip_call_id IS NOT NULL;

CREATE TRIGGER set_calls_updated_at
BEFORE UPDATE ON calls
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
