package trunks

import (
	"net/http/httptest"
	"testing"
)

func TestListFilters(t *testing.T) {
	for _, query := range []string{"", "status=disabled&direction=inbound&inbound_enabled=false"} {
		req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?"+query, nil))
		if err != nil {
			t.Fatal(err)
		}
		if err := validateFilters(req); err != nil {
			t.Fatal(err)
		}
	}
	for _, query := range []string{"status=removed", "direction=sideways", "inbound_enabled=1"} {
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

func TestListFiltersPreserveFalse(t *testing.T) {
	req, err := parseFilters(httptest.NewRequestWithContext(t.Context(), "GET", "/?inbound_enabled=false", nil))
	if err != nil {
		t.Fatal(err)
	}
	if req.InboundEnabled == nil || *req.InboundEnabled {
		t.Fatal("false must remain an explicit filter")
	}
}
