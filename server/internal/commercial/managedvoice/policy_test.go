package managedvoice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/telecom/calls"
)

type entitlementStub struct {
	set entitlements.EntitlementSet
	err error
}

func (s entitlementStub) EffectiveForOrganization(context.Context, uuid.UUID) (entitlements.EntitlementSet, error) {
	return s.set, s.err
}

type spendStub struct {
	amount int64
	err    error
}

func (s spendStub) ManagedSpendMicros(context.Context, uuid.UUID, time.Time) (int64, error) {
	return s.amount, s.err
}

func TestPolicyRequiresEntitlementAndAvailableSpend(t *testing.T) {
	set := entitlements.EntitlementSet{Features: map[entitlements.Feature]bool{FeatureEnabled: true}, Limits: map[string]int64{LimitDailySpendMicros: 1000}}
	policy, _ := NewPolicy(entitlementStub{set: set}, spendStub{amount: 999})
	policy.now = func() time.Time { return time.Date(2026, 9, 4, 12, 0, 0, 0, time.UTC) }
	if err := policy.AuthorizeManagedOutbound(context.Background(), uuid.New()); err != nil {
		t.Fatal(err)
	}
	policy.spend = spendStub{amount: 1000}
	if err := policy.AuthorizeManagedOutbound(context.Background(), uuid.New()); !errors.Is(err, calls.ErrManagedSpendLimit) {
		t.Fatalf("got %v", err)
	}
	set.Features[FeatureEnabled] = false
	policy.entitlements = entitlementStub{set: set}
	if err := policy.AuthorizeManagedOutbound(context.Background(), uuid.New()); !errors.Is(err, calls.ErrManagedVoiceDisabled) {
		t.Fatalf("got %v", err)
	}
}
