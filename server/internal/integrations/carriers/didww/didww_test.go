package didww

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/leamout/leamout/internal/telecom/numbers"
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
	result, err := client.SearchNumbers(context.Background(), SearchNumbersRequest{
		NumberContains: "212",
		Feature:        "voice_in",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].ID != "available-1" || result[0].Number != "12124727600" {
		t.Fatalf("numbers = %+v", result)
	}
}

func TestSearchAvailableResolvesCountryAndVoiceSKU(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", jsonAPIMediaType)
		switch r.URL.Path {
		case "/v3/countries":
			if got := r.URL.Query().Get("filter[iso]"); got != "US" {
				t.Fatalf("country iso filter = %q", got)
			}
			_, _ = w.Write([]byte(`{"data":[{"id":"country-us","type":"countries","attributes":{"iso":"US"}}]}`))
		case "/v3/available_dids":
			if got := r.URL.Query().Get("filter[country.id]"); got != "country-us" {
				t.Fatalf("country id filter = %q", got)
			}
			if got := r.URL.Query().Get("filter[did_group.features]"); got != "voice_in" {
				t.Fatalf("feature filter = %q", got)
			}
			_, _ = w.Write([]byte(`{
				"data":[{
					"id":"available-1",
					"type":"available_dids",
					"attributes":{"number":"12125550100"},
					"relationships":{"did_group":{"data":{"type":"did_groups","id":"group-1"}}}
				}],
				"included":[
					{"id":"group-1","type":"did_groups","attributes":{},"relationships":{"stock_keeping_units":{"data":[{"type":"stock_keeping_units","id":"sku-zero"},{"type":"stock_keeping_units","id":"sku-two"},{"type":"stock_keeping_units","id":"sku-five"}]}}},
					{"id":"sku-zero","type":"stock_keeping_units","attributes":{"channels_included_count":0}},
					{"id":"sku-two","type":"stock_keeping_units","attributes":{"channels_included_count":2}},
					{"id":"sku-five","type":"stock_keeping_units","attributes":{"channels_included_count":5}}
				]
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	result, err := client.SearchAvailable(context.Background(), numbers.AvailableSearchRequest{CountryCode: "US", Contains: "212"})
	if err != nil {
		t.Fatalf("SearchAvailable() error = %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("result count = %d", len(result))
	}
	if result[0].Provider != "didww" || result[0].ProviderInventoryID != "available-1" || result[0].ProviderProductID != "sku-two" {
		t.Fatalf("candidate = %#v", result[0])
	}
	if result[0].Number != "+12125550100" || result[0].CountryCode != "US" || result[0].ChannelsIncludedCount != 2 {
		t.Fatalf("candidate = %#v", result[0])
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

func TestEnsureInboundTrunkCreatesDeploymentScopedSIPTarget(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", jsonAPIMediaType)
		switch requests {
		case 1:
			if r.Method != http.MethodGet || r.URL.Path != "/v3/voice_in_trunks" {
				t.Fatalf("request = %s %s", r.Method, r.URL.Path)
			}
			if got := r.URL.Query().Get("filter[external_reference_id]"); got != "leamout:deployment-1:managed-ingress" {
				t.Fatalf("external reference filter = %q", got)
			}
			_, _ = w.Write([]byte(`{"data":[]}`))
		case 2:
			if r.Method != http.MethodPost || r.URL.Path != "/v3/voice_in_trunks" {
				t.Fatalf("request = %s %s", r.Method, r.URL.Path)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			payload := string(body)
			for _, expected := range []string{
				`"external_reference_id":"leamout:deployment-1:managed-ingress"`,
				`"username":"+{DID}"`,
				`"host":"sip.leamout.example"`,
				`"port":5060`,
				`"transport_protocol_id":1`,
				`"resolve_ruri":true`,
				`"auth_enabled":false`,
			} {
				if !strings.Contains(payload, expected) {
					t.Fatalf("body missing %s: %s", expected, payload)
				}
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"voice-in-1","type":"voice_in_trunks","attributes":{"name":"DIDWW Managed Ingress","external_reference_id":"leamout:deployment-1:managed-ingress","configuration":{"type":"sip_configurations","attributes":{"username":"+{DID}","host":"sip.leamout.example","port":5060,"transport_protocol_id":1,"resolve_ruri":true,"auth_enabled":false}}}}}`))
		default:
			t.Fatalf("unexpected extra request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	trunk, err := client.EnsureInboundTrunk(context.Background(), EnsureInboundTrunkRequest{
		Name:                "DIDWW Managed Ingress",
		ExternalReferenceID: "leamout:deployment-1:managed-ingress",
		Host:                "sip.leamout.example",
		Port:                5060,
		Transport:           "udp",
	})
	if err != nil {
		t.Fatal(err)
	}
	if trunk.ID != "voice-in-1" || trunk.Host != "sip.leamout.example" || trunk.TransportProtocolID != 1 {
		t.Fatalf("trunk = %+v", trunk)
	}
}

func TestConfigureRoutingAssignsVoiceInTrunk(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v3/dids/did-1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), `"voice_in_trunk":{"data":{"id":"trunk-1","type":"voice_in_trunks"}}`) {
			t.Fatalf("body = %s", body)
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

func TestReleaseNumberDeletesOwnedDID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v3/dids/did-1" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client, err := NewClient(Config{BaseURL: server.URL, APIKey: "secret", HTTPClient: server.Client()})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.ReleaseNumber(context.Background(), "did-1"); err != nil {
		t.Fatal(err)
	}
}
