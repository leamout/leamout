package entitlements

import (
	"errors"
	"testing"
	"time"
)

func TestNormalizeCreateFeature(t *testing.T) {
	t.Parallel()

	enabled := true
	input, err := normalizeCreate(CreateInput{
		Key:     "  recording.enabled  ",
		Kind:    KindFeature,
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("normalize feature: %v", err)
	}
	if input.Key != "recording.enabled" {
		t.Fatalf("expected normalized key, got %q", input.Key)
	}
}

func TestNormalizeCreateRejectsMixedValueShape(t *testing.T) {
	t.Parallel()

	enabled := true
	limit := int64(10)
	_, err := normalizeCreate(CreateInput{
		Key:        "max.calls",
		Kind:       KindFeature,
		Enabled:    &enabled,
		LimitValue: &limit,
	})
	if !errors.Is(err, ErrFeatureValueRequired) {
		t.Fatalf("expected ErrFeatureValueRequired, got %v", err)
	}
}

func TestNormalizeCreateRejectsNegativeLimit(t *testing.T) {
	t.Parallel()

	limit := int64(-1)
	_, err := normalizeCreate(CreateInput{
		Key:        "max.calls",
		Kind:       KindLimit,
		LimitValue: &limit,
	})
	if !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
}

func TestNormalizeCreateRejectsInvalidPeriod(t *testing.T) {
	t.Parallel()

	starts := time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)
	expires := starts.Add(-time.Hour)
	enabled := true
	_, err := normalizeCreate(CreateInput{
		Key:       "voice.enabled",
		Kind:      KindFeature,
		Enabled:   &enabled,
		StartsAt:  &starts,
		ExpiresAt: &expires,
	})
	if !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected ErrInvalidPeriod, got %v", err)
	}
}
