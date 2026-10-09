package audit

import (
	"net/http/httptest"
	"testing"
)

func TestListFilters(t *testing.T) {
	for _, query := range []string{"", "action=call.created&actor_id=11111111-1111-1111-1111-111111111111&occurred_from=2026-01-01T00:00:00Z"} {
		req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateFilters(req); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"actor_type=admin", "actor_id=bad", "target_id=00000000-0000-0000-0000-000000000000", "occurred_from=2026-01-01T00:00:00Z&occurred_before=2026-01-01T00:00:00Z"} {
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
