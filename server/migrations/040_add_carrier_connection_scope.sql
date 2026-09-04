-- Carrier connectivity has two ownership modes:
--
--   organization: customer-owned/BYOC connectivity
--   platform:     Leamout-managed upstream connectivity
--
-- The carrier connection is the authoritative owner. Existing organization_id
-- columns on child SIP resources remain as derived access keys for the public
-- BYOC API; platform-owned children carry NULL there.

ALTER TABLE carrier_connections
    ADD COLUMN scope TEXT NOT NULL DEFAULT 'organization';

ALTER TABLE carrier_connections
    ALTER COLUMN organization_id DROP NOT NULL;

ALTER TABLE carrier_connections
    ADD CONSTRAINT chk_carrier_connections_scope
        CHECK (scope IN ('organization', 'platform')),
    ADD CONSTRAINT chk_carrier_connections_owner
        CHECK (
            (scope = 'organization' AND organization_id IS NOT NULL)
            OR (scope = 'platform' AND organization_id IS NULL)
        );

CREATE UNIQUE INDEX uq_carrier_connections_platform_name
    ON carrier_connections (name)
    WHERE scope = 'platform';

CREATE INDEX idx_carrier_connections_scope_status
    ON carrier_connections (scope, status);

-- Ownership is immutable. Moving an existing SIP connection between a tenant
-- and the platform would make call attribution, credentials, trunks and source
-- networks ambiguous; create a new connection instead.
CREATE FUNCTION prevent_carrier_connection_owner_change()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF OLD.scope IS DISTINCT FROM NEW.scope
       OR OLD.organization_id IS DISTINCT FROM NEW.organization_id THEN
        RAISE EXCEPTION 'carrier connection ownership is immutable'
            USING ERRCODE = '23514';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER prevent_carrier_connection_owner_change
BEFORE UPDATE OF scope, organization_id ON carrier_connections
FOR EACH ROW
EXECUTE FUNCTION prevent_carrier_connection_owner_change();

-- Child resource organization IDs are derived from their carrier connection.
-- They stay populated for organization-scoped BYOC resources so existing public
-- tenant lookups remain cheap, and become NULL for platform-scoped resources.
ALTER TABLE carrier_connection_source_ips
    ALTER COLUMN organization_id DROP NOT NULL;

ALTER TABLE carrier_connection_source_ips
    ADD CONSTRAINT fk_carrier_connection_source_ips_connection
        FOREIGN KEY (carrier_connection_id)
        REFERENCES carrier_connections(id)
        ON DELETE CASCADE;

ALTER TABLE trunks
    ALTER COLUMN organization_id DROP NOT NULL;

ALTER TABLE trunks
    ADD CONSTRAINT fk_trunks_connection
        FOREIGN KEY (carrier_connection_id)
        REFERENCES carrier_connections(id)
        ON DELETE RESTRICT;

CREATE UNIQUE INDEX uq_trunks_platform_connection_name
    ON trunks (carrier_connection_id, name)
    WHERE organization_id IS NULL;

ALTER TABLE trunk_endpoints
    ALTER COLUMN organization_id DROP NOT NULL;

ALTER TABLE trunk_endpoints
    ADD CONSTRAINT fk_trunk_endpoints_trunk
        FOREIGN KEY (trunk_id)
        REFERENCES trunks(id)
        ON DELETE CASCADE;

ALTER TABLE carrier_digest_credentials
    ALTER COLUMN organization_id DROP NOT NULL;

CREATE FUNCTION derive_carrier_connection_organization_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT organization_id
    INTO NEW.organization_id
    FROM carrier_connections
    WHERE id = NEW.carrier_connection_id;

    RETURN NEW;
END;
$$;

CREATE TRIGGER derive_source_ip_organization_id
BEFORE INSERT OR UPDATE OF carrier_connection_id ON carrier_connection_source_ips
FOR EACH ROW
EXECUTE FUNCTION derive_carrier_connection_organization_id();

CREATE TRIGGER derive_trunk_organization_id
BEFORE INSERT OR UPDATE OF carrier_connection_id ON trunks
FOR EACH ROW
EXECUTE FUNCTION derive_carrier_connection_organization_id();

CREATE TRIGGER derive_carrier_digest_organization_id
BEFORE INSERT OR UPDATE OF carrier_connection_id ON carrier_digest_credentials
FOR EACH ROW
EXECUTE FUNCTION derive_carrier_connection_organization_id();

CREATE FUNCTION derive_trunk_endpoint_organization_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    SELECT cc.organization_id
    INTO NEW.organization_id
    FROM trunks AS t
    JOIN carrier_connections AS cc ON cc.id = t.carrier_connection_id
    WHERE t.id = NEW.trunk_id;

    RETURN NEW;
END;
$$;

CREATE TRIGGER derive_trunk_endpoint_organization_id
BEFORE INSERT OR UPDATE OF trunk_id ON trunk_endpoints
FOR EACH ROW
EXECUTE FUNCTION derive_trunk_endpoint_organization_id();

-- Rebuild the OpenSIPS credential views against carrier_connection ownership.
-- Platform inbound authentication is intentionally not activated here; tenant
-- resolution for shared managed ingress is part of the later routing phase.
CREATE OR REPLACE VIEW opensips_carrier_digest_credentials AS
SELECT
    d.carrier_connection_id,
    d.organization_id,
    d.direction,
    d.username,
    d.realm,
    CASE WHEN d.direction = 'outbound' THEN '0x' || d.ha1_md5 ELSE d.ha1_md5 END AS password
FROM carrier_digest_credentials AS d
JOIN carrier_connections AS c
  ON c.id = d.carrier_connection_id
WHERE c.status = 'active'
  AND (
      d.direction = 'outbound'
      OR (
          c.scope = 'organization'
          AND c.inbound_enabled = true
          AND c.inbound_auth_method = 'digest'
      )
  );

CREATE OR REPLACE VIEW opensips_inbound_carrier_credentials AS
SELECT
    d.carrier_connection_id,
    d.organization_id,
    d.username,
    d.realm AS domain,
    d.ha1_md5
FROM carrier_digest_credentials AS d
JOIN carrier_connections AS c
  ON c.id = d.carrier_connection_id
JOIN organizations AS o
  ON o.id = c.organization_id
WHERE d.direction = 'inbound'
  AND c.scope = 'organization'
  AND c.status = 'active'
  AND c.inbound_enabled = true
  AND c.inbound_auth_method = 'digest'
  AND o.status = 'active'
  AND o.deleted_at IS NULL;

CREATE OR REPLACE VIEW opensips_outbound_carrier_credentials AS
SELECT
    d.carrier_connection_id,
    d.organization_id,
    d.username,
    d.realm,
    '0x' || d.ha1_md5 AS password
FROM carrier_digest_credentials AS d
JOIN carrier_connections AS c
  ON c.id = d.carrier_connection_id
LEFT JOIN organizations AS o
  ON o.id = c.organization_id
WHERE d.direction = 'outbound'
  AND c.status = 'active'
  AND (
      c.scope = 'platform'
      OR (
          c.scope = 'organization'
          AND o.status = 'active'
          AND o.deleted_at IS NULL
      )
  );
