CREATE TABLE IF NOT EXISTS billing_customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_customer_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_billing_customers_customer_provider UNIQUE (customer_id, provider),
    CONSTRAINT uq_billing_customers_provider_customer UNIQUE (provider, provider_customer_id),
    CONSTRAINT chk_billing_customers_provider CHECK (
        length(trim(provider)) > 0 AND provider !~ '[[:space:]]'
    ),
    CONSTRAINT chk_billing_customers_provider_customer_id CHECK (
        length(trim(provider_customer_id)) > 0
    )
);

CREATE INDEX IF NOT EXISTS idx_billing_customers_customer_id
    ON billing_customers (customer_id);
