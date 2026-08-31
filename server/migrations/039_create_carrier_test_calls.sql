CREATE TABLE carrier_test_calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    carrier_connection_id UUID NOT NULL,
    trunk_id UUID NOT NULL,
    trunk_endpoint_id UUID,
    actor_type TEXT NOT NULL,
    actor_id UUID NOT NULL,
    from_number TEXT NOT NULL,
    to_number TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    sip_call_id TEXT,
    response_code TEXT,
    error_message TEXT,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    answered_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    CONSTRAINT fk_carrier_test_calls_connection
        FOREIGN KEY (carrier_connection_id, organization_id)
        REFERENCES carrier_connections (id, organization_id),
    CONSTRAINT fk_carrier_test_calls_trunk
        FOREIGN KEY (trunk_id, organization_id)
        REFERENCES trunks (id, organization_id),
    CONSTRAINT fk_carrier_test_calls_endpoint
        FOREIGN KEY (trunk_endpoint_id, organization_id)
        REFERENCES trunk_endpoints (id, organization_id),
    CONSTRAINT chk_carrier_test_call_actor CHECK (actor_type IN ('user', 'organization_token')),
    CONSTRAINT chk_carrier_test_call_status CHECK (status IN ('pending', 'answered', 'completed', 'failed', 'timed_out'))
);

CREATE INDEX idx_carrier_test_calls_connection_time
    ON carrier_test_calls (organization_id, carrier_connection_id, started_at DESC);
