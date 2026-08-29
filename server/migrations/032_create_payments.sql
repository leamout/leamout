CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    invoice_id UUID REFERENCES invoices(id) ON DELETE SET NULL,
    provider TEXT NOT NULL,
    provider_payment_id TEXT,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_payments_provider CHECK (
        length(trim(provider)) > 0 AND provider !~ '[[:space:]]'
    ),
    CONSTRAINT chk_payments_provider_payment_id CHECK (
        provider_payment_id IS NULL OR length(trim(provider_payment_id)) > 0
    ),
    CONSTRAINT chk_payments_amount CHECK (amount > 0),
    CONSTRAINT chk_payments_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_payments_status CHECK (
        status IN ('pending', 'succeeded', 'failed', 'refunded', 'partially_refunded')
    ),
    CONSTRAINT chk_payments_paid_at CHECK (
        status NOT IN ('succeeded', 'refunded', 'partially_refunded') OR paid_at IS NOT NULL
    ),
    CONSTRAINT chk_payments_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_payments_provider_payment_id
    ON payments (provider, provider_payment_id)
    WHERE provider_payment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_payments_customer_created
    ON payments (customer_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payments_invoice_created
    ON payments (invoice_id, created_at DESC)
    WHERE invoice_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_payments_status_created
    ON payments (status, created_at DESC);

CREATE TRIGGER set_payments_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
