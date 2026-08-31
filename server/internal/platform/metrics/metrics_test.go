package metrics

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type fakeSharedStore struct {
	counters map[string]string
	gauges   map[string]string
	fields   []string
	limits   []int64
}

func (f *fakeSharedStore) IncrementMetric(_ context.Context, field string, limit int64) error {
	f.fields = append(f.fields, field)
	f.limits = append(f.limits, limit)
	return nil
}
func (f *fakeSharedStore) SetMetricGauge(context.Context, string, float64, int64) error { return nil }
func (f *fakeSharedStore) TelecomMetrics(context.Context) (map[string]string, map[string]string, error) {
	return f.counters, f.gauges, nil
}

func TestRegistryUsesBoundedResourceLabels(t *testing.T) {
	store := &fakeSharedStore{}
	registry := New(store)
	carrier, trunk, endpoint := uuid.New(), uuid.New(), uuid.New()
	registry.EndpointSelection(context.Background(), carrier, trunk, endpoint, true)
	if len(store.fields) != 1 || store.fields[0] != series("endpoint_selections_total", carrier, trunk, endpoint, "failover") {
		t.Fatalf("fields = %v", store.fields)
	}
	if store.limits[0] != maxTelecomSeries {
		t.Fatalf("series limit = %d", store.limits[0])
	}
}

func TestHandlerPublishesPrometheusTelecomSeries(t *testing.T) {
	carrier := uuid.New()
	field := series("calls_attempted_total", carrier, uuid.Nil, uuid.Nil, "")
	registry := New(&fakeSharedStore{counters: map[string]string{field: "7"}, gauges: map[string]string{}})
	recorder := httptest.NewRecorder()
	Handler(registry).ServeHTTP(recorder, httptest.NewRequest("GET", "/metrics", nil))
	if recorder.Code != 200 {
		t.Fatalf("status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `leamout_calls_attempted_total{carrier_connection_id="`+carrier.String()+`"} 7`) {
		t.Fatalf("metric output missing attributed counter:\n%s", body)
	}
}
