package commpeak

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPollCDRsUsesUTCDateWindowAndCountsEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/call_records/termination" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		query := r.URL.Query()
		if got := query.Get("time_range"); got != "2026-09-08 - 2026-09-08" {
			t.Fatalf("time_range = %q", got)
		}
		if query.Get("page") != "3" || query.Get("per_page") != "1000" {
			t.Fatalf("pagination = %v", query)
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"a"},{"id":"b"}]}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, Authorization: "token", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	payload, count, err := client.PollCDRs(
		context.Background(),
		"termination",
		time.Date(2026, 9, 8, 23, 0, 0, 0, time.FixedZone("test", 2*60*60)),
		3,
		1000,
	)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || len(payload) == 0 {
		t.Fatalf("count=%d payload=%s", count, payload)
	}
}
