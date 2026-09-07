INSERT INTO organizations (id, name, status) VALUES
('00000000-0000-0000-0000-000000006001', 'Cloud Managed Acceptance', 'active');

INSERT INTO organization_tokens (id, organization_id, name, token_hash, token_prefix, scopes) VALUES (
    '00000000-0000-0000-0000-000000006002',
    '00000000-0000-0000-0000-000000006001',
    'cloud-managed-acceptance',
    'Y6rtC8BR465xPLxeDGcWiQyGBL6zR5L9JcqWYj8naWE',
    'lm_org_v1smoke0',
    '["numbers:read","numbers:write","voice-applications:read","voice-applications:write","calls:read","calls:write"]'::jsonb
);

INSERT INTO carrier_connections (
    id, provider_id, scope, name, status, inbound_enabled, inbound_auth_method
) VALUES (
    '00000000-0000-0000-0000-000000006010',
    '26c5448a-2540-4731-848d-9c713c19d8cd',
    'platform', 'Cloud managed DID ingress', 'active', true, 'ip'
);
INSERT INTO carrier_connection_provider_resources (
    carrier_connection_id, provider_id, resource_type, provider_resource_id
) VALUES (
    '00000000-0000-0000-0000-000000006010',
    '26c5448a-2540-4731-848d-9c713c19d8cd',
    'voice_in_trunk', 'voice-in-cloud-1'
);
