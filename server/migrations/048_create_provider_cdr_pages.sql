CREATE TABLE IF NOT EXISTS provider_cdr_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    direction TEXT NOT NULL,
    window_date DATE NOT NULL,
    page INTEGER NOT NULL,
    record_count INTEGER NOT NULL,
    payload_sha256 TEXT NOT NULL,
    raw JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    process_attempts INTEGER NOT NULL DEFAULT 0,
    next_process_at TIMESTAMPTZ DEFAULT now(),
    last_process_error TEXT,
    processed_at TIMESTAMPTZ,

    CONSTRAINT uq_provider_cdr_pages_payload UNIQUE (
        provider,
        direction,
        window_date,
        page,
        payload_sha256
    ),
    CONSTRAINT chk_provider_cdr_pages_provider CHECK (
        provider ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_provider_cdr_pages_direction CHECK (
        direction IN ('termination', 'origination')
    ),
    CONSTRAINT chk_provider_cdr_pages_page CHECK (page > 0),
    CONSTRAINT chk_provider_cdr_pages_count CHECK (record_count >= 0),
    CONSTRAINT chk_provider_cdr_pages_hash CHECK (
        payload_sha256 ~ '^[0-9a-f]{64}$'
    ),
    CONSTRAINT chk_provider_cdr_pages_raw CHECK (
        jsonb_typeof(raw) IN ('array', 'object')
    ),
    CONSTRAINT chk_provider_cdr_pages_process_attempts CHECK (
        process_attempts >= 0
    ),
    CONSTRAINT chk_provider_cdr_pages_process_error CHECK (
        last_process_error IS NULL OR length(btrim(last_process_error)) > 0
    ),
    CONSTRAINT chk_provider_cdr_pages_process_state CHECK (
        (processed_at IS NULL AND next_process_at IS NOT NULL)
        OR (processed_at IS NOT NULL AND next_process_at IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_provider_cdr_pages_unprocessed
    ON provider_cdr_pages (provider, direction, window_date, page, received_at);

CREATE INDEX IF NOT EXISTS idx_provider_cdr_pages_processing
    ON provider_cdr_pages (next_process_at, received_at)
    WHERE processed_at IS NULL;
