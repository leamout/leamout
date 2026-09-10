CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    checkout_id UUID NOT NULL,
    payment_id UUID NOT NULL,
    wallet_id UUID,
    price_id UUID,
    order_type TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,
    completed_at TIMESTAMPTZ NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_orders_checkout UNIQUE (checkout_id),
    CONSTRAINT uq_orders_payment UNIQUE (payment_id),
    CONSTRAINT uq_orders_id_organization UNIQUE (id, organization_id),
    CONSTRAINT fk_orders_checkout
        FOREIGN KEY (checkout_id, organization_id)
        REFERENCES checkouts (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_orders_payment
        FOREIGN KEY (payment_id, organization_id)
        REFERENCES payments (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT fk_orders_wallet_terms
        FOREIGN KEY (wallet_id, organization_id, currency)
        REFERENCES wallets (id, organization_id, currency)
        ON DELETE RESTRICT,
    CONSTRAINT fk_orders_price_currency
        FOREIGN KEY (price_id, currency)
        REFERENCES prices (id, currency)
        ON DELETE RESTRICT,
    CONSTRAINT chk_orders_type CHECK (order_type IN ('subscription', 'wallet_topup')),
    CONSTRAINT chk_orders_target CHECK (
        (order_type = 'subscription' AND price_id IS NOT NULL AND wallet_id IS NULL)
        OR (order_type = 'wallet_topup' AND wallet_id IS NOT NULL AND price_id IS NULL)
    ),
    CONSTRAINT chk_orders_amount_minor CHECK (amount_minor > 0),
    CONSTRAINT chk_orders_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_orders_metadata CHECK (jsonb_typeof(metadata) = 'object')
);

COMMENT ON TABLE orders IS
    'Durable records of completed Leamout purchases created only after payment is confirmed.';

CREATE INDEX IF NOT EXISTS idx_orders_organization_created
    ON orders (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_orders_price
    ON orders (price_id)
    WHERE price_id IS NOT NULL;
