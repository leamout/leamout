INSERT INTO organizations (id, name, status) VALUES
('00000000-0000-0000-0000-000000005001', 'Self Hosted Managed Acceptance', 'active');

INSERT INTO carrier_connections (
    id, organization_id, provider_id, scope, name, status, inbound_enabled, inbound_auth_method
) VALUES (
    '00000000-0000-0000-0000-000000005020',
    '00000000-0000-0000-0000-000000005001',
    (SELECT id FROM carrier_providers WHERE slug = 'leamout'),
    'organization', 'Leamout Managed Carrier', 'active', true, 'ip'
);
INSERT INTO carrier_connection_source_ips (organization_id, carrier_connection_id, cidr) VALUES
('00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000005020', '172.30.0.1/32'),
('00000000-0000-0000-0000-000000005001', '00000000-0000-0000-0000-000000005020', '172.32.0.1/32');

INSERT INTO phone_numbers (
    id, organization_id, number, country_code, provisioning_mode,
    carrier_connection_id, voice_enabled, status
) VALUES (
    '00000000-0000-0000-0000-000000005030',
    '00000000-0000-0000-0000-000000005001', '+15551235001', 'US', 'byoc',
    '00000000-0000-0000-0000-000000005020', true, 'active'
);
INSERT INTO voice_applications (id, organization_id, name, status) VALUES
('00000000-0000-0000-0000-000000005040', '00000000-0000-0000-0000-000000005001', 'Inbound Acceptance', 'active');
INSERT INTO voice_bindings (id, voice_application_id, phone_number_id) VALUES
('00000000-0000-0000-0000-000000005041', '00000000-0000-0000-0000-000000005040', '00000000-0000-0000-0000-000000005030');
