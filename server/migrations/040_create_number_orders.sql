CREATE TABLE IF NOT EXISTS number_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES carrier_providers(id) ON DELETE RESTRICT,

    -- Immutable selection context. These identifiers stay internal and are not
    -- exposed through the customer-facing number-order response.
    provider_inventory_id TEXT NOT NULL,
    provider_product_id TEXT NOT NULL,
    number TEXT NOT NULL,
    country_code CHAR(2) NOT NULL,

    -- Customer/business state only. Provider execution state belongs in
    -- provider_operations.
    status TEXT NOT NULL DEFAULT 'pending',
    phone_number_id UUID REFERENCES phone_numbers(id) ON DELETE SET NULL,

    error_code TEXT,
    error_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_number_orders_id_org_provider UNIQUE (
        id,
        organization_id,
        provider_id
    ),
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
        status IN ('pending', 'processing', 'completed', 'failed')
    ),
    CONSTRAINT chk_number_orders_state CHECK (
        (
            status IN ('pending', 'processing')
            AND phone_number_id IS NULL
            AND error_code IS NULL
            AND error_message IS NULL
        )
        OR (
            status = 'completed'
            AND phone_number_id IS NOT NULL
            AND error_code IS NULL
            AND error_message IS NULL
        )
        OR (
            status = 'failed'
            AND phone_number_id IS NULL
            AND error_message IS NOT NULL
            AND length(btrim(error_message)) > 0
        )
    )
);

COMMENT ON TABLE number_orders IS
    'Customer-visible managed-number acquisition intent. Provider execution and retry state lives in provider_operations.';

-- Only one live order may own a provider inventory selection. Completed or
-- terminally failed history does not permanently reserve recycled inventory.
CREATE UNIQUE INDEX IF NOT EXISTS uq_number_orders_provider_inventory_open
    ON number_orders (provider_id, provider_inventory_id)
    WHERE status IN ('pending', 'processing');

-- An E.164 number cannot be acquired concurrently through two provider
-- selections while an order is still live.
CREATE UNIQUE INDEX IF NOT EXISTS uq_number_orders_number_open
    ON number_orders (number)
    WHERE status IN ('pending', 'processing');

CREATE UNIQUE INDEX IF NOT EXISTS uq_number_orders_phone_number
    ON number_orders (phone_number_id)
    WHERE phone_number_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_number_orders_organization_created
    ON number_orders (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_number_orders_active_updated
    ON number_orders (status, updated_at)
    WHERE status IN ('pending', 'processing');

CREATE TRIGGER set_number_orders_updated_at
BEFORE UPDATE ON number_orders
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
