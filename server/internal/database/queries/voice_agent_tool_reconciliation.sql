-- name: ReconcileUncertainVoiceAgentToolExecutions :execrows
-- The worker may die after a webhook or call-control operation took effect,
-- but before its result could be stored. Never automatically replay that
-- operation: make the uncertainty durable after the maximum execution window.
-- The processing predicate also protects a result committed concurrently.
UPDATE voice_agent_tool_executions
SET
    state = 'failed',
    error_message = 'Tool outcome uncertain after an interrupted execution; the action may have completed. Do not retry automatically.',
    completed_at = NOW(),
    updated_at = NOW()
WHERE state = 'processing'
  AND started_at <= sqlc.arg(cutoff)::timestamptz;
