CREATE TABLE IF NOT EXISTS provider_cdrs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    carrier_connection_id UUID REFERENCES carrier_connections(id) ON DELETE RESTRICT,

    provider_record_id TEXT NOT NULL,
    direction TEXT NOT NULL,
    sip_call_id TEXT,

    call_id UUID,
    organization_id UUID,
    reconciled_at TIMESTAMPTZ,

    started_at TIMESTAMPTZ NOT NULL,
    duration_seconds BIGINT NOT NULL,
    currency TEXT NOT NULL,
    cost_micros BIGINT NOT NULL,
    raw JSONB NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_provider_cdr_source UNIQUE (
        provider,
        direction,
        provider_record_id
    ),
    CONSTRAINT uq_provider_cdr_id_call_organization UNIQUE (
        id,
        call_id,
        organization_id
    ),
    CONSTRAINT fk_provider_cdr_call_organization
        FOREIGN KEY (call_id, organization_id)
        REFERENCES calls (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_provider_cdr_provider CHECK (
        provider ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_provider_cdr_direction CHECK (
        direction IN ('termination', 'origination')
    ),
    CONSTRAINT chk_provider_cdr_record CHECK (
        length(btrim(provider_record_id)) > 0
    ),
    CONSTRAINT chk_provider_cdr_duration CHECK (
        duration_seconds >= 0
    ),
    CONSTRAINT chk_provider_cdr_currency CHECK (
        currency ~ '^[A-Z]{3}$'
    ),
    CONSTRAINT chk_provider_cdr_cost CHECK (
        cost_micros >= 0
    ),
    CONSTRAINT chk_provider_cdr_raw CHECK (
        jsonb_typeof(raw) = 'object'
    ),
    CONSTRAINT chk_provider_cdr_reconciliation CHECK (
        (
            call_id IS NULL
            AND organization_id IS NULL
            AND reconciled_at IS NULL
        )
        OR (
            call_id IS NOT NULL
            AND organization_id IS NOT NULL
            AND reconciled_at IS NOT NULL
        )
    )
);

COMMENT ON TABLE provider_cdrs IS
    'Immutable upstream call-detail records reconciled to Leamout-managed calls for wholesale cost accounting.';

CREATE INDEX IF NOT EXISTS idx_provider_cdr_unreconciled
    ON provider_cdrs (provider, started_at)
    WHERE reconciled_at IS NULL;
