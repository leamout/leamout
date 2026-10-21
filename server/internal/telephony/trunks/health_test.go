package trunks

import (
	"bufio"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

func TestReadSIPStatusAcceptsAuthenticationChallenge(t *testing.T) {
	code, err := readSIPStatus(bufio.NewReader(strings.NewReader(
		"SIP/2.0 401 Unauthorized\r\nContent-Length: 0\r\n\r\n",
	)))
	if err != nil {
		t.Fatalf("readSIPStatus() error = %v", err)
	}
	if code != 401 {
		t.Fatalf("readSIPStatus() code = %d, want 401", code)
	}
}

// The worker must reject private targets before opening TCP, TLS or UDP sockets,
// including endpoints that were stored before public-host validation existed.
func TestSIPOptionsProberBlocksPrivateTargets(t *testing.T) {
	for _, host := range []string{"127.0.0.1", "::1", "10.0.0.5", "100.64.1.2", "169.254.169.254", "localhost"} {
		for _, transport := range []string{"udp", "tcp", "tls"} {
			t.Run(host+"/"+transport, func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_, err := (SIPOptionsProber{}).Probe(ctx, sqlc.TrunkEndpoint{
					Host: host, Port: 5060, Transport: transport,
				})
				if err == nil || !strings.Contains(err.Error(), "non-public address") {
					t.Fatalf("Probe() error = %v, want non-public address rejection", err)
				}
			})
		}
	}
}

type fakeHealthRepository struct {
	endpoints       []sqlc.TrunkEndpoint
	healthy, failed []uuid.UUID
}

func (f *fakeHealthRepository) ListForHealthCheck(context.Context, time.Time, time.Time, int32) ([]sqlc.TrunkEndpoint, error) {
	return f.endpoints, nil
}
func (f *fakeHealthRepository) MarkHealthy(_ context.Context, id uuid.UUID, _ time.Time, _, _ int32) error {
	f.healthy = append(f.healthy, id)
	return nil
}
func (f *fakeHealthRepository) MarkProbeFailed(_ context.Context, id uuid.UUID, _ time.Time, _ int32, _ string, _ int32, _ time.Time) error {
	f.failed = append(f.failed, id)
	return nil
}

type fakeProber struct {
	mu       sync.Mutex
	failures map[uuid.UUID]bool
}

func (f *fakeProber) Probe(_ context.Context, endpoint sqlc.TrunkEndpoint) (int32, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failures[endpoint.ID] {
		return 0, context.DeadlineExceeded
	}
	return 200, nil
}

func TestHealthCheckJobPersistsProbeResults(t *testing.T) {
	healthyID, failedID := uuid.New(), uuid.New()
	repo := &fakeHealthRepository{endpoints: []sqlc.TrunkEndpoint{{ID: healthyID}, {ID: failedID}}}
	job, err := NewHealthCheckJob(repo, &fakeProber{failures: map[uuid.UUID]bool{failedID: true}}, DefaultHealthCheckConfig())
	if err != nil {
		t.Fatal(err)
	}
	if err := job.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if len(repo.healthy) != 1 || repo.healthy[0] != healthyID {
		t.Fatalf("healthy = %v", repo.healthy)
	}
	if len(repo.failed) != 1 || repo.failed[0] != failedID {
		t.Fatalf("failed = %v", repo.failed)
	}
}
