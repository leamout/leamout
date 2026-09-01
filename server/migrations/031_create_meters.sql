CREATE TABLE IF NOT EXISTS meters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    unit TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_meters_key CHECK (
        length(trim(key)) > 0 AND key !~ '[[:space:]]'
    ),
    CONSTRAINT chk_meters_name CHECK (length(trim(name)) > 0),
    CONSTRAINT chk_meters_unit CHECK (length(trim(unit)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_meters_active
    ON meters (created_at DESC)
    WHERE active;

CREATE TRIGGER set_meters_updated_at
BEFORE UPDATE ON meters
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
