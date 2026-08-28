package numbers

import "testing"

func TestNormalizeNumber(t *testing.T) {
	got, err := normalizeNumber("+14155552671")
	if err != nil || got != "+14155552671" {
		t.Fatalf("normalizeNumber() = %q, %v", got, err)
	}
	for _, v := range []string{"14155552671", "+0123", "+1415-555-2671"} {
		if _, err := normalizeNumber(v); err == nil {
			t.Errorf("accepted invalid number %q", v)
		}
	}
}
func TestNormalizeCountryCode(t *testing.T) {
	got, err := normalizeCountryCode(" us ")
	if err != nil || got != "US" {
		t.Fatalf("normalizeCountryCode() = %q, %v", got, err)
	}
	if _, err := normalizeCountryCode("USA"); err == nil {
		t.Fatal("accepted invalid country code")
	}
}
