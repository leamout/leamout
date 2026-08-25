CREATE TABLE IF NOT EXISTS sip_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    domain TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_sip_domains_status CHECK (status IN ('active', 'disabled'))
);

CREATE INDEX IF NOT EXISTS idx_sip_domains_tenant_id
    ON sip_domains (tenant_id);

CREATE INDEX IF NOT EXISTS idx_sip_domains_status
    ON sip_domains (status);
