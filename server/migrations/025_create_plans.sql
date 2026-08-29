CREATE TABLE IF NOT EXISTS plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id),
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_plans_code CHECK (
        length(trim(code)) > 0 AND code !~ '[[:space:]]'
    ),
    CONSTRAINT chk_plans_name CHECK (length(trim(name)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_plans_product_active
    ON plans (product_id, created_at DESC)
    WHERE active;

CREATE TRIGGER set_plans_updated_at
BEFORE UPDATE ON plans
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
