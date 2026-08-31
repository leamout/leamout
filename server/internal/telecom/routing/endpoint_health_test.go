package routing

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type fakeHealthStore struct {
	endpoints []sqlc.TrunkEndpoint
	list      sqlc.ListTrunkEndpointsForHealthCheckParams
	healthy   []sqlc.MarkTrunkEndpointHealthyParams
	failed    []sqlc.MarkTrunkEndpointProbeFailedParams
}

func (f *fakeHealthStore) ListTrunkEndpointsForHealthCheck(_ context.Context, arg sqlc.ListTrunkEndpointsForHealthCheckParams) ([]sqlc.TrunkEndpoint, error) {
	f.list = arg
	return f.endpoints, nil
}
func (f *fakeHealthStore) MarkTrunkEndpointHealthy(_ context.Context, arg sqlc.MarkTrunkEndpointHealthyParams) (sqlc.TrunkEndpoint, error) {
	f.healthy = append(f.healthy, arg)
	return sqlc.TrunkEndpoint{}, nil
}
func (f *fakeHealthStore) MarkTrunkEndpointProbeFailed(_ context.Context, arg sqlc.MarkTrunkEndpointProbeFailedParams) (sqlc.TrunkEndpoint, error) {
	f.failed = append(f.failed, arg)
	return sqlc.TrunkEndpoint{}, nil
}

type fakeEndpointProber struct {
	results map[uuid.UUID]ProbeResult
	errors  map[uuid.UUID]error
}

func (f fakeEndpointProber) Probe(_ context.Context, endpoint sqlc.TrunkEndpoint) (ProbeResult, error) {
	return f.results[endpoint.ID], f.errors[endpoint.ID]
}

func TestEndpointHealthJobPersistsSuccessAndCircuitBreakerFailure(t *testing.T) {
	healthyID, failedID := uuid.New(), uuid.New()
	store := &fakeHealthStore{endpoints: []sqlc.TrunkEndpoint{{ID: healthyID}, {ID: failedID}}}
	job, err := NewEndpointHealthJob(store, fakeEndpointProber{
		results: map[uuid.UUID]ProbeResult{
			healthyID: {ResponseCode: 200, Latency: 12 * time.Millisecond},
			failedID:  {Latency: 2 * time.Second},
		},
		errors: map[uuid.UUID]error{failedID: errors.New("probe timeout")},
	})
	if err != nil {
		t.Fatalf("new job: %v", err)
	}
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	job.now = func() time.Time { return now }

	if err := job.Check(context.Background()); err != nil {
		t.Fatalf("check endpoints: %v", err)
	}
	if len(store.healthy) != 1 || store.healthy[0].ID != healthyID || *store.healthy[0].ResponseCode != 200 || *store.healthy[0].LatencyMs != 12 {
		t.Fatalf("healthy update = %+v", store.healthy)
	}
	if len(store.failed) != 1 || store.failed[0].ID != failedID || store.failed[0].FailureThreshold != 3 {
		t.Fatalf("failure update = %+v", store.failed)
	}
	if !store.failed[0].CooldownUntil.Time.Equal(now.Add(30*time.Second)) || *store.failed[0].LastError != "probe timeout" {
		t.Fatalf("failure cooldown = %+v", store.failed[0])
	}
	if !store.list.DueBefore.Time.Equal(now.Add(-10*time.Second)) || store.list.BatchSize != 100 {
		t.Fatalf("health query = %+v", store.list)
	}
}

func TestSIPOptionsProberAcceptsReachableSIPResponse(t *testing.T) {
	listener, err := (&net.ListenConfig{}).ListenPacket(context.Background(), "udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen UDP: %v", err)
	}
	t.Cleanup(func() {
		if err := listener.Close(); err != nil {
			t.Errorf("close UDP listener: %v", err)
		}
	})

	received := make(chan string, 1)
	go func() {
		buffer := make([]byte, 4096)
		count, remote, readErr := listener.ReadFrom(buffer)
		if readErr != nil {
			return
		}
		received <- string(buffer[:count])
		_, _ = listener.WriteTo([]byte("SIP/2.0 200 OK\r\nContent-Length: 0\r\n\r\n"), remote)
	}()

	address := listener.LocalAddr().(*net.UDPAddr)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result, err := NewSIPOptionsProber().Probe(ctx, sqlc.TrunkEndpoint{
		Host: "127.0.0.1", Port: int32(address.Port), Transport: "udp",
	})
	if err != nil {
		t.Fatalf("probe endpoint: %v", err)
	}
	if result.ResponseCode != 200 {
		t.Fatalf("response code = %d", result.ResponseCode)
	}
	request := <-received
	if !strings.HasPrefix(request, "OPTIONS sip:") || !strings.Contains(request, "SIP/2.0/UDP") {
		t.Fatalf("unexpected OPTIONS request: %q", request)
	}
}
