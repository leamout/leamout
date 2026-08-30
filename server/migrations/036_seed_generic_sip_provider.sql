-- Every deployment includes a carrier-neutral provider so organizations can
-- configure standards-based SIP trunks without operator-managed catalog data.
INSERT INTO carrier_providers (id, slug, name, adapter, status)
VALUES (
    '00000000-0000-0000-0000-000000002000',
    'generic-sip',
    'Generic SIP',
    'sip',
    'active'
)
ON CONFLICT (slug) DO NOTHING;
