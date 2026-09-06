INSERT INTO organizations (id, name, status) VALUES
('00000000-0000-0000-0000-000000004001', 'Managed SIP Edge Acceptance', 'active');

INSERT INTO products (id, code, name) VALUES
('00000000-0000-0000-0000-000000004010', 'managed-edge-test', 'Managed Edge Test');
INSERT INTO plans (id, product_id, code, name) VALUES
('00000000-0000-0000-0000-000000004011', '00000000-0000-0000-0000-000000004010', 'managed-edge-test', 'Managed Edge Test');
INSERT INTO subscriptions (id, organization_id, plan_id, status) VALUES
('00000000-0000-0000-0000-000000004012', '00000000-0000-0000-0000-000000004001', '00000000-0000-0000-0000-000000004011', 'active');
INSERT INTO entitlements (plan_id, entitlement_key, kind, enabled) VALUES
('00000000-0000-0000-0000-000000004011', 'voice.managed.enabled', 'feature', true);
INSERT INTO entitlements (plan_id, entitlement_key, kind, limit_value) VALUES
('00000000-0000-0000-0000-000000004011', 'voice.managed.daily_spend_micros', 'limit', 1000000);

INSERT INTO carrier_connections (id, provider_id, scope, name, status) VALUES
('00000000-0000-0000-0000-000000004020', '300e6073-fe60-4d40-ac6d-808d74749a0c', 'platform', 'Managed edge wholesale', 'active');
INSERT INTO trunks (id, carrier_connection_id, provisioning_mode, name, direction, status, managed_default) VALUES
('00000000-0000-0000-0000-000000004021', '00000000-0000-0000-0000-000000004020', 'managed', 'Managed edge wholesale', 'outbound', 'active', true);
INSERT INTO trunk_endpoints (id, trunk_id, host, port, transport, direction, health_status) VALUES
('00000000-0000-0000-0000-000000004022', '00000000-0000-0000-0000-000000004021', 'managed-sip-edge-wholesale', 5060, 'udp', 'outbound', 'healthy');

INSERT INTO trunks (id, organization_id, provisioning_mode, name, direction, status) VALUES
('00000000-0000-0000-0000-000000004030', '00000000-0000-0000-0000-000000004001', 'managed', 'Customer managed trunk', 'outbound', 'active');
INSERT INTO trunk_credentials (trunk_id, organization_id, username, realm, ha1_md5) VALUES
('00000000-0000-0000-0000-000000004030', '00000000-0000-0000-0000-000000004001', 'edge-user', 'sip.leamout.com', md5('edge-user:sip.leamout.com:edge-password'));
INSERT INTO phone_numbers (id, organization_id, number, country_code, provisioning_mode, provider_id, provider_resource_id, voice_enabled, status) VALUES
('00000000-0000-0000-0000-000000004040', '00000000-0000-0000-0000-000000004001', '+15551234001', 'US', 'managed', '26c5448a-2540-4731-848d-9c713c19d8cd', 'managed-edge-caller', true, 'active');
