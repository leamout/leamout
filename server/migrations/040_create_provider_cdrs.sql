CREATE TABLE IF NOT EXISTS provider_cdrs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carrier_provider_id UUID NOT NULL REFERENCES carrier_providers(id) ON DELETE RESTRICT,
    carrier_connection_id UUID NOT NULL,

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

    CONSTRAINT uq_provider_cdr UNIQUE (
        carrier_provider_id,
        direction,
        provider_record_id
    ),
    CONSTRAINT uq_provider_cdr_id_call_organization UNIQUE (
        id,
        call_id,
        organization_id
    ),
    CONSTRAINT fk_provider_cdr_connection_provider
        FOREIGN KEY (carrier_connection_id, carrier_provider_id)
        REFERENCES carrier_connections (id, provider_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_provider_cdr_call_organization
        FOREIGN KEY (call_id, organization_id)
        REFERENCES calls (id, organization_id)
        ON DELETE RESTRICT,
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
    ON provider_cdrs (carrier_provider_id, started_at)
    WHERE reconciled_at IS NULL;

CREATE TABLE IF NOT EXISTS wholesale_charges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_cdr_id UUID NOT NULL,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    call_id UUID NOT NULL,

    amount_micros BIGINT NOT NULL,
    currency TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wholesale_charges_provider_cdr UNIQUE (provider_cdr_id),
    CONSTRAINT fk_wholesale_charge_provider_cdr
        FOREIGN KEY (provider_cdr_id, call_id, organization_id)
        REFERENCES provider_cdrs (id, call_id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_wholesale_charge_call_organization
        FOREIGN KEY (call_id, organization_id)
        REFERENCES calls (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_wholesale_charge_amount CHECK (
        amount_micros >= 0
    ),
    CONSTRAINT chk_wholesale_charge_currency CHECK (
        currency ~ '^[A-Z]{3}$'
    )
);

COMMENT ON TABLE wholesale_charges IS
    'Provider wholesale call cost attributed to an organization and Leamout call.';

CREATE INDEX IF NOT EXISTS idx_wholesale_charges_org_occurred
    ON wholesale_charges (organization_id, occurred_at);
