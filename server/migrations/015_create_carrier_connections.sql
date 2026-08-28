CREATE TABLE IF NOT EXISTS carrier_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    provider_id UUID NOT NULL REFERENCES carrier_providers(id) ON DELETE RESTRICT,

    name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    outbound_auth_method TEXT NOT NULL DEFAULT 'none',
    auth_username TEXT,
    auth_secret_ciphertext TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_carrier_connections_org_name UNIQUE (organization_id, name),
    CONSTRAINT uq_carrier_connections_id_org UNIQUE (id, organization_id),
    CONSTRAINT chk_carrier_connections_name CHECK (
        length(btrim(name)) > 0
    ),
    CONSTRAINT chk_carrier_connections_status CHECK (
        status IN ('active', 'disabled')
    ),
    CONSTRAINT chk_carrier_connections_auth_method CHECK (
        outbound_auth_method IN ('none', 'digest')
    ),
    CONSTRAINT chk_carrier_connections_auth_fields CHECK (
        (
            outbound_auth_method = 'none'
            AND auth_username IS NULL
            AND auth_secret_ciphertext IS NULL
        ) OR (
            outbound_auth_method = 'digest'
            AND auth_username IS NOT NULL
            AND length(btrim(auth_username)) > 0
            AND auth_secret_ciphertext IS NOT NULL
            AND length(auth_secret_ciphertext) > 0
        )
    )
);

CREATE INDEX IF NOT EXISTS idx_carrier_connections_organization_status
    ON carrier_connections (organization_id, status);

CREATE INDEX IF NOT EXISTS idx_carrier_connections_provider_id
    ON carrier_connections (provider_id);

CREATE TABLE IF NOT EXISTS carrier_connection_source_ips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    carrier_connection_id UUID NOT NULL,
    cidr CIDR NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_carrier_connection_source_ips_connection_org
        FOREIGN KEY (carrier_connection_id, organization_id)
        REFERENCES carrier_connections(id, organization_id)
        ON DELETE CASCADE,
    CONSTRAINT uq_carrier_connection_source_ips_connection_cidr
        UNIQUE (carrier_connection_id, cidr)
);

CREATE INDEX IF NOT EXISTS idx_carrier_connection_source_ips_organization_id
    ON carrier_connection_source_ips (organization_id);

CREATE INDEX IF NOT EXISTS idx_carrier_connection_source_ips_cidr
    ON carrier_connection_source_ips USING gist (cidr inet_ops);

CREATE TRIGGER set_carrier_connections_updated_at
BEFORE UPDATE ON carrier_connections
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
