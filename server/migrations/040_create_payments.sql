CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_order_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    provider TEXT NOT NULL,
    provider_payment_id TEXT,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_payments_checkout_order UNIQUE (checkout_order_id),
    CONSTRAINT uq_payments_id_organization UNIQUE (id, organization_id),
    CONSTRAINT fk_payments_checkout_terms
        FOREIGN KEY (checkout_order_id, organization_id, provider, amount, currency)
        REFERENCES checkout_orders (id, organization_id, provider, amount, currency)
        ON DELETE RESTRICT,
    CONSTRAINT chk_payments_provider CHECK (provider IN ('stripe', 'paystack')),
    CONSTRAINT chk_payments_provider_payment_id CHECK (
        provider_payment_id IS NULL OR length(trim(provider_payment_id)) > 0
    ),
    CONSTRAINT chk_payments_amount CHECK (amount > 0),
    CONSTRAINT chk_payments_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_payments_status CHECK (
        status IN ('pending', 'processing', 'succeeded', 'failed', 'cancelled', 'refunded', 'partially_refunded')
    ),
    CONSTRAINT chk_payments_paid_at CHECK (
        status NOT IN ('succeeded', 'refunded', 'partially_refunded') OR paid_at IS NOT NULL
    ),
    CONSTRAINT chk_payments_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

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

CREATE TABLE IF NOT EXISTS payment_provider_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    provider TEXT NOT NULL,
    provider_event_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload_sha256 TEXT NOT NULL,
    payload JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,

    CONSTRAINT fk_payment_provider_events_payment_organization
        FOREIGN KEY (payment_id, organization_id)
        REFERENCES payments (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT uq_payment_provider_events_identity UNIQUE (provider, provider_event_id),
    CONSTRAINT chk_payment_provider_events_provider CHECK (provider IN ('stripe', 'paystack')),
    CONSTRAINT chk_payment_provider_events_id CHECK (length(btrim(provider_event_id)) > 0),
    CONSTRAINT chk_payment_provider_events_type CHECK (length(btrim(event_type)) > 0),
    CONSTRAINT chk_payment_provider_events_hash CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT chk_payment_provider_events_payload CHECK (jsonb_typeof(payload) = 'object')
);

COMMENT ON TABLE payment_provider_events IS
    'Authenticated provider events retained once by provider identity so duplicate webhooks cannot repeat commercial effects.';

CREATE INDEX IF NOT EXISTS idx_payment_provider_events_unprocessed
    ON payment_provider_events (received_at, id)
    WHERE processed_at IS NULL;
