CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email CITEXT,
    status TEXT NOT NULL DEFAULT 'active',
    external_reference TEXT,
    billing_provider TEXT,
    provider_customer_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_customers_name CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_customers_status CHECK (status IN ('active', 'suspended', 'closed')),
    CONSTRAINT chk_customers_external_reference CHECK (
        external_reference IS NULL OR length(trim(external_reference)) > 0
    ),
    CONSTRAINT chk_customers_billing_provider CHECK (
        billing_provider IS NULL OR (length(trim(billing_provider)) > 0 AND billing_provider !~ '[[:space:]]')
    ),
    CONSTRAINT chk_customers_provider_customer_id CHECK (
        provider_customer_id IS NULL OR length(trim(provider_customer_id)) > 0
    ),
    CONSTRAINT chk_customers_billing_pair CHECK (
        (billing_provider IS NULL AND provider_customer_id IS NULL)
        OR (billing_provider IS NOT NULL AND provider_customer_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_email
    ON customers (email)
    WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_external_reference
    ON customers (external_reference)
    WHERE external_reference IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_provider_customer
    ON customers (billing_provider, provider_customer_id)
    WHERE billing_provider IS NOT NULL AND provider_customer_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_customers_status_created
    ON customers (status, created_at DESC);

CREATE TRIGGER set_customers_updated_at
BEFORE UPDATE ON customers
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
