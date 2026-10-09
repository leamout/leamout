package numbers

import (
	"net/http/httptest"
	"testing"
)

func TestListFilters(t *testing.T) {
	for _, query := range []string{"", "status=active&country_code=gh&voice_enabled=false"} {
		req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateFilters(req); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"status=released", "country_code=GHA", "voice_enabled=0", "trunk_id=bad"} {
		t.Run(query, func(t *testing.T) {
			req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?"+query, nil))
			if err == nil {
				err = validateFilters(req)
			}
			if err == nil {
				t.Fatal("expected invalid filter error")
			}
		})
	}
}

func TestListFiltersPreserveFalseAndNormalizeCountry(t *testing.T) {
	req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?country_code=gh&voice_enabled=false", nil))
	if err != nil {
		t.Fatal(err)
	}
	if req.VoiceEnabled == nil || *req.VoiceEnabled || req.CountryCode == nil || *req.CountryCode != "GH" {
		t.Fatalf("unexpected filters: %+v", req)
	}
}
