package catalog

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNormalizeCodeTrimsWhitespace(t *testing.T) {
	got, err := normalizeCode("  voice-pro  ")
	if err != nil {
		t.Fatalf("normalizeCode() error = %v", err)
	}
	if got != "voice-pro" {
		t.Fatalf("normalizeCode() = %q, want %q", got, "voice-pro")
	}
}

func TestNormalizeCodeRejectsWhitespace(t *testing.T) {
	_, err := normalizeCode("voice pro")
	if !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("normalizeCode() error = %v, want %v", err, ErrInvalidCode)
	}
}

func TestNormalizeIDRequiresID(t *testing.T) {
	if err := normalizeID(uuid.Nil); !errors.Is(err, ErrIDRequired) {
		t.Fatalf("normalizeID() error = %v, want %v", err, ErrIDRequired)
	}
}

func TestPriceEffectiveAt(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	end := start.Add(30 * 24 * time.Hour)
	price := Price{Active: true, EffectiveFrom: start, EffectiveUntil: &end}

	if price.EffectiveAt(start.Add(-time.Second)) {
		t.Fatal("price must not be available before effective_from")
	}
	if !price.EffectiveAt(start) {
		t.Fatal("price must be available at effective_from")
	}
	if price.EffectiveAt(end) {
		t.Fatal("price must not be available at effective_until")
	}

	price.Active = false
	if price.EffectiveAt(start) {
		t.Fatal("inactive price must not be available")
	}
}
