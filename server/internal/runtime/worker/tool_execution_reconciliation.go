package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// A processing record is deliberately never re-dispatched. A worker may have
// performed the remote action and crashed before committing the response.
const toolExecutionUncertaintyWindow = 90 * time.Second

type toolExecutionReconciliationJob struct {
	db *pgxpool.Pool
}

func (j *toolExecutionReconciliationJob) Reconcile(ctx context.Context) (int64, error) {
	if j == nil || j.db == nil {
		return 0, fmt.Errorf("tool execution reconciliation database is unavailable")
	}
	// A conditional, row-locked update is safe with concurrent workers. An
	// execution which completed after this scan started cannot be overwritten.
	tag, err := j.db.Exec(ctx, `
UPDATE voice_agent_tool_executions
SET state = 'failed',
    error_message = 'Tool outcome is uncertain after worker interruption; do not retry this tool call automatically',
    completed_at = NOW(),
    updated_at = NOW()
WHERE state = 'processing'
  AND started_at < NOW() - INTERVAL '90 seconds'
`)
	if err != nil {
		return 0, fmt.Errorf("reconcile uncertain voice agent tool executions: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (j *toolExecutionReconciliationJob) Run(ctx context.Context) error {
	if count, err := j.Reconcile(ctx); err != nil {
		return err
	} else if count > 0 {
		log.Printf("marked %d interrupted tool executions as uncertain", count)
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			count, err := j.Reconcile(ctx)
			if err != nil {
				return err
			}
			if count > 0 {
				log.Printf("marked %d interrupted tool executions as uncertain", count)
			}
		}
	}
}
