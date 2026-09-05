CREATE TABLE IF NOT EXISTS provider_cdrs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    carrier_provider_id UUID NOT NULL REFERENCES carrier_providers(id) ON DELETE RESTRICT,
    carrier_connection_id UUID NOT NULL REFERENCES carrier_connections(id) ON DELETE RESTRICT,

    provider_record_id TEXT NOT NULL,
    direction TEXT NOT NULL,
    sip_call_id TEXT,

    call_id UUID REFERENCES calls(id) ON DELETE SET NULL,
    organization_id UUID REFERENCES organizations(id) ON DELETE RESTRICT,
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

-- Keep relationship validation local to this new table instead of altering the
-- older carrier_connections or calls tables to manufacture composite keys.
CREATE FUNCTION validate_provider_cdr_relationships()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM carrier_connections AS cc
        WHERE cc.id = NEW.carrier_connection_id
          AND cc.provider_id = NEW.carrier_provider_id
          AND cc.scope = 'platform'
    ) THEN
        RAISE EXCEPTION 'provider CDR must reference a platform carrier connection for the same provider'
            USING ERRCODE = '23514';
    END IF;

    IF NEW.call_id IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM calls AS c
        WHERE c.id = NEW.call_id
          AND c.organization_id = NEW.organization_id
    ) THEN
        RAISE EXCEPTION 'provider CDR call and organization must match'
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER validate_provider_cdr_relationships
BEFORE INSERT OR UPDATE OF
    carrier_provider_id,
    carrier_connection_id,
    call_id,
    organization_id,
    reconciled_at
ON provider_cdrs
FOR EACH ROW
EXECUTE FUNCTION validate_provider_cdr_relationships();

CREATE INDEX IF NOT EXISTS idx_provider_cdr_unreconciled
    ON provider_cdrs (carrier_provider_id, started_at)
    WHERE reconciled_at IS NULL;

CREATE TABLE IF NOT EXISTS wholesale_charges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_cdr_id UUID NOT NULL UNIQUE REFERENCES provider_cdrs(id) ON DELETE RESTRICT,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    call_id UUID NOT NULL REFERENCES calls(id) ON DELETE RESTRICT,

    amount_micros BIGINT NOT NULL,
    currency TEXT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_wholesale_charge_amount CHECK (
        amount_micros >= 0
    ),
    CONSTRAINT chk_wholesale_charge_currency CHECK (
        currency ~ '^[A-Z]{3}$'
    )
);

COMMENT ON TABLE wholesale_charges IS
    'Provider wholesale call cost attributed to an organization and Leamout call.';

CREATE FUNCTION validate_wholesale_charge_source()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM provider_cdrs AS cdr
        WHERE cdr.id = NEW.provider_cdr_id
          AND cdr.call_id = NEW.call_id
          AND cdr.organization_id = NEW.organization_id
          AND cdr.reconciled_at IS NOT NULL
          AND cdr.cost_micros = NEW.amount_micros
          AND cdr.currency = NEW.currency
          AND cdr.started_at = NEW.occurred_at
    ) THEN
        RAISE EXCEPTION 'wholesale charge must match its reconciled provider CDR'
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER validate_wholesale_charge_source
BEFORE INSERT OR UPDATE OF
    provider_cdr_id,
    organization_id,
    call_id,
    amount_micros,
    currency,
    occurred_at
ON wholesale_charges
FOR EACH ROW
EXECUTE FUNCTION validate_wholesale_charge_source();

CREATE INDEX IF NOT EXISTS idx_wholesale_charges_org_occurred
    ON wholesale_charges (organization_id, occurred_at);
