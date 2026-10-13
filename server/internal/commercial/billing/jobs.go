package billing

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/pkg/apperror"
)

const (
	webhookCleanupInterval = 24 * time.Hour
	webhookMaxAttempts     = int32(8)
	webhookPollInterval    = time.Second
)

type WebhookJob struct {
	repo    *Repository
	service *Service
	report  func(context.Context, error)
}

func NewWebhookJob(
	repo *Repository,
	service *Service,
	report func(context.Context, error),
) (*WebhookJob, error) {
	if repo == nil || repo.queries == nil || repo.database == nil || service == nil || report == nil {
		return nil, fmt.Errorf("billing webhook worker dependencies required")
	}
	return &WebhookJob{
		repo:    repo,
		service: service,
		report:  report,
	}, nil
}

func (j *WebhookJob) Run(ctx context.Context) error {
	ticker := time.NewTicker(webhookPollInterval)
	defer ticker.Stop()
	cleanupTicker := time.NewTicker(webhookCleanupInterval)
	defer cleanupTicker.Stop()
	if err := j.repo.queries.DeleteOldBillingWebhookEvents(ctx); err != nil && ctx.Err() == nil {
		j.report(ctx, fmt.Errorf("delete old billing webhook receipts: %w", err))
	}

	for {
		worked, err := j.RunOnce(ctx)
		if err != nil && ctx.Err() == nil {
			j.report(ctx, err)
		}
		if worked {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		case <-cleanupTicker.C:
			if err := j.repo.queries.DeleteOldBillingWebhookEvents(ctx); err != nil && ctx.Err() == nil {
				j.report(ctx, fmt.Errorf("delete old billing webhook receipts: %w", err))
			}
		}
	}
}

func (j *WebhookJob) RunOnce(ctx context.Context) (bool, error) {
	if err := j.repo.queries.ReapExhaustedBillingWebhookEvents(
		ctx,
		sqlc.ReapExhaustedBillingWebhookEventsParams{
			Environment: j.service.polar.Environment(),
			MaxAttempts: webhookMaxAttempts,
		},
	); err != nil {
		return false, fmt.Errorf("reap expired billing webhook leases: %w", err)
	}

	lease := uuid.New()
	receipt, err := j.repo.queries.ClaimBillingWebhookEvent(
		ctx,
		sqlc.ClaimBillingWebhookEventParams{
			LeaseToken:  &lease,
			Environment: j.service.polar.Environment(),
			MaxAttempts: webhookMaxAttempts,
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim billing webhook receipt: %w", err)
	}

	if err := j.process(ctx, receipt.ID, lease); err != nil {
		if failErr := j.fail(ctx, receipt, lease, err); failErr != nil {
			return true, errors.Join(err, failErr)
		}
		return true, err
	}
	return true, nil
}

func (j *WebhookJob) process(ctx context.Context, id, lease uuid.UUID) error {
	tx, err := j.repo.database.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin billing webhook transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	queries := j.repo.queries.WithTx(tx)
	receipt, err := queries.GetClaimedBillingWebhookEvent(
		ctx,
		sqlc.GetClaimedBillingWebhookEventParams{
			ID:         id,
			LeaseToken: &lease,
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("lock billing webhook receipt: %w", err)
	}

	transactionalService := *j.service
	transactionalService.repo = j.repo.WithTx(tx)
	transactionalService.subscriptions = subscriptions.NewService(
		subscriptions.NewRepository(queries),
	)
	ignored, err := NewConsumer(&transactionalService).Handle(ctx, receipt.Payload)
	if err != nil {
		return err
	}
	status := "processed"
	if ignored {
		status = "ignored"
	}
	rows, err := queries.CompleteBillingWebhookEvent(
		ctx,
		sqlc.CompleteBillingWebhookEventParams{
			Status:     status,
			ID:         id,
			LeaseToken: &lease,
		},
	)
	if err != nil {
		return fmt.Errorf("complete billing webhook receipt: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("billing webhook lease was lost")
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit billing webhook transaction: %w", err)
	}
	return nil
}

func (j *WebhookJob) fail(
	ctx context.Context,
	receipt sqlc.BillingWebhookEvent,
	lease uuid.UUID,
	processErr error,
) error {
	status := "pending"
	if permanentWebhookError(processErr) || receipt.Attempts >= webhookMaxAttempts {
		status = "dead_letter"
	}
	message := processErr.Error()
	if len(message) > 2_000 {
		message = message[:2_000]
	}
	rows, err := j.repo.queries.FailBillingWebhookEvent(
		ctx,
		sqlc.FailBillingWebhookEventParams{
			Status:       status,
			LastError:    &message,
			RetrySeconds: webhookRetrySeconds(receipt.Attempts),
			ID:           receipt.ID,
			LeaseToken:   &lease,
		},
	)
	if err != nil {
		return fmt.Errorf("schedule billing webhook retry: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("billing webhook lease was lost while scheduling retry")
	}
	return nil
}

func permanentWebhookError(err error) bool {
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Status >= http.StatusBadRequest && appErr.Status < http.StatusInternalServerError
}

func webhookRetrySeconds(attempt int32) int32 {
	delay := int32(1)
	for current := int32(1); current < attempt && delay < 300; current++ {
		delay *= 2
	}
	if delay > 300 {
		return 300
	}
	return delay
}

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
