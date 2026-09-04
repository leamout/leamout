package didww

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchNumbersUsesDIDWWHeadersAndFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v3/available_dids" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Api-Key"); got != "secret" {
			t.Fatalf("Api-Key = %q", got)
		}
		if got := r.Header.Get("X-DIDWW-API-Version"); got != DefaultAPIVersion {
			t.Fatalf("X-DIDWW-API-Version = %q", got)
		}
		if got := r.URL.Query().Get("filter[number_contains]"); got != "212" {
			t.Fatalf("number_contains = %q", got)
		}
		if got := r.URL.Query().Get("filter[did_group.features]"); got != "voice_in" {
			t.Fatalf("feature = %q", got)
		}
		w.Header().Set("Content-Type", jsonAPIMediaType)
		_, _ = w.Write([]byte(`{"data":[{"id":"available-1","type":"available_dids","attributes":{"number":"12124727600"}}]}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	numbers, err := client.SearchNumbers(context.Background(), SearchNumbersRequest{
		NumberContains: "212",
		Feature:        "voice_in",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(numbers) != 1 || numbers[0].ID != "available-1" || numbers[0].Number != "12124727600" {
		t.Fatalf("numbers = %+v", numbers)
	}
}

func TestOrderNumberPurchasesSpecificAvailableDID(t *testing.T) {
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
		if attributes["allow_back_ordering"] != false {
			t.Fatalf("allow_back_ordering = %v", attributes["allow_back_ordering"])
		}
		items := attributes["items"].([]any)
		item := items[0].(map[string]any)["attributes"].(map[string]any)
		if item["available_did_id"] != "available-1" || item["sku_id"] != "sku-1" {
			t.Fatalf("order item = %+v", item)
		}
		w.Header().Set("Content-Type", jsonAPIMediaType)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"id":"order-1","type":"orders","attributes":{"reference":"ABC-123","amount":"10.0","status":"pending","description":"DID","created_at":"2026-09-04T00:00:00Z"}}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	order, err := client.OrderNumber(context.Background(), OrderNumberRequest{
		AvailableDIDID: "available-1",
		SKUID:          "sku-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if order.ID != "order-1" || order.Reference != "ABC-123" || order.Status != "pending" {
		t.Fatalf("order = %+v", order)
	}
}

func TestConfigureRoutingAssignsVoiceInTrunk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v3/dids/did-1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		body := new(strings.Builder)
		_, _ = body.ReadFrom(r.Body)
		if !strings.Contains(body.String(), `"voice_in_trunk":{"data":{"id":"trunk-1","type":"voice_in_trunks"}}`) {
			t.Fatalf("body = %s", body.String())
		}
		w.Header().Set("Content-Type", jsonAPIMediaType)
		_, _ = w.Write([]byte(`{"data":{"id":"did-1","type":"dids","attributes":{"number":"12124727600"},"relationships":{"voice_in_trunk":{"data":{"id":"trunk-1","type":"voice_in_trunks"}}}}}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	did, err := client.ConfigureRouting(context.Background(), "did-1", "trunk-1")
	if err != nil {
		t.Fatal(err)
	}
	if did.ID != "did-1" || did.VoiceInTrunkID != "trunk-1" {
		t.Fatalf("DID = %+v", did)
	}
}
