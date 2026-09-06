CREATE TABLE IF NOT EXISTS runtime_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id),
    ingress_host TEXT NOT NULL,
    ingress_port INTEGER NOT NULL DEFAULT 5061,
    transport TEXT NOT NULL DEFAULT 'tls',
    verification_status TEXT NOT NULL DEFAULT 'pending',
    health_status TEXT NOT NULL DEFAULT 'unknown',
    verified_at TIMESTAMPTZ,
    last_checked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_runtime_attachments_deployment UNIQUE (deployment_id),
    CONSTRAINT chk_runtime_attachments_host CHECK (length(trim(ingress_host)) > 0),
    CONSTRAINT chk_runtime_attachments_port CHECK (ingress_port BETWEEN 1 AND 65535),
    CONSTRAINT chk_runtime_attachments_transport CHECK (transport IN ('tcp', 'tls')),
    CONSTRAINT chk_runtime_attachments_verification CHECK (
        verification_status IN ('pending', 'verified', 'rejected')
    ),
    CONSTRAINT chk_runtime_attachments_health CHECK (
        health_status IN ('unknown', 'healthy', 'unhealthy')
    ),
    CONSTRAINT chk_runtime_attachments_verified_at CHECK (
        (verification_status = 'verified' AND verified_at IS NOT NULL)
        OR (verification_status <> 'verified' AND verified_at IS NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_runtime_attachments_deliverable
    ON runtime_attachments (deployment_id)
    WHERE verification_status = 'verified' AND health_status = 'healthy';

CREATE TRIGGER set_runtime_attachments_updated_at
BEFORE UPDATE ON runtime_attachments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
