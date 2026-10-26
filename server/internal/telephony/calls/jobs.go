package calls

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

type ReconciliationJobConfig struct {
	Interval time.Duration
}

func DefaultReconciliationJobConfig() ReconciliationJobConfig {
	return ReconciliationJobConfig{
		Interval: 30 * time.Second,
	}
}

type ReconciliationJob struct {
	repo                   *Repository
	service                *Service
	config                 ReconciliationJobConfig
	failures               atomic.Uint64
	cleanupRetries         atomic.Uint64
	unconfirmedChecks      atomic.Uint64
	recoveredOriginations   atomic.Uint64
}

func NewReconciliationJob(
	repo *Repository,
	service *Service,
	config ReconciliationJobConfig,
) (*ReconciliationJob, error) {
	if repo == nil {
		return nil, fmt.Errorf("call reconciliation repository is required")
	}
	if service == nil {
		return nil, fmt.Errorf("call reconciliation service is required")
	}
	if config.Interval <= 0 {
		config.Interval = 30 * time.Second
	}

	return &ReconciliationJob{
		repo:    repo,
		service: service,
		config:  config,
	}, nil
}

func (j *ReconciliationJob) Run(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("call reconciliation context is required")
	}

	j.runPass(ctx)

	ticker := time.NewTicker(j.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			j.runPass(ctx)
		}
	}
}

func (j *ReconciliationJob) runPass(ctx context.Context) {
	if err := j.Reconcile(ctx); err != nil && !errors.Is(err, context.Canceled) {
		j.failures.Add(1)
		log.Printf("call reconciliation pass failed: %v", err)
	}
}

func (j *ReconciliationJob) Reconcile(ctx context.Context) error {
	active, err := j.repo.ListActiveForAdmissionReconciliation(ctx)
	if err != nil {
		return fmt.Errorf("list calls for admission reconciliation: %w", err)
	}

	for _, call := range active {
		if err := j.service.admission.Refresh(ctx, call.TrunkID, call.ID); err != nil {
			return fmt.Errorf("refresh trunk admission lease for call %s: %w", call.ID, err)
		}
		if err := j.service.admission.RefreshOrganization(
			ctx,
			call.OrganizationID,
			call.ID,
		); err != nil {
			return fmt.Errorf("refresh organization admission lease for call %s: %w", call.ID, err)
		}
	}

	if err := j.reconcileOriginations(ctx); err != nil {
		return fmt.Errorf("reconcile outbound originations: %w", err)
	}
	if err := j.reconcileTerminalCleanup(ctx); err != nil {
		return fmt.Errorf("reconcile terminal admission cleanup: %w", err)
	}

	return nil
}

// Reconcile only from positive channel evidence. Absence alone cannot prove
// that a queued bgapi originate never ran, so no automatic redial or release.
func (j *ReconciliationJob) reconcileOriginations(ctx context.Context) error {
	attempts, err := j.repo.ListUnresolvedOriginations(ctx)
	if err != nil {
		return err
	}
	for _, attempt := range attempts {
		channelID := attempt.ChannelID
		// A disconnected FreeSWITCH client cannot establish channel absence.
		exists, err := j.service.controller.ChannelExists(ctx, channelID)
		if err != nil {
			return fmt.Errorf("query originate channel %s: %w", channelID, err)
		}
		if !exists {
			j.unconfirmedChecks.Add(1)
			if err := j.repo.DeferOriginationCheck(ctx, channelID); err != nil {
				return fmt.Errorf("defer unresolved originate channel %s: %w", channelID, err)
			}
			continue
		}
		if err := j.service.channels.Bind(ctx, attempt.CallID, channelID); err != nil {
			return fmt.Errorf("bind recovered channel %s: %w", channelID, err)
		}
		if err := j.repo.MarkOriginationObserved(ctx, attempt.CallID, channelID); err != nil {
			return fmt.Errorf("mark recovered channel %s: %w", channelID, err)
		}
		j.recoveredOriginations.Add(1)
	}
	return nil
}

// Successful channel discovery does not imply the call was answered; only
// lifecycle events may advance the call state.

// Terminal rows are durable even if the lifecycle handler or Redis restarts.
// Redis ZREM and channel deletion are safe to repeat. Only delete the durable
// cleanup marker once every required release has succeeded.
func (j *ReconciliationJob) reconcileTerminalCleanup(ctx context.Context) error {
	pending, err := j.repo.ListDueAdmissionCleanup(ctx)
	if err != nil {
		return err
	}
	for _, item := range pending {
		if err := j.service.cleanupTerminalAdmission(ctx, item); err != nil {
			j.cleanupRetries.Add(1)
			if deferErr := j.repo.DeferAdmissionCleanup(ctx, item.CallID); deferErr != nil {
				return errors.Join(err, fmt.Errorf("defer terminal cleanup for %s: %w", item.CallID, deferErr))
			}
			log.Printf("retry terminal call cleanup for %s: %v", item.CallID, err)
		}
	}
	return nil
}

// ReliabilityStats is a point-in-time view of durable reconciliation backlogs
// plus in-process counters. The backlog queries are unpaginated counts rather
// than the 100-row work queue, so saturation remains visible.
type ReliabilityStats struct {
	UnresolvedOriginations int64
	PendingAdmissionCleanup int64
	ReconciliationFailures uint64
	CleanupRetries uint64
	UnconfirmedChecks uint64
	RecoveredOriginations uint64
}

func (j *ReconciliationJob) ReliabilityStats(ctx context.Context) (ReliabilityStats, error) {
	unresolved, err := j.repo.CountUnresolvedOriginations(ctx)
	if err != nil {
		return ReliabilityStats{}, fmt.Errorf("count unresolved originations: %w", err)
	}
	pending, err := j.repo.CountPendingAdmissionCleanup(ctx)
	if err != nil {
		return ReliabilityStats{}, fmt.Errorf("count pending admission cleanup: %w", err)
	}
	return ReliabilityStats{
		UnresolvedOriginations:  unresolved,
		PendingAdmissionCleanup: pending,
		ReconciliationFailures:  j.failures.Load(),
		CleanupRetries:          j.cleanupRetries.Load(),
		UnconfirmedChecks:       j.unconfirmedChecks.Load(),
		RecoveredOriginations:   j.recoveredOriginations.Load(),
	}, nil
}
