package entitlements

import (
	"errors"
	"testing"
	"time"
)

func boolPtr(value bool) *bool           { return &value }
func int64Ptr(value int64) *int64        { return &value }
func timePtr(value time.Time) *time.Time { return &value }

func TestResolveAppliesOverrides(t *testing.T) {
	t.Parallel()

	base := []Entitlement{
		{Key: "recording.enabled", Kind: KindFeature, Enabled: boolPtr(true)},
		{Key: "max.concurrent.calls", Kind: KindLimit, LimitValue: int64Ptr(10)},
	}
	overrides := []Entitlement{
		{Key: "recording.enabled", Kind: KindFeature, Enabled: boolPtr(false)},
		{Key: "max.concurrent.calls", Kind: KindLimit, LimitValue: int64Ptr(25)},
		{Key: "ai.enabled", Kind: KindFeature, Enabled: boolPtr(true)},
	}

	set, err := resolve(time.Now(), base, overrides)
	if err != nil {
		t.Fatalf("resolve entitlements: %v", err)
	}
	if set.Enabled("recording.enabled") {
		t.Fatal("expected override to disable recording")
	}
	if !set.Enabled("ai.enabled") {
		t.Fatal("expected organization-only feature to be enabled")
	}
	if limit, ok := set.Limit("max.concurrent.calls"); !ok || limit != 25 {
		t.Fatalf("expected overridden limit 25, got %d, %v", limit, ok)
	}
}

func TestResolveFiltersByTime(t *testing.T) {
	t.Parallel()

	at := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	future := at.Add(time.Hour)
	expired := at.Add(-time.Hour)
	base := []Entitlement{
		{Key: "current.enabled", Kind: KindFeature, Enabled: boolPtr(true)},
		{Key: "future.enabled", Kind: KindFeature, Enabled: boolPtr(true), StartsAt: timePtr(future)},
		{Key: "expired.enabled", Kind: KindFeature, Enabled: boolPtr(true), ExpiresAt: timePtr(expired)},
	}

	set, err := resolve(at, base, nil)
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

func TestResolveRejectsKindMismatch(t *testing.T) {
	t.Parallel()

	base := []Entitlement{{Key: "voice", Kind: KindFeature, Enabled: boolPtr(true)}}
	overrides := []Entitlement{{Key: "voice", Kind: KindLimit, LimitValue: int64Ptr(1)}}

	_, err := resolve(time.Now(), base, overrides)
	if !errors.Is(err, ErrKindMismatch) {
		t.Fatalf("expected ErrKindMismatch, got %v", err)
	}
}

func TestResolveRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	_, err := resolve(time.Now(), []Entitlement{{Key: "voice", Kind: KindFeature}}, nil)
	if !errors.Is(err, ErrInvalidEntitlement) {
		t.Fatalf("expected ErrInvalidEntitlement, got %v", err)
	}
}
