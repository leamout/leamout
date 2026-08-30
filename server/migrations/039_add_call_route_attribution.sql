ALTER TABLE calls
    ADD COLUMN IF NOT EXISTS carrier_connection_id UUID,
    ADD COLUMN IF NOT EXISTS trunk_id UUID,
    ADD COLUMN IF NOT EXISTS trunk_endpoint_id UUID;

ALTER TABLE trunk_endpoints
    ADD CONSTRAINT uq_trunk_endpoints_id_org UNIQUE (id, organization_id);

ALTER TABLE calls
    ADD CONSTRAINT fk_calls_carrier_connection_org
        FOREIGN KEY (carrier_connection_id, organization_id)
        REFERENCES carrier_connections(id, organization_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_calls_trunk_org
        FOREIGN KEY (trunk_id, organization_id)
        REFERENCES trunks(id, organization_id)
        ON DELETE RESTRICT,
    ADD CONSTRAINT fk_calls_trunk_endpoint_org
        FOREIGN KEY (trunk_endpoint_id, organization_id)
        REFERENCES trunk_endpoints(id, organization_id)
        ON DELETE RESTRICT;

CREATE INDEX IF NOT EXISTS idx_calls_carrier_connection_id
    ON calls (carrier_connection_id)
    WHERE carrier_connection_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_calls_trunk_id
    ON calls (trunk_id)
    WHERE trunk_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_calls_trunk_endpoint_id
    ON calls (trunk_endpoint_id)
    WHERE trunk_endpoint_id IS NOT NULL;
