package state

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

// Service resolves the current organization commercial state from its subscription and entitlements.
type Service struct {
	subscriptions *subscriptions.Service
	entitlements  *entitlements.Service
	now           func() time.Time
}

func NewService(subscriptions *subscriptions.Service, entitlements *entitlements.Service) *Service {
	return &Service{
		subscriptions: subscriptions,
		entitlements:  entitlements,
		now:           time.Now,
	}
}

func (s *Service) Resolve(ctx context.Context, organizationID uuid.UUID) (OrganizationState, error) {
	return s.ResolveAt(ctx, organizationID, s.now())
}

func (s *Service) ResolveAt(ctx context.Context, organizationID uuid.UUID, at time.Time) (OrganizationState, error) {
	if organizationID == uuid.Nil {
		return OrganizationState{}, ErrOrganizationIDRequired
	}

	current, err := s.subscriptions.Current(ctx, organizationID)
	if err != nil {
		if errors.Is(err, subscriptions.ErrSubscriptionNotFound) {
			return unsubscribedState(organizationID, at), nil
		}
		return OrganizationState{}, err
	}

	resolution, err := s.entitlements.ResolveForOrganizationPlanAt(ctx, organizationID, current.PlanID, at)
	if err != nil {
		return OrganizationState{}, err
	}

	return organizationState(organizationID, current, resolution, at), nil
}

func organizationState(organizationID uuid.UUID, current subscriptions.Subscription, resolution entitlements.Resolution, at time.Time) OrganizationState {
	subscriptionID := current.ID
	planID := current.PlanID
	return OrganizationState{
		OrganizationID: organizationID,
		Standing:       standingFromSubscription(current.Status),
		SubscriptionID: &subscriptionID,
		PlanID:         &planID,
		Features:       cloneFeatures(resolution.Set.Features),
		Limits:         cloneLimits(resolution.Set.Limits),
		EffectiveAt:    at,
		NextChangeAt:   earliestFuture(at, resolution.NextChangeAt, current.EndsAt),
	}
}

func unsubscribedState(organizationID uuid.UUID, at time.Time) OrganizationState {
	return OrganizationState{
		OrganizationID: organizationID,
		Standing:       StandingUnsubscribed,
		Features:       map[string]bool{},
		Limits:         map[string]int64{},
		EffectiveAt:    at,
	}
}

func standingFromSubscription(status subscriptions.Status) Standing {
	if status == subscriptions.StatusPastDue {
		return StandingPastDue
	}
	return StandingActive
}

func earliestFuture(at time.Time, values ...*time.Time) *time.Time {
	var next *time.Time
	for _, value := range values {
		if value == nil || !value.After(at) {
			continue
		}
		if next == nil || value.Before(*next) {
			candidate := *value
			next = &candidate
		}
	}
	return next
}

func cloneFeatures(features map[entitlements.Feature]bool) map[string]bool {
	cloned := make(map[string]bool, len(features))
	for feature, enabled := range features {
		cloned[string(feature)] = enabled
	}
	return cloned
}

func cloneLimits(limits map[string]int64) map[string]int64 {
	cloned := make(map[string]int64, len(limits))
	for name, value := range limits {
		cloned[name] = value
	}
	return cloned
}
