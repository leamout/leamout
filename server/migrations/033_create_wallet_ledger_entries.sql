CREATE TABLE IF NOT EXISTS wallet_ledger_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL,
    organization_id UUID NOT NULL,
    entry_type TEXT NOT NULL,
    amount BIGINT NOT NULL,
    source_type TEXT NOT NULL,
    source_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_wallet_ledger_wallet_organization
        FOREIGN KEY (wallet_id, organization_id)
        REFERENCES wallets (id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT uq_wallet_ledger_idempotency UNIQUE (wallet_id, idempotency_key),
    CONSTRAINT chk_wallet_ledger_type CHECK (
        entry_type IN ('topup', 'capture', 'refund', 'chargeback', 'adjustment_credit', 'adjustment_debit')
    ),
    CONSTRAINT chk_wallet_ledger_amount CHECK (
        (entry_type IN ('topup', 'refund', 'adjustment_credit') AND amount > 0)
        OR (entry_type IN ('capture', 'chargeback', 'adjustment_debit') AND amount < 0)
    ),
    CONSTRAINT chk_wallet_ledger_source_type CHECK (
        source_type ~ '^[a-z0-9]+(?:_[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_wallet_ledger_source_id CHECK (length(btrim(source_id)) > 0),
    CONSTRAINT chk_wallet_ledger_idempotency_key CHECK (length(btrim(idempotency_key)) > 0),
    CONSTRAINT chk_wallet_ledger_metadata CHECK (jsonb_typeof(metadata) = 'object')
);

COMMENT ON TABLE wallet_ledger_entries IS
    'Immutable posted monetary movements. Corrections are new compensating entries; rows are never updated or deleted.';

CREATE INDEX IF NOT EXISTS idx_wallet_ledger_wallet_occurred
    ON wallet_ledger_entries (wallet_id, occurred_at, id);

CREATE INDEX IF NOT EXISTS idx_wallet_ledger_source
    ON wallet_ledger_entries (source_type, source_id);

CREATE OR REPLACE FUNCTION reject_wallet_ledger_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'wallet ledger entries are immutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER reject_wallet_ledger_update
BEFORE UPDATE ON wallet_ledger_entries
FOR EACH ROW
EXECUTE FUNCTION reject_wallet_ledger_mutation();

CREATE TRIGGER reject_wallet_ledger_delete
BEFORE DELETE ON wallet_ledger_entries
FOR EACH ROW
EXECUTE FUNCTION reject_wallet_ledger_mutation();
