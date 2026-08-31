package state

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

func TestOrganizationStateBuildsResolvedState(t *testing.T) {
	t.Parallel()

	organizationID := uuid.New()
	subscriptionID := uuid.New()
	planID := uuid.New()
	at := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	set := entitlements.EntitlementSet{
		Features: map[entitlements.Feature]bool{"recording.enabled": true},
		Limits:   map[string]int64{"max.concurrent.calls": 25},
	}
	current := subscriptions.Subscription{
		ID:             subscriptionID,
		OrganizationID: organizationID,
		PlanID:         planID,
		Status:         subscriptions.StatusActive,
	}

	resolved := organizationState(organizationID, current, entitlements.Resolution{Set: set}, at)
	if resolved.OrganizationID != organizationID {
		t.Fatalf("OrganizationID = %v, want %v", resolved.OrganizationID, organizationID)
	}
	if resolved.Standing != StandingActive {
		t.Fatalf("Standing = %q, want %q", resolved.Standing, StandingActive)
	}
	if resolved.SubscriptionID == nil || *resolved.SubscriptionID != subscriptionID {
		t.Fatalf("SubscriptionID = %v, want %v", resolved.SubscriptionID, subscriptionID)
	}
	if resolved.PlanID == nil || *resolved.PlanID != planID {
		t.Fatalf("PlanID = %v, want %v", resolved.PlanID, planID)
	}
	if resolved.EffectiveAt != at {
		t.Fatalf("EffectiveAt = %v, want %v", resolved.EffectiveAt, at)
	}
	if resolved.NextChangeAt != nil {
		t.Fatalf("NextChangeAt = %v, want nil", resolved.NextChangeAt)
	}
	if !resolved.Enabled("recording.enabled") {
		t.Fatal("expected recording.enabled to be enabled")
	}
	if limit, ok := resolved.Limit("max.concurrent.calls"); !ok || limit != 25 {
		t.Fatalf("Limit(max.concurrent.calls) = %d, %v, want 25, true", limit, ok)
	}
}

func TestOrganizationStateUsesEarliestKnownChange(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	entitlementChange := at.Add(time.Hour)
	subscriptionEnd := at.Add(2 * time.Hour)
	resolved := organizationState(
		uuid.New(),
		subscriptions.Subscription{
			ID:     uuid.New(),
			PlanID: uuid.New(),
			Status: subscriptions.StatusActive,
			EndsAt: &subscriptionEnd,
		},
		entitlements.Resolution{
			Set: entitlements.EntitlementSet{
				Features: map[entitlements.Feature]bool{},
				Limits:   map[string]int64{},
			},
			NextChangeAt: &entitlementChange,
		},
		at,
	)

	if resolved.NextChangeAt == nil || !resolved.NextChangeAt.Equal(entitlementChange) {
		t.Fatalf("NextChangeAt = %v, want %v", resolved.NextChangeAt, entitlementChange)
	}
}

func TestOrganizationStateDoesNotAliasEntitlementMaps(t *testing.T) {
	t.Parallel()

	set := entitlements.EntitlementSet{
		Features: map[entitlements.Feature]bool{"recording.enabled": true},
		Limits:   map[string]int64{"max.concurrent.calls": 25},
	}
	resolved := organizationState(
		uuid.New(),
		subscriptions.Subscription{ID: uuid.New(), PlanID: uuid.New(), Status: subscriptions.StatusActive},
		entitlements.Resolution{Set: set},
		time.Now(),
	)

	resolved.Features["recording.enabled"] = false
	resolved.Limits["max.concurrent.calls"] = 1

	if !set.Features["recording.enabled"] {
		t.Fatal("resolved features must not alias entitlement features")
	}
	if set.Limits["max.concurrent.calls"] != 25 {
		t.Fatal("resolved limits must not alias entitlement limits")
	}
}

func TestUnsubscribedStateIsKnownCommercialState(t *testing.T) {
	t.Parallel()

	organizationID := uuid.New()
	at := time.Date(2026, 8, 31, 11, 0, 0, 0, time.UTC)

	resolved := unsubscribedState(organizationID, at)
	if resolved.OrganizationID != organizationID {
		t.Fatalf("OrganizationID = %v, want %v", resolved.OrganizationID, organizationID)
	}
	if resolved.Standing != StandingUnsubscribed {
		t.Fatalf("Standing = %q, want %q", resolved.Standing, StandingUnsubscribed)
	}
	if resolved.SubscriptionID != nil {
		t.Fatalf("SubscriptionID = %v, want nil", resolved.SubscriptionID)
	}
	if resolved.PlanID != nil {
		t.Fatalf("PlanID = %v, want nil", resolved.PlanID)
	}
	if len(resolved.Features) != 0 {
		t.Fatalf("Features = %#v, want empty", resolved.Features)
	}
	if len(resolved.Limits) != 0 {
		t.Fatalf("Limits = %#v, want empty", resolved.Limits)
	}
	if resolved.EffectiveAt != at {
		t.Fatalf("EffectiveAt = %v, want %v", resolved.EffectiveAt, at)
	}
	if resolved.NextChangeAt != nil {
		t.Fatalf("NextChangeAt = %v, want nil", resolved.NextChangeAt)
	}
}

func TestStandingFromSubscription(t *testing.T) {
	t.Parallel()

	if got := standingFromSubscription(subscriptions.StatusActive); got != StandingActive {
		t.Fatalf("active standing = %q, want %q", got, StandingActive)
	}
	if got := standingFromSubscription(subscriptions.StatusPastDue); got != StandingPastDue {
		t.Fatalf("past_due standing = %q, want %q", got, StandingPastDue)
	}
}
