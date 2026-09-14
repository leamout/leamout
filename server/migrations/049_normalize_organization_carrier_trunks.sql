-- Organization-scoped carrier connections are customer-selected connectivity.
-- Provider brand does not change that ownership model: a Leamout Carrier
-- connection selected by a customer is BYOC just like any other carrier.
CREATE OR REPLACE FUNCTION validate_trunk_connectivity()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    connection_scope TEXT;
    connection_org UUID;
BEGIN
    IF NEW.carrier_connection_id IS NULL THEN
        -- Retain the Leamout-operated hosted-edge credential record used for
        -- carrier access. It is not a physical self-hosted runtime trunk.
        IF NEW.organization_id IS NULL OR NEW.provisioning_mode <> 'managed' THEN
            RAISE EXCEPTION 'only tenant carrier-edge access may omit carrier_connection_id';
        END IF;
        RETURN NEW;
    END IF;

    SELECT cc.scope, cc.organization_id
    INTO connection_scope, connection_org
    FROM carrier_connections AS cc
    WHERE cc.id = NEW.carrier_connection_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'carrier connection does not exist';
    END IF;

    IF NEW.organization_id IS NULL THEN
        IF NEW.provisioning_mode <> 'managed'
           OR connection_scope <> 'platform'
           OR connection_org IS NOT NULL THEN
            RAISE EXCEPTION 'platform trunks require a platform managed carrier connection';
        END IF;
        RETURN NEW;
    END IF;

    IF connection_scope <> 'organization'
       OR connection_org IS DISTINCT FROM NEW.organization_id THEN
        RAISE EXCEPTION 'tenant trunk carrier connection must belong to the same organization';
    END IF;

    IF NEW.provisioning_mode <> 'byoc' THEN
        RAISE EXCEPTION 'organization carrier connections require BYOC trunks';
    END IF;

    RETURN NEW;
END;
$$;

-- Reclassify any old organization-scoped Leamout Carrier trunks. Hosted-edge
-- credential rows have no carrier_connection_id and are intentionally excluded.
UPDATE trunks AS t
SET provisioning_mode = 'byoc'
FROM carrier_connections AS cc
WHERE cc.id = t.carrier_connection_id
  AND cc.scope = 'organization'
  AND cc.organization_id = t.organization_id
  AND t.organization_id IS NOT NULL
  AND t.provisioning_mode = 'managed';
