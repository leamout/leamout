CREATE TABLE IF NOT EXISTS invoice_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    meter_id UUID REFERENCES meters(id) ON DELETE SET NULL,
    usage_rate_id UUID REFERENCES usage_rates(id) ON DELETE SET NULL,
    type TEXT NOT NULL,
    description TEXT NOT NULL,
    quantity BIGINT NOT NULL DEFAULT 1,
    unit_amount_micros BIGINT,
    amount BIGINT NOT NULL,
    period_start TIMESTAMPTZ,
    period_end TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_invoice_items_type CHECK (type IN ('fixed', 'usage')),
    CONSTRAINT chk_invoice_items_description CHECK (length(trim(description)) > 0),
    CONSTRAINT chk_invoice_items_quantity CHECK (quantity > 0),
    CONSTRAINT chk_invoice_items_unit_amount CHECK (
        unit_amount_micros IS NULL OR unit_amount_micros >= 0
    ),
    CONSTRAINT chk_invoice_items_amount CHECK (amount >= 0),
    CONSTRAINT chk_invoice_items_usage CHECK (
        type <> 'usage' OR (meter_id IS NOT NULL AND unit_amount_micros IS NOT NULL)
    ),
    CONSTRAINT chk_invoice_items_period CHECK (
        period_start IS NULL OR period_end IS NULL OR period_end >= period_start
    ),
    CONSTRAINT chk_invoice_items_metadata_object CHECK (jsonb_typeof(metadata) = 'object')
);

CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice_created
    ON invoice_items (invoice_id, created_at);

CREATE INDEX IF NOT EXISTS idx_invoice_items_meter
    ON invoice_items (meter_id, period_start, period_end)
    WHERE meter_id IS NOT NULL;
