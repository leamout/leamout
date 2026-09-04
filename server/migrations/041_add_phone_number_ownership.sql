-- Phone numbers remain organization-owned customer resources regardless of how
-- they were acquired. provisioning_mode records who owns the upstream
-- provisioning relationship:
--
--   byoc:    the organization supplied the number/carrier relationship
--   managed: Leamout provisioned the number through an upstream provider

ALTER TABLE phone_numbers
    ADD COLUMN provisioning_mode TEXT NOT NULL DEFAULT 'byoc',
    ADD COLUMN provider_id UUID REFERENCES carrier_providers(id) ON DELETE RESTRICT;

-- provider_resource_id existed before provider attribution. There is no safe
-- way to infer a provider for historical non-NULL values, so fail rather than
-- silently assign the wrong upstream owner.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM phone_numbers
        WHERE provider_resource_id IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'existing phone_numbers.provider_resource_id values must be attributed before enabling managed number ownership';
    END IF;
END;
$$;

ALTER TABLE phone_numbers
    ADD CONSTRAINT chk_phone_numbers_provisioning_mode
        CHECK (provisioning_mode IN ('byoc', 'managed')),
    ADD CONSTRAINT chk_phone_numbers_provider_ownership
        CHECK (
            (
                provisioning_mode = 'byoc'
                AND provider_id IS NULL
                AND provider_resource_id IS NULL
            )
            OR (
                provisioning_mode = 'managed'
                AND provider_id IS NOT NULL
                AND provider_resource_id IS NOT NULL
                AND length(btrim(provider_resource_id)) > 0
            )
        );

CREATE UNIQUE INDEX uq_phone_numbers_provider_resource
    ON phone_numbers (provider_id, provider_resource_id)
    WHERE provider_id IS NOT NULL AND provider_resource_id IS NOT NULL;

CREATE INDEX idx_phone_numbers_organization_provisioning_status
    ON phone_numbers (organization_id, provisioning_mode, status);

-- BYOC numbers may bind only to their organization's carrier connection.
-- Managed numbers may bind only to a Leamout platform connection. A NULL
-- carrier_connection_id is valid while provisioning is incomplete or when the
-- number does not use SIP voice.
CREATE FUNCTION validate_phone_number_carrier_scope()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    connection_scope TEXT;
    connection_organization_id UUID;
BEGIN
    IF NEW.carrier_connection_id IS NULL THEN
        RETURN NEW;
    END IF;

    SELECT scope, organization_id
    INTO connection_scope, connection_organization_id
    FROM carrier_connections
    WHERE id = NEW.carrier_connection_id;

    IF NEW.provisioning_mode = 'byoc' THEN
        IF connection_scope <> 'organization'
           OR connection_organization_id IS DISTINCT FROM NEW.organization_id THEN
            RAISE EXCEPTION 'BYOC phone number must use an organization-owned carrier connection'
                USING ERRCODE = '23514';
        END IF;
    ELSIF NEW.provisioning_mode = 'managed' THEN
        IF connection_scope <> 'platform' THEN
            RAISE EXCEPTION 'managed phone number must use a platform-owned carrier connection'
                USING ERRCODE = '23514';
        END IF;
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER validate_phone_number_carrier_scope
BEFORE INSERT OR UPDATE OF carrier_connection_id, provisioning_mode, organization_id
ON phone_numbers
FOR EACH ROW
EXECUTE FUNCTION validate_phone_number_carrier_scope();
