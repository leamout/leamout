CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    subscription_id UUID REFERENCES subscriptions(id) ON DELETE SET NULL,
    invoice_number TEXT NOT NULL UNIQUE,
    currency TEXT NOT NULL,
    subtotal BIGINT NOT NULL,
    tax BIGINT NOT NULL DEFAULT 0,
    total BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    issued_at TIMESTAMPTZ,
    due_at TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_invoices_invoice_number CHECK (length(trim(invoice_number)) > 0),
    CONSTRAINT chk_invoices_currency CHECK (currency ~ '^[A-Z]{3}$'),
    CONSTRAINT chk_invoices_amounts CHECK (subtotal >= 0 AND tax >= 0 AND total >= 0),
    CONSTRAINT chk_invoices_status CHECK (status IN ('draft', 'open', 'paid', 'void', 'overdue')),
    CONSTRAINT chk_invoices_due_at CHECK (
        due_at IS NULL OR issued_at IS NULL OR due_at >= issued_at
    ),
    CONSTRAINT chk_invoices_paid_at CHECK (
        status <> 'paid' OR paid_at IS NOT NULL
    ),
    CONSTRAINT chk_invoices_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX IF NOT EXISTS idx_invoices_customer_created
    ON invoices (customer_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_invoices_subscription_created
    ON invoices (subscription_id, created_at DESC)
    WHERE subscription_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_invoices_status_due
    ON invoices (status, due_at)
    WHERE status IN ('open', 'overdue');

CREATE TRIGGER set_invoices_updated_at
BEFORE UPDATE ON invoices
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
