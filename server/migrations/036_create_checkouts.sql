CREATE TABLE IF NOT EXISTS checkouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    wallet_id UUID,
    price_id UUID REFERENCES prices(id) ON DELETE RESTRICT,
    checkout_type TEXT NOT NULL,
    provider TEXT,
    payment_method TEXT,
    reference TEXT NOT NULL UNIQUE,
    amount_minor BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    next_action TEXT NOT NULL DEFAULT 'wait',
    provider_message TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    completed_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_checkouts_id_organization UNIQUE (id, organization_id),
    CONSTRAINT fk_checkouts_wallet_terms FOREIGN KEY (wallet_id, organization_id, currency)
        REFERENCES wallets (id, organization_id, currency) ON DELETE RESTRICT,
    CONSTRAINT fk_checkouts_price_currency FOREIGN KEY (price_id, currency)
        REFERENCES prices (id, currency) ON DELETE RESTRICT,
    CONSTRAINT chk_checkouts_type CHECK (checkout_type IN ('subscription', 'wallet_topup')),
    CONSTRAINT chk_checkouts_target CHECK (
        (checkout_type = 'subscription' AND price_id IS NOT NULL AND wallet_id IS NULL)
        OR (checkout_type = 'wallet_topup' AND wallet_id IS NOT NULL AND price_id IS NULL)
    ),
    CONSTRAINT chk_checkouts_payment_binding CHECK (
        (provider IS NULL AND payment_method IS NULL)
        OR (provider IS NOT NULL AND payment_method IS NOT NULL AND (
            (provider = 'stripe' AND payment_method = 'card')
            OR (provider = 'paystack' AND payment_method = 'mobile_money')
        ))
    ),
    CONSTRAINT chk_checkouts_reference CHECK (reference ~ '^[A-Za-z0-9.=-]+$'),
    CONSTRAINT chk_checkouts_amount_minor CHECK (amount_minor > 0),
    CONSTRAINT chk_checkouts_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_checkouts_status CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'cancelled', 'expired')),
    CONSTRAINT chk_checkouts_action CHECK (next_action IN ('none', 'wait', 'authorize_mobile_money', 'submit_otp', 'submit_phone', 'unsupported')),
    CONSTRAINT chk_checkouts_message CHECK (provider_message IS NULL OR length(btrim(provider_message)) > 0),
    CONSTRAINT chk_checkouts_expiry CHECK (expires_at > created_at),
    CONSTRAINT chk_checkouts_completion CHECK (
        (status IN ('pending', 'processing') AND completed_at IS NULL)
        OR (status IN ('succeeded', 'failed', 'cancelled', 'expired') AND completed_at IS NOT NULL)
    ),
    CONSTRAINT chk_checkouts_terminal_action CHECK (status IN ('pending', 'processing') OR next_action = 'none'),
    CONSTRAINT chk_checkouts_succeeded_payment CHECK (status <> 'succeeded' OR (provider IS NOT NULL AND payment_method IS NOT NULL)),
    CONSTRAINT chk_checkouts_metadata CHECK (jsonb_typeof(metadata) = 'object')
);

COMMENT ON TABLE checkouts IS
    'Provider-neutral, short-lived commercial purchase sessions. Payment selection is bound at confirmation and successful settlement is fulfilled by Checkout.';

CREATE INDEX IF NOT EXISTS idx_checkouts_organization_created ON checkouts (organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_checkouts_pending_expiry ON checkouts (expires_at, created_at)
    WHERE status IN ('pending', 'processing');

CREATE TRIGGER set_checkouts_updated_at BEFORE UPDATE ON checkouts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
