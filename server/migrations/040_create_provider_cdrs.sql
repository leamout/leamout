ALTER TABLE carrier_connections
    ADD CONSTRAINT uq_carrier_connections_id_provider UNIQUE (id, provider_id);

ALTER TABLE calls
    ADD CONSTRAINT uq_calls_id_organization UNIQUE (id, organization_id);

CREATE TABLE IF NOT EXISTS provider_cdrs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carrier_provider_id UUID NOT NULL REFERENCES carrier_providers(id) ON DELETE RESTRICT,
    carrier_connection_id UUID NOT NULL,
    provider_record_id TEXT NOT NULL,
    direction TEXT NOT NULL,
    sip_call_id TEXT,
    call_id UUID REFERENCES calls(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE RESTRICT,
    started_at TIMESTAMPTZ NOT NULL,
    duration_seconds BIGINT NOT NULL,
    currency TEXT NOT NULL,
    cost_micros BIGINT NOT NULL,
    raw JSONB NOT NULL,
    reconciled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_provider_cdr UNIQUE (carrier_provider_id, direction, provider_record_id),
    CONSTRAINT fk_provider_cdr_connection_provider
        FOREIGN KEY (carrier_connection_id, carrier_provider_id)
        REFERENCES carrier_connections (id, provider_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_provider_cdr_direction CHECK (direction IN ('termination', 'origination')),
    CONSTRAINT chk_provider_cdr_record CHECK (length(btrim(provider_record_id)) > 0),
    CONSTRAINT chk_provider_cdr_duration CHECK (duration_seconds >= 0),
    CONSTRAINT chk_provider_cdr_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_provider_cdr_cost CHECK (cost_micros >= 0),
    CONSTRAINT chk_provider_cdr_raw CHECK (jsonb_typeof(raw) = 'object'),
    CONSTRAINT chk_provider_cdr_reconciliation CHECK (
        (call_id IS NULL AND organization_id IS NULL AND reconciled_at IS NULL)
        OR (call_id IS NOT NULL AND organization_id IS NOT NULL AND reconciled_at IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_provider_cdr_unreconciled
    ON provider_cdrs (carrier_provider_id, started_at)
    WHERE reconciled_at IS NULL;

CREATE TABLE IF NOT EXISTS wholesale_charges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_cdr_id UUID NOT NULL UNIQUE REFERENCES provider_cdrs(id) ON DELETE RESTRICT,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    call_id UUID NOT NULL,
    amount_micros BIGINT NOT NULL,
    currency TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_wholesale_charge_amount CHECK (amount_micros >= 0),
    CONSTRAINT chk_wholesale_charge_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT fk_wholesale_charge_call_organization
        FOREIGN KEY (call_id, organization_id)
        REFERENCES calls (id, organization_id)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_wholesale_charges_org_occurred
    ON wholesale_charges (organization_id, occurred_at);
