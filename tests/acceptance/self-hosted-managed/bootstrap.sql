INSERT INTO organizations (id, name, status) VALUES
('00000000-0000-0000-0000-000000005001', 'Self Hosted Managed Acceptance', 'active');

INSERT INTO licenses (id, organization_id, status, max_deployments) VALUES
('00000000-0000-0000-0000-000000005010', '00000000-0000-0000-0000-000000005001', 'active', 1);
INSERT INTO deployments (id, license_id, deployment_id, name, status) VALUES
('00000000-0000-0000-0000-000000005011', '00000000-0000-0000-0000-000000005010', 'acceptance-runtime', 'Acceptance Runtime', 'active');
INSERT INTO runtime_attachments (
    id, deployment_id, ingress_host, ingress_port, transport,
    verification_status, health_status, verified_at, last_checked_at
) VALUES (
    '00000000-0000-0000-0000-000000005012',
    '00000000-0000-0000-0000-000000005011',
    '172.30.0.11', 5060, 'tcp', 'verified', 'healthy', now(), now()
);

INSERT INTO carrier_connections (
    id, provider_id, scope, name, status, inbound_enabled, inbound_auth_method
) VALUES (
    '00000000-0000-0000-0000-000000005020',
    '26c5448a-2540-4731-848d-9c713c19d8cd',
    'platform', 'Managed inbound acceptance', 'active', true, 'ip'
);
INSERT INTO carrier_connection_source_ips (carrier_connection_id, cidr) VALUES
('00000000-0000-0000-0000-000000005020', '172.30.0.0/24');

INSERT INTO phone_numbers (
    id, organization_id, number, country_code, provisioning_mode,
    carrier_connection_id, provider_id, provider_resource_id, voice_enabled, status
) VALUES (
    '00000000-0000-0000-0000-000000005030',
    '00000000-0000-0000-0000-000000005001', '+15551235001', 'US', 'managed',
    '00000000-0000-0000-0000-000000005020',
    '26c5448a-2540-4731-848d-9c713c19d8cd', 'acceptance-managed-did', true, 'active'
);
INSERT INTO voice_applications (id, organization_id, name, status) VALUES
('00000000-0000-0000-0000-000000005040', '00000000-0000-0000-0000-000000005001', 'Inbound Acceptance', 'active');
INSERT INTO voice_bindings (id, voice_application_id, phone_number_id) VALUES
('00000000-0000-0000-0000-000000005041', '00000000-0000-0000-0000-000000005040', '00000000-0000-0000-0000-000000005030');
