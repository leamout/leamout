package didww

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leamout/leamout/internal/telecom/numbers"
)

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
