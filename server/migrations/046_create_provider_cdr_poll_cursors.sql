CREATE TABLE IF NOT EXISTS provider_cdr_poll_cursors (
    provider TEXT NOT NULL,
    direction TEXT NOT NULL,
    window_date DATE NOT NULL,
    page INTEGER NOT NULL DEFAULT 1,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error TEXT,
    last_success_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (provider, direction),
    CONSTRAINT chk_provider_cdr_poll_cursor_provider CHECK (
        provider ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_provider_cdr_poll_cursor_direction CHECK (
        direction IN ('termination', 'origination')
    ),
    CONSTRAINT chk_provider_cdr_poll_cursor_page CHECK (page > 0),
    CONSTRAINT chk_provider_cdr_poll_cursor_attempts CHECK (attempt_count >= 0),
    CONSTRAINT chk_provider_cdr_poll_cursor_error CHECK (
        last_error IS NULL OR length(btrim(last_error)) > 0
    )
);

CREATE INDEX IF NOT EXISTS idx_provider_cdr_poll_cursors_ready
    ON provider_cdr_poll_cursors (next_attempt_at);

CREATE TRIGGER set_provider_cdr_poll_cursors_updated_at
BEFORE UPDATE ON provider_cdr_poll_cursors
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
