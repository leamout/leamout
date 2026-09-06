CREATE TABLE IF NOT EXISTS provider_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,

    execution_target TEXT NOT NULL,
    carrier_provider_id UUID REFERENCES carrier_providers(id) ON DELETE RESTRICT,

    operation_type TEXT NOT NULL,
    number_order_id UUID,
    phone_number_id UUID,

    idempotency_key TEXT NOT NULL,
    state TEXT NOT NULL DEFAULT 'pending',

    provider_operation_id TEXT,
    provider_resource_id TEXT,

    request JSONB NOT NULL,
    response JSONB,

    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_provider_operations_idempotency UNIQUE (
        organization_id,
        execution_target,
        idempotency_key
    ),
    CONSTRAINT fk_provider_operations_number_order
        FOREIGN KEY (number_order_id, organization_id)
        REFERENCES number_orders (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_provider_operations_phone_number
        FOREIGN KEY (phone_number_id, organization_id)
        REFERENCES phone_numbers (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_provider_operations_execution_target CHECK (
        execution_target IN ('direct', 'transit')
    ),
    CONSTRAINT chk_provider_operations_provider_target CHECK (
        (
            execution_target = 'direct'
            AND carrier_provider_id IS NOT NULL
        )
        OR (
            execution_target = 'transit'
            AND carrier_provider_id IS NULL
        )
    ),
    CONSTRAINT chk_provider_operations_type CHECK (
        operation_type IN ('number_order', 'number_release', 'number_reconcile')
    ),
    CONSTRAINT chk_provider_operations_state CHECK (
        state IN ('pending', 'provider_accepted', 'succeeded', 'failed')
    ),
    CONSTRAINT chk_provider_operations_key CHECK (
        length(btrim(idempotency_key)) > 0
    ),
    CONSTRAINT chk_provider_operations_request CHECK (
        jsonb_typeof(request) = 'object'
    ),
    CONSTRAINT chk_provider_operations_response CHECK (
        response IS NULL OR jsonb_typeof(response) = 'object'
    ),
    CONSTRAINT chk_provider_operations_attempts CHECK (
        attempts >= 0
    ),
    CONSTRAINT chk_provider_operations_target CHECK (
        (
            operation_type = 'number_order'
            AND number_order_id IS NOT NULL
        )
        OR (
            operation_type IN ('number_release', 'number_reconcile')
            AND number_order_id IS NULL
            AND phone_number_id IS NOT NULL
            AND provider_resource_id IS NOT NULL
            AND length(btrim(provider_resource_id)) > 0
        )
    ),
    CONSTRAINT chk_provider_operations_completion CHECK (
        (
            state IN ('pending', 'provider_accepted')
            AND completed_at IS NULL
            AND next_attempt_at IS NOT NULL
        )
        OR (
            state IN ('succeeded', 'failed')
            AND completed_at IS NOT NULL
            AND next_attempt_at IS NULL
        )
    ),
    CONSTRAINT chk_provider_operations_failure CHECK (
        state <> 'failed'
        OR (
            last_error IS NOT NULL
            AND length(btrim(last_error)) > 0
        )
    ),
    CONSTRAINT chk_provider_operations_succeeded_order CHECK (
        state <> 'succeeded'
        OR operation_type <> 'number_order'
        OR (
            provider_resource_id IS NOT NULL
            AND length(btrim(provider_resource_id)) > 0
            AND phone_number_id IS NOT NULL
        )
    )
);

COMMENT ON TABLE provider_operations IS
    'Internal durable journal for Managed Carrier side effects. execution_target chooses local provider execution or Leamout Transit without coupling that choice to hosting mode.';

CREATE UNIQUE INDEX IF NOT EXISTS uq_provider_operations_number_order
    ON provider_operations (number_order_id)
    WHERE operation_type = 'number_order';

CREATE UNIQUE INDEX IF NOT EXISTS uq_provider_operations_direct_provider_operation
    ON provider_operations (carrier_provider_id, provider_operation_id)
    WHERE execution_target = 'direct'
      AND provider_operation_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_provider_operations_transit_operation
    ON provider_operations (provider_operation_id)
    WHERE execution_target = 'transit'
      AND provider_operation_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_provider_operations_retry
    ON provider_operations (next_attempt_at, created_at)
    WHERE state IN ('pending', 'provider_accepted')
      AND next_attempt_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_provider_operations_direct_resource
    ON provider_operations (carrier_provider_id, provider_resource_id)
    WHERE execution_target = 'direct'
      AND provider_resource_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_provider_operations_transit_resource
    ON provider_operations (provider_resource_id)
    WHERE execution_target = 'transit'
      AND provider_resource_id IS NOT NULL;

CREATE TRIGGER set_provider_operations_updated_at
BEFORE UPDATE ON provider_operations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
