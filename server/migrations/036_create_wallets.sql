CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE RESTRICT,
    currency TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_wallets_id_organization UNIQUE (id, organization_id),
    CONSTRAINT uq_wallets_payment_terms UNIQUE (id, organization_id, currency),
    CONSTRAINT uq_wallets_organization_currency UNIQUE (organization_id, currency),
    CONSTRAINT chk_wallets_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_wallets_status CHECK (status IN ('active', 'restricted', 'closed'))
);

COMMENT ON TABLE wallets IS
    'Organization-scoped prepaid accounts. Balances are derived from immutable ledger entries and active reservations.';

CREATE INDEX IF NOT EXISTS idx_wallets_organization_status
    ON wallets (organization_id, status, created_at DESC);

CREATE TRIGGER set_wallets_updated_at
BEFORE UPDATE ON wallets
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
