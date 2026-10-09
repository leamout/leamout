package calls

import (
	"net/http/httptest"
	"testing"
)

func TestListFilters(t *testing.T) {
	for _, query := range []string{"", "direction=inbound&trunk_id=11111111-1111-1111-1111-111111111111&created_from=2026-01-01T00:00:00Z&created_before=2026-02-01T00:00:00Z"} {
		req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateFilters(req); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"direction=sideways", "trunk_id=bad", "voice_agent_id=00000000-0000-0000-0000-000000000000", "created_from=2026-02-01T00:00:00Z&created_before=2026-01-01T00:00:00Z"} {
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

func TestListRequestRetainsExistingPagination(t *testing.T) {
	req, err := listRequest(httptest.NewRequestWithContext(t.Context(), "GET", "/?state=active&direction=inbound&limit=20&offset=10", nil))
	if err != nil {
		t.Fatal(err)
	}
	if req.State == nil || *req.State != "active" || req.Direction == nil || *req.Direction != "inbound" || req.Limit != 20 || req.Offset != 10 {
		t.Fatalf("unexpected request: %+v", req)
	}
	if err := validateListRequest(req); err != nil {
		t.Fatal(err)
	}
}
