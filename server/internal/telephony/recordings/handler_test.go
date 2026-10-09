package recordings

import (
	"net/http/httptest"
	"testing"
)

func TestListFilters(t *testing.T) {
	for _, query := range []string{"", "status=completed&call_id=11111111-1111-1111-1111-111111111111"} {
		req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateFilters(req); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"status=deleted", "call_id=bad", "created_from=bad"} {
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
