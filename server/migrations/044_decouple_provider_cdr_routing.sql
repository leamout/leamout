ALTER TABLE provider_cdrs
    ADD COLUMN provider TEXT;

UPDATE provider_cdrs AS cdr
SET provider = cp.slug
FROM carrier_providers AS cp
WHERE cp.id = cdr.carrier_provider_id;

ALTER TABLE provider_cdrs
    ALTER COLUMN provider SET NOT NULL;

DROP INDEX IF EXISTS idx_provider_cdr_unreconciled;
ALTER TABLE provider_cdrs DROP CONSTRAINT IF EXISTS uq_provider_cdr;
ALTER TABLE provider_cdrs DROP CONSTRAINT IF EXISTS fk_provider_cdr_connection_provider;

ALTER TABLE provider_cdrs
    ADD CONSTRAINT uq_provider_cdr_source UNIQUE (
        provider,
        direction,
        provider_record_id
    ),
    ADD CONSTRAINT fk_provider_cdr_connection
        FOREIGN KEY (carrier_connection_id)
        REFERENCES carrier_connections (id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT chk_provider_cdr_provider CHECK (
        provider ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    );

ALTER TABLE provider_cdrs
    DROP COLUMN carrier_provider_id;

CREATE INDEX idx_provider_cdr_unreconciled
    ON provider_cdrs (provider, started_at)
    WHERE reconciled_at IS NULL;

CREATE TABLE provider_cdr_routes (
    provider TEXT NOT NULL,
    direction TEXT NOT NULL,
    carrier_connection_id UUID NOT NULL REFERENCES carrier_connections(id) ON DELETE RESTRICT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (provider, direction),
    CONSTRAINT uq_provider_cdr_routes_connection UNIQUE (carrier_connection_id, direction),
    CONSTRAINT chk_provider_cdr_routes_provider CHECK (
        provider ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_provider_cdr_routes_direction CHECK (
        direction IN ('termination', 'origination')
    ),
    CONSTRAINT chk_provider_cdr_routes_status CHECK (
        status IN ('active', 'disabled')
    )
);

CREATE TRIGGER set_provider_cdr_routes_updated_at
BEFORE UPDATE ON provider_cdr_routes
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

ALTER TABLE provider_cdr_pages
    ADD COLUMN process_attempts INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN next_process_at TIMESTAMPTZ DEFAULT now(),
    ADD COLUMN last_process_error TEXT,
    ADD COLUMN processed_at TIMESTAMPTZ,
    ADD CONSTRAINT chk_provider_cdr_pages_process_attempts CHECK (process_attempts >= 0),
    ADD CONSTRAINT chk_provider_cdr_pages_process_error CHECK (
        last_process_error IS NULL OR length(btrim(last_process_error)) > 0
    ),
    ADD CONSTRAINT chk_provider_cdr_pages_process_state CHECK (
        (processed_at IS NULL AND next_process_at IS NOT NULL)
        OR (processed_at IS NOT NULL AND next_process_at IS NULL)
    );

CREATE INDEX idx_provider_cdr_pages_processing
    ON provider_cdr_pages (next_process_at, received_at)
    WHERE processed_at IS NULL;
