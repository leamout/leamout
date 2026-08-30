CREATE TABLE IF NOT EXISTS carrier_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    adapter TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_carrier_providers_slug UNIQUE (slug),
    CONSTRAINT chk_carrier_providers_slug CHECK (
        slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'
    ),
    CONSTRAINT chk_carrier_providers_name CHECK (
        length(btrim(name)) > 0
    ),
    CONSTRAINT chk_carrier_providers_adapter CHECK (
        length(btrim(adapter)) > 0
    ),
    CONSTRAINT chk_carrier_providers_status CHECK (
        status IN ('active', 'disabled')
    )
);

CREATE INDEX IF NOT EXISTS idx_carrier_providers_status
    ON carrier_providers (status);

CREATE TRIGGER set_carrier_providers_updated_at
BEFORE UPDATE ON carrier_providers
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

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
