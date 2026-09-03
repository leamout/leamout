package otp

import (
	"strings"
	"testing"
	"unicode"
)

func TestGenerateNumeric(t *testing.T) {
	for range 32 {
		code, err := GenerateNumeric(6)
		if err != nil {
			t.Fatalf("GenerateNumeric() error = %v", err)
		}
		if len(code) != 6 {
			t.Fatalf("GenerateNumeric() length = %d, want 6", len(code))
		}
		if strings.IndexFunc(code, func(r rune) bool { return !unicode.IsDigit(r) }) >= 0 {
			t.Fatalf("GenerateNumeric() returned non-numeric code %q", code)
		}
	}
}

func TestGenerateNumericRejectsInvalidDigits(t *testing.T) {
	for _, digits := range []int{0, -1, maxNumericDigits + 1} {
		if _, err := GenerateNumeric(digits); err == nil {
			t.Fatalf("GenerateNumeric(%d) unexpectedly succeeded", digits)
		}
	}
}
