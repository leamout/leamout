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

	resolved := organizationState(organizationID, current, set, at)
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
}

func TestOrganizationStateDoesNotAliasEntitlementMaps(t *testing.T) {
	t.Parallel()

	set := entitlements.EntitlementSet{
		Features: map[entitlements.Feature]bool{"recording.enabled": true},
		Limits:   map[string]int64{"max.concurrent.calls": 25},
	}
	resolved := organizationState(uuid.New(), subscriptions.Subscription{ID: uuid.New(), PlanID: uuid.New()}, set, time.Now())

	resolved.Features["recording.enabled"] = false
	resolved.Limits["max.concurrent.calls"] = 1

	if !set.Features["recording.enabled"] {
		t.Fatal("resolved features must not alias entitlement features")
	}
	if set.Limits["max.concurrent.calls"] != 25 {
		t.Fatal("resolved limits must not alias entitlement limits")
	}
}
