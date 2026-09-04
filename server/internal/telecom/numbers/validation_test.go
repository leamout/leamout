package numbers

import "testing"

func TestNormalizeNumber(t *testing.T) {
	for _, v := range []string{"+1234567", "+14155552671"} {
		got, err := normalizeNumber(v)
		if err != nil || got != v {
			t.Fatalf("normalizeNumber(%q) = %q, %v", v, got, err)
		}
	}

	for _, v := range []string{"1234567", "+0123456", "+123456", "+1415-555-2671"} {
		if _, err := normalizeNumber(v); err == nil {
			t.Errorf("accepted invalid number %q", v)
		}
	}
}

func TestNormalizeCountryCode(t *testing.T) {
	got, err := normalizeCountryCode(" gh ")
	if err != nil || got != "GH" {
		t.Fatalf("normalizeCountryCode() = %q, %v", got, err)
	}
	if _, err := normalizeCountryCode("GHA"); err == nil {
		t.Fatal("accepted invalid country code")
	}
}
