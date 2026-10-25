package calls

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	repo    *Repository
	service *Service
	config  ReconciliationJobConfig
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
			continue
		}
		if err := j.service.channels.Bind(ctx, attempt.CallID, channelID); err != nil {
			return fmt.Errorf("bind recovered channel %s: %w", channelID, err)
		}
		if err := j.repo.MarkOriginationObserved(ctx, attempt.CallID, channelID); err != nil {
			return fmt.Errorf("mark recovered channel %s: %w", channelID, err)
		}
	}
	return nil
}

// Successful channel discovery does not imply the call was answered; only
// lifecycle events may advance the call state.
