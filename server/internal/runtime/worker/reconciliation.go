package worker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/database/pgconv"
)

// sessionReconciliationJob repairs Voice Agent sessions left active after a
// terminal call when a worker crashed or missed a FreeSWITCH event.
type sessionReconciliationJob struct {
	queries *sqlc.Queries
}

func (j *sessionReconciliationJob) Reconcile(ctx context.Context) (int64, error) {
	if j == nil || j.queries == nil {
		return 0, fmt.Errorf("voice agent session reconciliation store is unavailable")
	}
	updated, err := j.queries.ReconcileTerminatedVoiceAgentSessions(ctx)
	if err != nil {
		return 0, fmt.Errorf("reconcile terminated Voice Agent sessions: %w", err)
	}
	return updated, nil
}

func (j *sessionReconciliationJob) Run(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		if count, err := j.Reconcile(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Printf("Voice Agent session reconciliation failed: %v", err)
		} else if count > 0 {
			log.Printf("reconciled %d terminated Voice Agent sessions", count)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

// toolExecutionReconciliationJob finalizes abandoned execution claims without
// dispatching the external tool again. An old "processing" row may represent
// a side effect that succeeded before the worker lost its result.
const toolExecutionUncertaintyWindow = 90 * time.Second

type toolExecutionReconciliationJob struct {
	queries *sqlc.Queries
}

func (j *toolExecutionReconciliationJob) Reconcile(ctx context.Context) (int64, error) {
	if j == nil || j.queries == nil {
		return 0, fmt.Errorf("tool execution reconciliation store is unavailable")
	}

	cutoff := pgconv.TimeToTimestamptz(
		time.Now().UTC().Add(-toolExecutionUncertaintyWindow),
	)
	updated, err := j.queries.ReconcileUncertainVoiceAgentToolExecutions(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("reconcile uncertain Voice Agent tool executions: %w", err)
	}
	return updated, nil
}

func (j *toolExecutionReconciliationJob) Run(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		if count, err := j.Reconcile(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			log.Printf("Voice Agent tool execution reconciliation failed: %v", err)
		} else if count > 0 {
			log.Printf("marked %d interrupted Voice Agent tool executions as uncertain", count)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}
