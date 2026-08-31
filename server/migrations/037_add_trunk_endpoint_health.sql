ALTER TABLE trunk_endpoints
    ADD COLUMN health_status TEXT NOT NULL DEFAULT 'unknown',
    ADD COLUMN consecutive_failures INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN last_checked_at TIMESTAMPTZ,
    ADD COLUMN last_response_code INTEGER,
    ADD COLUMN last_latency_ms INTEGER,
    ADD COLUMN last_error TEXT,
    ADD COLUMN cooldown_until TIMESTAMPTZ,
    ADD CONSTRAINT chk_trunk_endpoint_health_status
        CHECK (health_status IN ('unknown', 'healthy', 'unhealthy')),
    ADD CONSTRAINT chk_trunk_endpoint_consecutive_failures
        CHECK (consecutive_failures >= 0),
    ADD CONSTRAINT chk_trunk_endpoint_response_code
        CHECK (last_response_code IS NULL OR last_response_code BETWEEN 100 AND 699),
    ADD CONSTRAINT chk_trunk_endpoint_latency
        CHECK (last_latency_ms IS NULL OR last_latency_ms >= 0);

CREATE INDEX idx_trunk_endpoints_health_probe
    ON trunk_endpoints (health_status, cooldown_until, last_checked_at)
    WHERE enabled = true;
