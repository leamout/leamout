package access

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

func (s *Service) Resolve(ctx context.Context, organizationID uuid.UUID) (OrganizationAccess, error) {
	return s.ResolveAt(ctx, organizationID, s.now())
}

func (s *Service) ResolveAt(ctx context.Context, organizationID uuid.UUID, at time.Time) (OrganizationAccess, error) {
	if organizationID == uuid.Nil {
		return OrganizationAccess{}, ErrOrganizationIDRequired
	}

	current, err := s.subscriptions.Current(ctx, organizationID)
	if err != nil {
		if errors.Is(err, subscriptions.ErrSubscriptionNotFound) {
			return unsubscribedAccess(organizationID, at), nil
		}
		return OrganizationAccess{}, err
	}

	resolution, err := s.entitlements.ResolveForOrganizationPlanAt(ctx, organizationID, current.PlanID, at)
	if err != nil {
		return OrganizationAccess{}, err
	}

	return organizationAccess(organizationID, current, resolution, at), nil
}

func organizationAccess(organizationID uuid.UUID, current subscriptions.Subscription, resolution entitlements.Resolution, at time.Time) OrganizationAccess {
	subscriptionID := current.ID
	planID := current.PlanID
	return OrganizationAccess{
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

func unsubscribedAccess(organizationID uuid.UUID, at time.Time) OrganizationAccess {
	return OrganizationAccess{
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
