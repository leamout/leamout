ALTER TABLE phone_numbers DROP CONSTRAINT chk_phone_numbers_status;
ALTER TABLE phone_numbers ADD CONSTRAINT chk_phone_numbers_status CHECK (
    status IN ('provisioning', 'active', 'disabled', 'porting', 'releasing', 'released')
);

CREATE TABLE provider_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    carrier_provider_id UUID NOT NULL REFERENCES carrier_providers(id) ON DELETE RESTRICT,
    operation_type TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'pending',
    provider_operation_id TEXT,
    provider_resource_id TEXT,
    phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE SET NULL,
    request JSONB NOT NULL,
    response JSONB,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    next_attempt_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_provider_operations_idempotency UNIQUE (carrier_provider_id, idempotency_key),
    CONSTRAINT chk_provider_operations_type CHECK (operation_type IN ('number_order', 'number_release', 'number_reconcile')),
    CONSTRAINT chk_provider_operations_state CHECK (state IN ('pending', 'provider_accepted', 'succeeded', 'failed')),
    CONSTRAINT chk_provider_operations_key CHECK (length(btrim(idempotency_key)) > 0),
    CONSTRAINT chk_provider_operations_request CHECK (jsonb_typeof(request) = 'object'),
    CONSTRAINT chk_provider_operations_response CHECK (response IS NULL OR jsonb_typeof(response) = 'object'),
    CONSTRAINT chk_provider_operations_attempts CHECK (attempts >= 0)
);

CREATE INDEX idx_provider_operations_retry
    ON provider_operations (next_attempt_at, created_at)
    WHERE state IN ('pending', 'provider_accepted', 'failed');

CREATE INDEX idx_provider_operations_resource
    ON provider_operations (carrier_provider_id, provider_resource_id)
    WHERE provider_resource_id IS NOT NULL;

CREATE TRIGGER set_provider_operations_updated_at
BEFORE UPDATE ON provider_operations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
