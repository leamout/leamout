package didww

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOrderNumberDefaultsToNoAutomaticRenewal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/orders" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		data := payload["data"].(map[string]any)
		attributes := data["attributes"].(map[string]any)
		items := attributes["items"].([]any)
		item := items[0].(map[string]any)["attributes"].(map[string]any)
		if got, ok := item["billing_cycles_count"].(float64); !ok || got != 0 {
			t.Fatalf("billing_cycles_count = %#v, want 0", item["billing_cycles_count"])
		}
		w.Header().Set("Content-Type", jsonAPIMediaType)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"order-1","type":"orders","attributes":{"reference":"ABC-123","amount":"10.0","status":"pending","description":"DID","created_at":"2026-09-12T00:00:00Z"}}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.OrderNumber(context.Background(), OrderNumberRequest{AvailableDIDID: "available-1", SKUID: "sku-1"}); err != nil {
		t.Fatal(err)
	}
}
