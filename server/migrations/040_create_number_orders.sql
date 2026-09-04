CREATE TABLE IF NOT EXISTS number_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES carrier_providers(id) ON DELETE RESTRICT,

    provider_inventory_id TEXT NOT NULL,
    provider_product_id TEXT NOT NULL,
    number TEXT NOT NULL,
    country_code CHAR(2) NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',
    provider_order_id TEXT,
    provider_resource_id TEXT,
    phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE SET NULL,

    failed_stage TEXT,
    error_code TEXT,
    error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_number_orders_inventory_id CHECK (
        length(btrim(provider_inventory_id)) > 0
    ),
    CONSTRAINT chk_number_orders_product_id CHECK (
        length(btrim(provider_product_id)) > 0
    ),
    CONSTRAINT chk_number_orders_number CHECK (
        number ~ '^\+[1-9][0-9]{6,14}$'
    ),
    CONSTRAINT chk_number_orders_country_code CHECK (
        country_code ~ '^[A-Z]{2}$'
    ),
    CONSTRAINT chk_number_orders_status CHECK (
        status IN (
            'pending',
            'purchasing',
            'purchased',
            'persisting',
            'configuring',
            'completed',
            'failed'
        )
    ),
    CONSTRAINT chk_number_orders_failed_stage CHECK (
        failed_stage IS NULL
        OR failed_stage IN ('purchasing', 'persisting', 'configuring')
    ),
    CONSTRAINT chk_number_orders_failure_state CHECK (
        (
            status = 'failed'
            AND failed_stage IS NOT NULL
            AND error_message IS NOT NULL
            AND length(btrim(error_message)) > 0
        )
        OR (
            status <> 'failed'
            AND failed_stage IS NULL
            AND error_code IS NULL
            AND error_message IS NULL
        )
    )
);

-- A live/recoverable order owns its provider inventory selection. Completed
-- orders remain historical records without assuming a provider will never
-- recycle an inventory identifier in the future.
CREATE UNIQUE INDEX IF NOT EXISTS uq_number_orders_provider_inventory_open
    ON number_orders (provider_id, provider_inventory_id)
    WHERE status <> 'completed';

CREATE UNIQUE INDEX IF NOT EXISTS uq_number_orders_provider_order
    ON number_orders (provider_id, provider_order_id)
    WHERE provider_order_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_number_orders_provider_resource
    ON number_orders (provider_id, provider_resource_id)
    WHERE provider_resource_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_number_orders_phone_number
    ON number_orders (phone_number_id)
    WHERE phone_number_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_number_orders_organization_created
    ON number_orders (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_number_orders_status_updated
    ON number_orders (status, updated_at)
    WHERE status <> 'completed';

CREATE TRIGGER set_number_orders_updated_at
BEFORE UPDATE ON number_orders
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
