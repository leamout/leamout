package auth

import (
	"strings"
	"testing"
)

func TestValidateNewPassword(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		value string
		valid bool
	}{
		{"empty", "", false},
		{"too short", "short", false},
		{"minimum length", "a secure passphrase", true},
		{"spaces allowed", "   12 chars  ", true},
		{"maximum", strings.Repeat("z", 1024), true},
		{"too long", strings.Repeat("z", 1025), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateNewPassword(tc.value)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}
