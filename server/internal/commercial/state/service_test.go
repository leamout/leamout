package state

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

type fakeSubscriptions struct {
	current subscriptions.Subscription
	err     error
}

func (f fakeSubscriptions) Current(context.Context, uuid.UUID) (subscriptions.Subscription, error) {
	return f.current, f.err
}

type fakeEntitlements struct {
	set            entitlements.EntitlementSet
	err            error
	organizationID uuid.UUID
	planID         uuid.UUID
	at             time.Time
}

func (f *fakeEntitlements) EffectiveForOrganizationPlanAt(_ context.Context, organizationID, planID uuid.UUID, at time.Time) (entitlements.EntitlementSet, error) {
	f.organizationID = organizationID
	f.planID = planID
	f.at = at
	return f.set, f.err
}

func TestResolveBuildsOrganizationState(t *testing.T) {
	t.Parallel()

	organizationID := uuid.New()
	subscriptionID := uuid.New()
	planID := uuid.New()
	at := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	set := entitlements.EntitlementSet{
		Features: map[entitlements.Feature]bool{"recording.enabled": true},
		Limits:   map[string]int64{"max.concurrent.calls": 25},
	}
	resolvedEntitlements := &fakeEntitlements{set: set}
	service := NewService(fakeSubscriptions{current: subscriptions.Subscription{
		ID:             subscriptionID,
		OrganizationID: organizationID,
		PlanID:         planID,
		Status:         subscriptions.StatusActive,
	}}, resolvedEntitlements)

	resolved, err := service.ResolveAt(context.Background(), organizationID, at)
	if err != nil {
		t.Fatalf("ResolveAt() error = %v", err)
	}
	if resolved.OrganizationID != organizationID {
		t.Fatalf("OrganizationID = %v, want %v", resolved.OrganizationID, organizationID)
	}
	if resolved.SubscriptionID != subscriptionID {
		t.Fatalf("SubscriptionID = %v, want %v", resolved.SubscriptionID, subscriptionID)
	}
	if resolved.PlanID != planID {
		t.Fatalf("PlanID = %v, want %v", resolved.PlanID, planID)
	}
	if resolved.EffectiveAt != at {
		t.Fatalf("EffectiveAt = %v, want %v", resolved.EffectiveAt, at)
	}
	if !resolved.Enabled("recording.enabled") {
		t.Fatal("expected recording.enabled to be enabled")
	}
	if limit, ok := resolved.Limit("max.concurrent.calls"); !ok || limit != 25 {
		t.Fatalf("Limit(max.concurrent.calls) = %d, %v, want 25, true", limit, ok)
	}
	if resolvedEntitlements.organizationID != organizationID || resolvedEntitlements.planID != planID || resolvedEntitlements.at != at {
		t.Fatal("expected entitlement resolution to use the current subscription plan and evaluation time")
	}

	resolved.Features["recording.enabled"] = false
	resolved.Limits["max.concurrent.calls"] = 1
	if !set.Features["recording.enabled"] || set.Limits["max.concurrent.calls"] != 25 {
		t.Fatal("resolved state must not alias entitlement maps")
	}
}

func TestResolveMapsMissingSubscriptionToUnavailable(t *testing.T) {
	t.Parallel()

	service := NewService(fakeSubscriptions{err: subscriptions.ErrSubscriptionNotFound}, &fakeEntitlements{})
	_, err := service.Resolve(context.Background(), uuid.New())
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("Resolve() error = %v, want %v", err, ErrUnavailable)
	}
}

func TestResolveRequiresOrganizationID(t *testing.T) {
	t.Parallel()

	service := NewService(fakeSubscriptions{}, &fakeEntitlements{})
	_, err := service.Resolve(context.Background(), uuid.Nil)
	if !errors.Is(err, ErrOrganizationIDRequired) {
		t.Fatalf("Resolve() error = %v, want %v", err, ErrOrganizationIDRequired)
	}
}

func TestResolvePropagatesEntitlementError(t *testing.T) {
	t.Parallel()

	want := errors.New("resolve entitlements")
	service := NewService(fakeSubscriptions{current: subscriptions.Subscription{
		ID:     uuid.New(),
		PlanID: uuid.New(),
	}}, &fakeEntitlements{err: want})

	_, err := service.Resolve(context.Background(), uuid.New())
	if !errors.Is(err, want) {
		t.Fatalf("Resolve() error = %v, want %v", err, want)
	}
}
