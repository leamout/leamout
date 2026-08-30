-- auth_db requires its configured domain and password column names, while
-- uac_auth requires a 0x-prefixed HA1. Separate views keep those contracts
-- explicit and prevent credentials from one direction being selected in the
-- other direction.
CREATE VIEW opensips_inbound_carrier_credentials AS
SELECT
    d.carrier_connection_id,
    d.organization_id,
    d.username,
    d.realm AS domain,
    d.ha1_md5
FROM carrier_digest_credentials AS d
JOIN carrier_connections AS c
  ON c.id = d.carrier_connection_id
 AND c.organization_id = d.organization_id
JOIN organizations AS o
  ON o.id = d.organization_id
WHERE d.direction = 'inbound'
  AND c.status = 'active'
  AND c.inbound_enabled = true
  AND c.inbound_auth_method = 'digest'
  AND o.status = 'active'
  AND o.deleted_at IS NULL;

CREATE VIEW opensips_outbound_carrier_credentials AS
SELECT
    d.carrier_connection_id,
    d.organization_id,
    d.username,
    d.realm,
    '0x' || d.ha1_md5 AS password
FROM carrier_digest_credentials AS d
JOIN carrier_connections AS c
  ON c.id = d.carrier_connection_id
 AND c.organization_id = d.organization_id
JOIN organizations AS o
  ON o.id = d.organization_id
WHERE d.direction = 'outbound'
  AND c.status = 'active'
  AND o.status = 'active'
  AND o.deleted_at IS NULL;
