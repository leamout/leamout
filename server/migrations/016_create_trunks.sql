CREATE TABLE IF NOT EXISTS trunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID,
    carrier_connection_id UUID REFERENCES carrier_connections(id) ON DELETE RESTRICT,

    provisioning_mode TEXT NOT NULL DEFAULT 'byoc',
    name TEXT NOT NULL,
    direction TEXT NOT NULL DEFAULT 'bidirectional',
    status TEXT NOT NULL DEFAULT 'active',
    managed_default BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_trunks_org_name UNIQUE (organization_id, name),
    CONSTRAINT uq_trunks_id_org UNIQUE (id, organization_id),
    CONSTRAINT chk_trunks_provisioning_mode CHECK (
        provisioning_mode IN ('byoc', 'managed')
    ),
    CONSTRAINT chk_trunks_ownership_shape CHECK (
        (
            organization_id IS NOT NULL
            AND provisioning_mode = 'byoc'
            AND carrier_connection_id IS NOT NULL
        )
        OR (
            organization_id IS NOT NULL
            AND provisioning_mode = 'managed'
        )
        OR (
            organization_id IS NULL
            AND provisioning_mode = 'managed'
            AND carrier_connection_id IS NOT NULL
        )
    ),
    CONSTRAINT chk_trunks_managed_default CHECK (
        managed_default = false
        OR (
            organization_id IS NULL
            AND provisioning_mode = 'managed'
            AND carrier_connection_id IS NOT NULL
        )
    ),
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

CREATE UNIQUE INDEX IF NOT EXISTS uq_trunks_platform_name
    ON trunks (name)
    WHERE organization_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_trunks_managed_default
    ON trunks (managed_default)
    WHERE organization_id IS NULL AND managed_default = true;

CREATE INDEX IF NOT EXISTS idx_trunks_organization_status
    ON trunks (organization_id, status);

CREATE INDEX IF NOT EXISTS idx_trunks_connection_id
    ON trunks (carrier_connection_id)
    WHERE carrier_connection_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_trunks_organization_mode
    ON trunks (organization_id, provisioning_mode, status)
    WHERE organization_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS trunk_endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID,
    trunk_id UUID NOT NULL REFERENCES trunks(id) ON DELETE CASCADE,

    host TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 5060,
    transport TEXT NOT NULL DEFAULT 'udp',
    direction TEXT NOT NULL DEFAULT 'bidirectional',
    priority INTEGER NOT NULL DEFAULT 10,
    weight INTEGER NOT NULL DEFAULT 100,
    enabled BOOLEAN NOT NULL DEFAULT true,
    health_status TEXT NOT NULL DEFAULT 'unknown',
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    last_checked_at TIMESTAMPTZ,
    last_response_code INTEGER,
    last_latency_ms INTEGER,
    last_error TEXT,
    cooldown_until TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

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
    ),
    CONSTRAINT chk_trunk_endpoint_health_status CHECK (
        health_status IN ('unknown', 'healthy', 'unhealthy')
    ),
    CONSTRAINT chk_trunk_endpoint_consecutive_failures CHECK (
        consecutive_failures >= 0
    ),
    CONSTRAINT chk_trunk_endpoint_response_code CHECK (
        last_response_code IS NULL OR last_response_code BETWEEN 100 AND 699
    ),
    CONSTRAINT chk_trunk_endpoint_latency CHECK (
        last_latency_ms IS NULL OR last_latency_ms >= 0
    )
);

CREATE INDEX IF NOT EXISTS idx_trunk_endpoints_trunk_enabled
    ON trunk_endpoints (trunk_id, enabled, priority);

CREATE INDEX IF NOT EXISTS idx_trunk_endpoints_organization_id
    ON trunk_endpoints (organization_id);

CREATE INDEX IF NOT EXISTS idx_trunk_endpoints_health_probe
    ON trunk_endpoints (health_status, cooldown_until, last_checked_at)
    WHERE enabled = true;

-- BYOC, installed Leamout Carrier, and internal platform trunks inherit
-- organization ownership from their carrier connection. A Cloud-authoritative
-- tenant managed trunk intentionally has no carrier connection and preserves
-- the organization supplied by the control plane.
CREATE FUNCTION derive_trunk_organization_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.carrier_connection_id IS NOT NULL THEN
        SELECT organization_id
        INTO NEW.organization_id
        FROM carrier_connections
        WHERE id = NEW.carrier_connection_id;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER derive_trunk_organization_id
BEFORE INSERT OR UPDATE OF carrier_connection_id ON trunks
FOR EACH ROW
EXECUTE FUNCTION derive_trunk_organization_id();

-- Enforce the runtime/connectivity matrix at the persistence boundary:
--   tenant BYOC trunk     -> any active organization carrier except Leamout
--   tenant managed/Cloud  -> no tenant carrier connection
--   tenant managed/local  -> organization Leamout Carrier connection only
--   platform managed      -> internal platform carrier connection
CREATE FUNCTION validate_trunk_connectivity()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    connection_scope TEXT;
    connection_org UUID;
    provider_slug TEXT;
BEGIN
    IF NEW.carrier_connection_id IS NULL THEN
        IF NEW.organization_id IS NULL OR NEW.provisioning_mode <> 'managed' THEN
            RAISE EXCEPTION 'only tenant managed trunks may omit carrier_connection_id';
        END IF;
        RETURN NEW;
    END IF;

    SELECT cc.scope, cc.organization_id, cp.slug
    INTO connection_scope, connection_org, provider_slug
    FROM carrier_connections AS cc
    JOIN carrier_providers AS cp ON cp.id = cc.provider_id
    WHERE cc.id = NEW.carrier_connection_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'carrier connection does not exist';
    END IF;

    IF NEW.organization_id IS NULL THEN
        IF NEW.provisioning_mode <> 'managed' OR connection_scope <> 'platform' OR connection_org IS NOT NULL THEN
            RAISE EXCEPTION 'platform trunks require a platform managed carrier connection';
        END IF;
        RETURN NEW;
    END IF;

    IF connection_scope <> 'organization' OR connection_org IS DISTINCT FROM NEW.organization_id THEN
        RAISE EXCEPTION 'tenant trunk carrier connection must belong to the same organization';
    END IF;

    IF NEW.provisioning_mode = 'managed' AND provider_slug <> 'leamout' THEN
        RAISE EXCEPTION 'tenant managed trunks may only use the Leamout Carrier connection';
    END IF;

    IF NEW.provisioning_mode = 'byoc' AND provider_slug = 'leamout' THEN
        RAISE EXCEPTION 'Leamout Carrier connections require managed trunks';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER validate_trunk_connectivity
BEFORE INSERT OR UPDATE OF organization_id, carrier_connection_id, provisioning_mode ON trunks
FOR EACH ROW
EXECUTE FUNCTION validate_trunk_connectivity();

CREATE FUNCTION derive_trunk_endpoint_organization_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT organization_id
    INTO NEW.organization_id
    FROM trunks
    WHERE id = NEW.trunk_id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER derive_trunk_endpoint_organization_id
BEFORE INSERT OR UPDATE OF trunk_id ON trunk_endpoints
FOR EACH ROW
EXECUTE FUNCTION derive_trunk_endpoint_organization_id();

CREATE TRIGGER set_trunks_updated_at
BEFORE UPDATE ON trunks
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER set_trunk_endpoints_updated_at
BEFORE UPDATE ON trunk_endpoints
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
