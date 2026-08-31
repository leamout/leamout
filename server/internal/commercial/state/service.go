package state

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

type subscriptionReader interface {
	Current(context.Context, uuid.UUID) (subscriptions.Subscription, error)
}

type entitlementReader interface {
	EffectiveForOrganizationPlanAt(context.Context, uuid.UUID, uuid.UUID, time.Time) (entitlements.EntitlementSet, error)
}

// Service resolves the current organization commercial state from its subscription and entitlements.
type Service struct {
	subscriptions subscriptionReader
	entitlements  entitlementReader
	now           func() time.Time
}

func NewService(subscriptions subscriptionReader, entitlements entitlementReader) *Service {
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
			return OrganizationState{}, ErrUnavailable
		}
		return OrganizationState{}, err
	}

	set, err := s.entitlements.EffectiveForOrganizationPlanAt(ctx, organizationID, current.PlanID, at)
	if err != nil {
		return OrganizationState{}, err
	}

	return OrganizationState{
		OrganizationID: organizationID,
		SubscriptionID: current.ID,
		PlanID:         current.PlanID,
		Features:       cloneFeatures(set.Features),
		Limits:         cloneLimits(set.Limits),
		EffectiveAt:    at,
	}, nil
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
