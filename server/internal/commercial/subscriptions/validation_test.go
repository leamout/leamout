package subscriptions

import (
	"errors"
	"testing"
	"time"
)

func TestValidateTransition(t *testing.T) {
	t.Parallel()

	allowed := [][2]Status{
		{StatusPending, StatusActive},
		{StatusActive, StatusPastDue},
		{StatusActive, StatusCancelled},
		{StatusActive, StatusExpired},
		{StatusPastDue, StatusActive},
		{StatusPastDue, StatusCancelled},
		{StatusPastDue, StatusExpired},
		{StatusActive, StatusActive},
	}
	for _, transition := range allowed {
		if err := validateTransition(transition[0], transition[1]); err != nil {
			t.Fatalf("expected %s -> %s to be valid: %v", transition[0], transition[1], err)
		}
	}

	blocked := [][2]Status{
		{StatusPending, StatusPastDue},
		{StatusCancelled, StatusActive},
		{StatusExpired, StatusActive},
	}
	for _, transition := range blocked {
		if err := validateTransition(transition[0], transition[1]); !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("expected %s -> %s to fail with ErrInvalidTransition, got %v", transition[0], transition[1], err)
		}
	}
}

func TestNormalizeProvider(t *testing.T) {
	t.Parallel()

	got, err := normalizeProvider(ProviderReference{Provider: " Stripe ", SubscriptionID: " sub_123 "})
	if err != nil {
		t.Fatalf("normalize provider: %v", err)
	}
	if got.Provider != "stripe" || got.SubscriptionID != "sub_123" {
		t.Fatalf("unexpected normalized provider: %#v", got)
	}
}

func TestValidatePeriod(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	renew := start.Add(30 * 24 * time.Hour)
	end := start.Add(60 * 24 * time.Hour)
	if err := validatePeriod(start, &renew, &end); err != nil {
		t.Fatalf("expected valid period: %v", err)
	}

	badRenew := start.Add(-time.Hour)
	if err := validatePeriod(start, &badRenew, &end); !errors.Is(err, ErrInvalidPeriod) {
		t.Fatalf("expected invalid renewal period, got %v", err)
	}
}
