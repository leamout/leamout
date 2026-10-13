package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type reconciliationRepository interface {
	ListReconciliationTargets(context.Context, uuid.UUID) ([]ReconciliationTarget, error)
}

type subscriptionReconciler interface {
	ReconcileSubscription(context.Context, ReconciliationTarget) error
}

type ReconciliationJob struct {
	repo     reconciliationRepository
	service  subscriptionReconciler
	interval time.Duration
	report   func(context.Context, error)
}

func NewReconciliationJob(
	repo reconciliationRepository,
	service subscriptionReconciler,
	interval time.Duration,
	report func(context.Context, error),
) (*ReconciliationJob, error) {
	if repo == nil || service == nil || interval <= 0 || report == nil {
		return nil, fmt.Errorf("billing reconciliation dependencies and positive interval required")
	}
	return &ReconciliationJob{
		repo:     repo,
		service:  service,
		interval: interval,
		report:   report,
	}, nil
}

func (j *ReconciliationJob) Run(ctx context.Context) error {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		if err := j.reconcile(ctx); err != nil && ctx.Err() == nil {
			j.report(ctx, err)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (j *ReconciliationJob) reconcile(ctx context.Context) error {
	cursor := uuid.Nil
	for {
		targets, err := j.repo.ListReconciliationTargets(ctx, cursor)
		if err != nil {
			return err
		}
		for _, target := range targets {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if err := j.service.ReconcileSubscription(ctx, target); err != nil {
				j.report(ctx, fmt.Errorf("reconcile organization %s: %w", target.OrganizationID, err))
			}
		}
		if len(targets) < 100 {
			return nil
		}
		next := targets[len(targets)-1].OrganizationID
		if next == cursor {
			return fmt.Errorf("billing reconciliation cursor did not advance")
		}
		cursor = next
	}
}
