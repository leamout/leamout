CREATE TABLE IF NOT EXISTS checkout_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    wallet_id UUID,
    price_id UUID REFERENCES prices(id) ON DELETE RESTRICT,
    order_type TEXT NOT NULL,
    provider TEXT NOT NULL,
    payment_method TEXT NOT NULL,
    reference TEXT NOT NULL UNIQUE,
    amount BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    next_action TEXT NOT NULL DEFAULT 'wait',
    provider_message TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_checkout_orders_id_organization UNIQUE (id, organization_id),
    CONSTRAINT uq_checkout_orders_payment_terms UNIQUE (id, organization_id, provider, amount, currency),
    CONSTRAINT fk_checkout_orders_wallet_terms
        FOREIGN KEY (wallet_id, organization_id, currency)
        REFERENCES wallets (id, organization_id, currency)
        ON DELETE RESTRICT,
    CONSTRAINT fk_checkout_orders_price_currency
        FOREIGN KEY (price_id, currency)
        REFERENCES prices (id, currency)
        ON DELETE RESTRICT,
    CONSTRAINT chk_checkout_orders_type CHECK (order_type IN ('subscription', 'wallet_topup')),
    CONSTRAINT chk_checkout_orders_target CHECK (
        (order_type = 'subscription' AND price_id IS NOT NULL AND wallet_id IS NULL)
        OR (order_type = 'wallet_topup' AND wallet_id IS NOT NULL AND price_id IS NULL)
    ),
    CONSTRAINT chk_checkout_orders_provider CHECK (provider IN ('stripe', 'paystack')),
    CONSTRAINT chk_checkout_orders_method CHECK (
        (provider = 'stripe' AND payment_method = 'card')
        OR (provider = 'paystack' AND payment_method = 'mobile_money')
    ),
    CONSTRAINT chk_checkout_orders_reference CHECK (
        reference ~ '^[A-Za-z0-9.=-]+$'
    ),
    CONSTRAINT chk_checkout_orders_amount CHECK (amount > 0),
    CONSTRAINT chk_checkout_orders_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_checkout_orders_status CHECK (
        status IN ('pending', 'processing', 'succeeded', 'failed', 'cancelled', 'expired')
    ),
    CONSTRAINT chk_checkout_orders_action CHECK (
        next_action IN ('none', 'wait', 'authorize_mobile_money', 'submit_otp', 'submit_phone', 'unsupported')
    ),
    CONSTRAINT chk_checkout_orders_message CHECK (
        provider_message IS NULL OR length(btrim(provider_message)) > 0
    ),
    CONSTRAINT chk_checkout_orders_expiry CHECK (expires_at > created_at),
    CONSTRAINT chk_checkout_orders_completion CHECK (
        (status IN ('pending', 'processing') AND completed_at IS NULL)
        OR (status IN ('succeeded', 'failed', 'cancelled', 'expired') AND completed_at IS NOT NULL)
    ),
    CONSTRAINT chk_checkout_orders_terminal_action CHECK (
        status IN ('pending', 'processing') OR next_action = 'none'
    ),
    CONSTRAINT chk_checkout_orders_metadata CHECK (jsonb_typeof(metadata) = 'object')
);

COMMENT ON TABLE checkout_orders IS
    'Server-priced intent to collect prepaid money. Provider success is required before subscription activation or wallet credit.';

CREATE INDEX IF NOT EXISTS idx_checkout_orders_organization_created
    ON checkout_orders (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_checkout_orders_pending_expiry
    ON checkout_orders (expires_at, created_at)
    WHERE status IN ('pending', 'processing');

CREATE TRIGGER set_checkout_orders_updated_at
BEFORE UPDATE ON checkout_orders
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
