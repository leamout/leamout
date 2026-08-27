CREATE TABLE IF NOT EXISTS phone_numbers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    number TEXT NOT NULL,
    country_code CHAR(2) NOT NULL,
    provider_id UUID,
    voice_enabled BOOLEAN NOT NULL DEFAULT true,
    sms_enabled BOOLEAN NOT NULL DEFAULT false,
    status TEXT NOT NULL DEFAULT 'active',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_phone_numbers_number UNIQUE (number),
    CONSTRAINT chk_phone_numbers_number CHECK (
        number ~ '^\+[1-9][0-9]{6,14}$'
    ),
    CONSTRAINT chk_phone_numbers_country_code CHECK (
        country_code ~ '^[A-Z]{2}$'
    ),
    CONSTRAINT chk_phone_numbers_status CHECK (
        status IN ('active', 'disabled', 'porting', 'released')
    )
);

CREATE INDEX IF NOT EXISTS idx_phone_numbers_organization_status
    ON phone_numbers (organization_id, status);

CREATE INDEX IF NOT EXISTS idx_phone_numbers_organization_country
    ON phone_numbers (organization_id, country_code);

CREATE INDEX IF NOT EXISTS idx_phone_numbers_number_prefix
    ON phone_numbers (number text_pattern_ops);

CREATE TRIGGER set_phone_numbers_updated_at
BEFORE UPDATE ON phone_numbers
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
