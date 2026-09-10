CREATE TABLE IF NOT EXISTS wallet_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    amount_minor BIGINT NOT NULL,
    captured_amount_minor BIGINT,
    operation_type TEXT NOT NULL,
    operation_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    expires_at TIMESTAMPTZ NOT NULL,
    captured_at TIMESTAMPTZ,
    released_at TIMESTAMPTZ,
    expired_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_wallet_reservations_wallet_organization
        FOREIGN KEY (wallet_id, organization_id)
        REFERENCES wallets (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT uq_wallet_reservations_operation UNIQUE (wallet_id, operation_type, operation_id),
    CONSTRAINT chk_wallet_reservations_amount_minor CHECK (amount_minor > 0),
    CONSTRAINT chk_wallet_reservations_captured_amount_minor CHECK (
        captured_amount_minor IS NULL OR (
            captured_amount_minor >= 0
            AND captured_amount_minor <= amount_minor
        )
    ),
    CONSTRAINT chk_wallet_reservations_operation_type CHECK (
        operation_type ~ '^[a-z0-9]+(?:_[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_wallet_reservations_operation_id CHECK (length(btrim(operation_id)) > 0),
    CONSTRAINT chk_wallet_reservations_status CHECK (
        status IN ('active', 'captured', 'released', 'expired')
    ),
    CONSTRAINT chk_wallet_reservations_expiry CHECK (expires_at > created_at),
    CONSTRAINT chk_wallet_reservations_lifecycle CHECK (
        (
            status = 'active'
            AND captured_amount_minor IS NULL
            AND captured_at IS NULL
            AND released_at IS NULL
            AND expired_at IS NULL
        )
        OR (
            status = 'captured'
            AND captured_amount_minor IS NOT NULL
            AND captured_at IS NOT NULL
            AND released_at IS NULL
            AND expired_at IS NULL
        )
        OR (
            status = 'released'
            AND captured_amount_minor IS NULL
            AND captured_at IS NULL
            AND released_at IS NOT NULL
            AND expired_at IS NULL
        )
        OR (
            status = 'expired'
            AND captured_amount_minor IS NULL
            AND captured_at IS NULL
            AND released_at IS NULL
            AND expired_at IS NOT NULL
        )
    )
);

COMMENT ON TABLE wallet_reservations IS
    'Funds in currency minor units committed before Leamout incurs a managed-provider obligation. Active reservations reduce spendable balance.';

CREATE INDEX IF NOT EXISTS idx_wallet_reservations_active
    ON wallet_reservations (wallet_id, expires_at, created_at)
    WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_wallet_reservations_operation
    ON wallet_reservations (organization_id, operation_type, operation_id);

CREATE TRIGGER set_wallet_reservations_updated_at
BEFORE UPDATE ON wallet_reservations
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
