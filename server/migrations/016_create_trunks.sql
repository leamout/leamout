CREATE TABLE IF NOT EXISTS trunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    carrier_connection_id UUID NOT NULL,

    name TEXT NOT NULL,
    direction TEXT NOT NULL DEFAULT 'bidirectional',
    status TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_trunks_connection_org
        FOREIGN KEY (carrier_connection_id, organization_id)
        REFERENCES carrier_connections(id, organization_id)
        ON DELETE RESTRICT,
    CONSTRAINT uq_trunks_org_name UNIQUE (organization_id, name),
    CONSTRAINT uq_trunks_id_org UNIQUE (id, organization_id),
    CONSTRAINT chk_trunks_name CHECK (
        length(btrim(name)) > 0
    ),
    CONSTRAINT chk_trunks_direction CHECK (
        direction IN ('inbound', 'outbound', 'bidirectional')
    ),
    CONSTRAINT chk_trunks_status CHECK (
        status IN ('active', 'disabled')
    )
);

CREATE INDEX IF NOT EXISTS idx_trunks_organization_status
    ON trunks (organization_id, status);

CREATE INDEX IF NOT EXISTS idx_trunks_connection_id
    ON trunks (carrier_connection_id);

CREATE TABLE IF NOT EXISTS trunk_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    trunk_id UUID NOT NULL,

    host TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 5060,
    transport TEXT NOT NULL DEFAULT 'udp',
    direction TEXT NOT NULL DEFAULT 'bidirectional',
    priority INTEGER NOT NULL DEFAULT 10,
    weight INTEGER NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_trunk_endpoints_trunk_org
        FOREIGN KEY (trunk_id, organization_id)
        REFERENCES trunks(id, organization_id)
        ON DELETE CASCADE,
    CONSTRAINT uq_trunk_endpoints_target
        UNIQUE (trunk_id, host, port, transport, direction),
    CONSTRAINT uq_trunk_endpoints_id_org UNIQUE (id, organization_id),
    CONSTRAINT chk_trunk_endpoints_host CHECK (
        length(btrim(host)) > 0
        AND host !~ '\\s'
    ),
    CONSTRAINT chk_trunk_endpoints_port CHECK (
        port BETWEEN 1 AND 65535
    ),
    CONSTRAINT chk_trunk_endpoints_transport CHECK (
        transport IN ('udp', 'tcp', 'tls')
    ),
    CONSTRAINT chk_trunk_endpoints_direction CHECK (
        direction IN ('inbound', 'outbound', 'bidirectional')
    ),
    CONSTRAINT chk_trunk_endpoints_priority CHECK (
        priority >= 0
    ),
    CONSTRAINT chk_trunk_endpoints_weight CHECK (
        weight > 0
    )
);

CREATE INDEX IF NOT EXISTS idx_trunk_endpoints_trunk_enabled
    ON trunk_endpoints (trunk_id, enabled, priority);

CREATE INDEX IF NOT EXISTS idx_trunk_endpoints_organization_id
    ON trunk_endpoints (organization_id);

CREATE TRIGGER set_trunks_updated_at
BEFORE UPDATE ON trunks
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_trunk_endpoints_updated_at
BEFORE UPDATE ON trunk_endpoints
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
