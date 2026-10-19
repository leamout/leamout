CREATE TABLE IF NOT EXISTS recordings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    call_id UUID NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
    storage_integration_id UUID,
    status TEXT NOT NULL DEFAULT 'recording',
    storage_key TEXT,
    storage_provider TEXT,
    storage_bucket TEXT,
    storage_url TEXT,
    source_path TEXT,
    stopped_at TIMESTAMPTZ,
    upload_attempts INTEGER NOT NULL DEFAULT 0,
    next_upload_at TIMESTAMPTZ,
    upload_error TEXT,
    file_size_bytes BIGINT,
    format TEXT,
    duration_seconds INTEGER,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_recordings_storage_integration_scope
        FOREIGN KEY (storage_integration_id, organization_id)
        REFERENCES storage_integrations(id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT chk_recordings_status CHECK (
        status IN ('recording', 'uploading', 'completed', 'failed', 'deleted')
    ),
    -- All permanent object fields must be absent before a destination is
    -- pinned, or present together for a tenant-owned S3-compatible bucket.
    CONSTRAINT chk_recordings_byos_destination CHECK (
        (
            storage_integration_id IS NULL
            AND storage_key IS NULL
            AND storage_provider IS NULL
            AND storage_bucket IS NULL
        )
        OR
        (
            storage_integration_id IS NOT NULL
            AND storage_key IS NOT NULL
            AND length(btrim(storage_key)) > 0
            AND storage_provider = 's3'
            AND storage_bucket IS NOT NULL
            AND length(btrim(storage_bucket)) > 0
        )
    ),
    CONSTRAINT chk_recordings_completed_byos CHECK (
        status <> 'completed' OR storage_integration_id IS NOT NULL
    ),
    -- Playback URLs are signed on demand; never persist a reusable URL.
    CONSTRAINT chk_recordings_no_storage_url CHECK (
        storage_url IS NULL
    ),
    CONSTRAINT chk_recordings_upload_attempts CHECK (
        upload_attempts >= 0
    ),
    CONSTRAINT chk_recordings_source_path CHECK (
        source_path IS NULL OR (source_path LIKE '/%' AND source_path !~ '(^|/)[.][.](/|$)')
    ),
    CONSTRAINT chk_recordings_duration CHECK (
        duration_seconds IS NULL OR duration_seconds >= 0
    ),
    CONSTRAINT chk_recordings_file_size CHECK (
        file_size_bytes IS NULL OR file_size_bytes >= 0
    ),
    CONSTRAINT chk_recordings_completed_at CHECK (
        completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at
    )
);

CREATE INDEX IF NOT EXISTS idx_recordings_organization_created
    ON recordings (organization_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recordings_call_created
    ON recordings (call_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recordings_storage_integration
    ON recordings (organization_id, storage_integration_id)
    WHERE storage_integration_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_recordings_status
    ON recordings (organization_id, status, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_recordings_call_storage_key
    ON recordings (call_id, storage_key)
    WHERE storage_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_recordings_upload_queue
    ON recordings (next_upload_at, updated_at)
    WHERE status = 'uploading';

CREATE UNIQUE INDEX IF NOT EXISTS uq_recordings_call_source_path
    ON recordings (call_id, source_path)
    WHERE source_path IS NOT NULL;

CREATE TRIGGER set_recordings_updated_at
BEFORE UPDATE ON recordings
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
