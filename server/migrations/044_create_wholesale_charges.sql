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
