package commpeak

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListTerminationCDRsUsesAuthorizationAndFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/call_records/termination" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "token-value" {
			t.Fatalf("Authorization = %q", got)
		}
		query := r.URL.Query()
		if query.Get("time_range") != "2026-09-01 - 2026-09-04" {
			t.Fatalf("time_range = %q", query.Get("time_range"))
		}
		if query.Get("destination") != "233" || query.Get("sip_account_id") != "account-1" {
			t.Fatalf("query = %v", query)
		}
		if query.Get("page") != "2" || query.Get("per_page") != "500" {
			t.Fatalf("pagination = %v", query)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"cdr-1"}],"page":2}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:       server.URL,
		Authorization: "token-value",
		HTTPClient:    server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := client.ListTerminationCDRs(context.Background(), CDRRequest{
		TimeRange:    "2026-09-01 - 2026-09-04",
		Destination:  "233",
		SIPAccountID: "account-1",
		Page:         2,
		PerPage:      500,
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.Direction != CDRDirectionTermination || len(page.Raw) == 0 {
		t.Fatalf("page = %+v", page)
	}
}

func TestListOriginationCDRsUsesDIDArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/call_records/origination" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		dids := r.URL.Query()["did[]"]
		if len(dids) != 2 || dids[0] != "+233201111111" || dids[1] != "+233202222222" {
			t.Fatalf("dids = %v", dids)
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:       server.URL,
		Authorization: "token-value",
		HTTPClient:    server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListOriginationCDRs(context.Background(), CDRRequest{
		DIDs: []string{"+233201111111", "+233202222222"},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestCDRRequestRejectsInvalidPageSize(t *testing.T) {
	client, err := NewClient(Config{Authorization: "token-value"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListTerminationCDRs(context.Background(), CDRRequest{PerPage: 1001}); err == nil {
		t.Fatal("expected validation error")
	}
}
