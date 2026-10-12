-- name: LockOrganizationForCapacity :one
SELECT id FROM organizations
WHERE id = sqlc.arg(organization_id)
  AND status = 'active'
  AND deleted_at IS NULL
FOR UPDATE;

-- name: CountOrganizationVoiceAgents :one
SELECT COUNT(*)::bigint AS count FROM voice_agents
WHERE organization_id = sqlc.arg(organization_id);

-- name: CountOrganizationSIPTrunks :one
SELECT COUNT(*)::bigint AS count FROM trunks
WHERE organization_id = sqlc.arg(organization_id);
