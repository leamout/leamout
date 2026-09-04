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

-- Built-in provider definitions establish stable provider identities for
-- generic SIP connectivity and provider-specific managed integrations.
INSERT INTO carrier_providers (id, slug, name, adapter, status)
VALUES
    (
        'eb3a4a81-ca51-44a7-9410-0841428ae43c',
        'generic-sip',
        'Generic SIP',
        'sip',
        'active'
    ),
    (
        '26c5448a-2540-4731-848d-9c713c19d8cd',
        'didww',
        'DIDWW',
        'didww',
        'active'
    ),
    (
        '300e6073-fe60-4d40-ac6d-808d74749a0c',
        'commpeak',
        'CommPeak',
        'commpeak',
        'active'
    )
ON CONFLICT (slug) DO NOTHING;
