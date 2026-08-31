package entitlements

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

type fakeStore struct {
	plan         []Entitlement
	organization []Entitlement
	license      []Entitlement
}

func (f *fakeStore) CreatePlan(context.Context, uuid.UUID, CreateInput) (Entitlement, error) {
	return Entitlement{}, nil
}
func (f *fakeStore) CreateOrganization(context.Context, uuid.UUID, CreateInput) (Entitlement, error) {
	return Entitlement{}, nil
}
func (f *fakeStore) CreateLicense(context.Context, uuid.UUID, uuid.UUID, CreateInput) (Entitlement, error) {
	return Entitlement{}, nil
}
func (f *fakeStore) ListPlan(context.Context, uuid.UUID) ([]Entitlement, error) {
	return f.plan, nil
}
func (f *fakeStore) ListOrganization(context.Context, uuid.UUID) ([]Entitlement, error) {
	return f.organization, nil
}
func (f *fakeStore) ListLicense(context.Context, uuid.UUID, uuid.UUID) ([]Entitlement, error) {
	return f.license, nil
}
func (f *fakeStore) DeletePlan(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (f *fakeStore) DeleteOrganization(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (f *fakeStore) DeleteLicense(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

type fakeSubscriptions struct {
	current subscriptions.Subscription
	err     error
}

func (f fakeSubscriptions) Current(context.Context, uuid.UUID) (subscriptions.Subscription, error) {
	return f.current, f.err
}

func boolPtr(value bool) *bool       { return &value }
func int64Ptr(value int64) *int64    { return &value }
func timePtr(value time.Time) *time.Time { return &value }

func TestEffectiveForOrganizationAppliesOverrides(t *testing.T) {
	t.Parallel()

	organizationID := uuid.New()
	planID := uuid.New()
	store := &fakeStore{
		plan: []Entitlement{
			{Key: "recording.enabled", Kind: KindFeature, Enabled: boolPtr(true)},
			{Key: "max.concurrent.calls", Kind: KindLimit, LimitValue: int64Ptr(10)},
		},
		organization: []Entitlement{
			{Key: "recording.enabled", Kind: KindFeature, Enabled: boolPtr(false)},
			{Key: "max.concurrent.calls", Kind: KindLimit, LimitValue: int64Ptr(25)},
			{Key: "ai.enabled", Kind: KindFeature, Enabled: boolPtr(true)},
		},
	}
	service := NewService(store, fakeSubscriptions{current: subscriptions.Subscription{PlanID: planID}})

	set, err := service.EffectiveForOrganization(context.Background(), organizationID)
	if err != nil {
		t.Fatalf("resolve entitlements: %v", err)
	}
	if set.Enabled("recording.enabled") {
		t.Fatal("expected organization override to disable recording")
	}
	if !set.Enabled("ai.enabled") {
		t.Fatal("expected organization-only feature to be enabled")
	}
	if limit, ok := set.Limit("max.concurrent.calls"); !ok || limit != 25 {
		t.Fatalf("expected overridden limit 25, got %d, %v", limit, ok)
	}
}

func TestEffectiveForOrganizationFiltersByTime(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	future := at.Add(time.Hour)
	expired := at.Add(-time.Hour)
	planID := uuid.New()
	store := &fakeStore{plan: []Entitlement{
		{Key: "current.enabled", Kind: KindFeature, Enabled: boolPtr(true)},
		{Key: "future.enabled", Kind: KindFeature, Enabled: boolPtr(true), StartsAt: timePtr(future)},
		{Key: "expired.enabled", Kind: KindFeature, Enabled: boolPtr(true), ExpiresAt: timePtr(expired)},
	}}
	service := NewService(store, fakeSubscriptions{current: subscriptions.Subscription{PlanID: planID}})

	set, err := service.EffectiveForOrganizationAt(context.Background(), uuid.New(), at)
	if err != nil {
		t.Fatalf("resolve entitlements: %v", err)
	}
	if !set.Enabled("current.enabled") {
		t.Fatal("expected current entitlement")
	}
	if set.Enabled("future.enabled") || set.Enabled("expired.enabled") {
		t.Fatal("expected future and expired entitlements to be excluded")
	}
}

func TestEffectiveForOrganizationRejectsKindMismatch(t *testing.T) {
	t.Parallel()

	store := &fakeStore{
		plan:         []Entitlement{{Key: "voice", Kind: KindFeature, Enabled: boolPtr(true)}},
		organization: []Entitlement{{Key: "voice", Kind: KindLimit, LimitValue: int64Ptr(1)}},
	}
	service := NewService(store, fakeSubscriptions{current: subscriptions.Subscription{PlanID: uuid.New()}})

	_, err := service.EffectiveForOrganization(context.Background(), uuid.New())
	if !errors.Is(err, ErrKindMismatch) {
		t.Fatalf("expected ErrKindMismatch, got %v", err)
	}
}

func TestEffectiveForOrganizationRequiresSubscription(t *testing.T) {
	t.Parallel()

	service := NewService(&fakeStore{}, fakeSubscriptions{err: subscriptions.ErrSubscriptionNotFound})
	_, err := service.EffectiveForOrganization(context.Background(), uuid.New())
	if !errors.Is(err, ErrSubscriptionUnavailable) {
		t.Fatalf("expected ErrSubscriptionUnavailable, got %v", err)
	}
}

func TestForLicenseUsesOnlyLicenseSnapshot(t *testing.T) {
	t.Parallel()

	store := &fakeStore{license: []Entitlement{
		{Key: "max.deployments", Kind: KindLimit, LimitValue: int64Ptr(3)},
	}}
	service := NewService(store, nil)

	set, err := service.ForLicense(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("resolve license entitlements: %v", err)
	}
	if limit, ok := set.Limit("max.deployments"); !ok || limit != 3 {
		t.Fatalf("expected license snapshot limit 3, got %d, %v", limit, ok)
	}
}
