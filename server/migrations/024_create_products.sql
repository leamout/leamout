CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_products_code CHECK (
        length(trim(code)) > 0 AND code !~ '[[:space:]]'
    ),
    CONSTRAINT chk_products_name CHECK (length(trim(name)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_products_active
    ON products (created_at DESC)
    WHERE active;

CREATE TRIGGER set_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
