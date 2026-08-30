package entitlements

import (
	"errors"
	"testing"
	"time"
)

func TestValidateAndActiveAt(t *testing.T) {
	startsAt := time.Date(2026, time.August, 30, 0, 0, 0, 0, time.UTC)
	expiresAt := startsAt.Add(time.Hour)
	enabled := true
	entitlement := Entitlement{
		Key: "voice.recording", Kind: KindFeature, Enabled: &enabled,
		StartsAt: &startsAt, ExpiresAt: &expiresAt,
	}
	if err := Validate(entitlement); err != nil {
		t.Fatal(err)
	}
	if entitlement.ActiveAt(startsAt.Add(-time.Nanosecond)) {
		t.Fatal("entitlement was active before its start")
	}
	if !entitlement.ActiveAt(startsAt) {
		t.Fatal("entitlement was not active at its start")
	}
	if entitlement.ActiveAt(expiresAt) {
		t.Fatal("entitlement was active at its expiration")
	}
}

func TestValidateRejectsMismatchedValue(t *testing.T) {
	enabled := true
	limit := int64(1)
	err := Validate(Entitlement{Key: "voice.calls", Kind: KindLimit, Enabled: &enabled, Limit: &limit})
	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("expected invalid value, got %v", err)
	}
}
