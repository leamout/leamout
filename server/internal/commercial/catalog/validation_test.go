package catalog

import (
	"errors"
	"testing"

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
