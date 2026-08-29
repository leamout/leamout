CREATE TABLE IF NOT EXISTS deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    license_id UUID NOT NULL REFERENCES licenses(id),
    deployment_id TEXT NOT NULL,
    name TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    activated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ,
    deactivated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_deployments_license_deployment UNIQUE (license_id, deployment_id),
    CONSTRAINT chk_deployments_deployment_id CHECK (length(trim(deployment_id)) > 0),
    CONSTRAINT chk_deployments_name CHECK (
        name IS NULL OR length(trim(name)) > 0
    ),
    CONSTRAINT chk_deployments_status CHECK (status IN ('active', 'deactivated')),
    CONSTRAINT chk_deployments_deactivated CHECK (
        (status = 'active' AND deactivated_at IS NULL)
        OR (status = 'deactivated' AND deactivated_at IS NOT NULL)
    ),
    CONSTRAINT chk_deployments_last_seen_at CHECK (
        last_seen_at IS NULL OR last_seen_at >= activated_at
    ),
    CONSTRAINT chk_deployments_deactivated_at CHECK (
        deactivated_at IS NULL OR deactivated_at >= activated_at
    )
);

CREATE INDEX IF NOT EXISTS idx_deployments_license_status
    ON deployments (license_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_deployments_last_seen_at
    ON deployments (last_seen_at)
    WHERE status = 'active';

CREATE TRIGGER set_deployments_updated_at
BEFORE UPDATE ON deployments
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
