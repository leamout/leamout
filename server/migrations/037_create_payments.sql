CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    provider TEXT NOT NULL,
    provider_payment_id TEXT,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_payments_checkout UNIQUE (checkout_id),
    CONSTRAINT uq_payments_id_organization UNIQUE (id, organization_id),
    CONSTRAINT fk_payments_checkout_terms
        FOREIGN KEY (checkout_id, organization_id, provider, amount_minor, currency)
        REFERENCES checkouts (id, organization_id, provider, amount_minor, currency)
        ON DELETE RESTRICT,
    CONSTRAINT chk_payments_provider CHECK (provider IN ('stripe', 'paystack')),
    CONSTRAINT chk_payments_provider_payment_id CHECK (
        provider_payment_id IS NULL OR length(trim(provider_payment_id)) > 0
    ),
    CONSTRAINT chk_payments_amount_minor CHECK (amount_minor > 0),
    CONSTRAINT chk_payments_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_payments_status CHECK (
        status IN ('pending', 'processing', 'succeeded', 'failed', 'cancelled', 'refunded', 'partially_refunded')
    ),
    CONSTRAINT chk_payments_paid_at CHECK (
        status NOT IN ('succeeded', 'refunded', 'partially_refunded') OR paid_at IS NOT NULL
    ),
    CONSTRAINT chk_payments_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

COMMENT ON TABLE payments IS
    'Provider-independent payment state associated with a checkout.';

CREATE UNIQUE INDEX IF NOT EXISTS uq_payments_provider_payment_id
    ON payments (provider, provider_payment_id)
    WHERE provider_payment_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_payments_organization_created
    ON payments (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_payments_status_created
    ON payments (status, created_at DESC);

CREATE TRIGGER set_payments_updated_at
BEFORE UPDATE ON payments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
