-- Carrier connectivity has two ownership modes:
--
--   organization: customer-owned/BYOC connectivity
--   platform:     Leamout-managed upstream connectivity
--
-- Phase 1 changes only the carrier-connection owner. Trunks, endpoints, source
-- IPs, and runtime credential rows remain organization-scoped until the later
-- managed-routing phase explicitly teaches those resources how to inherit a
-- platform-owned connection.

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

-- Ownership is immutable. Moving an existing connection between a tenant and
-- the platform would make historical attribution ambiguous; create a new
-- connection instead.
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
