package agents

import (
	"net/http/httptest"
	"testing"
)

func TestListFilters(t *testing.T) {
	for _, query := range []string{"", "engine=realtime&language=en"} {
		req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateFilters(req); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"engine=bad", "language=", "engine=composable&engine=realtime"} {
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
