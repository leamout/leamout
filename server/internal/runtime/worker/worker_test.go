package worker

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
)

func TestWorkerCompositionRootsRejectWrongModeBeforeConnecting(t *testing.T) {
	if _, err := NewCloud(t.Context(), config.Config{DeploymentMode: config.DeploymentModeSelfHosted}); err == nil {
		t.Fatal("NewCloud() accepted self-hosted mode")
	}
	if _, err := NewSelfHosted(t.Context(), config.Config{DeploymentMode: config.DeploymentModeCloud}); err == nil {
		t.Fatal("NewSelfHosted() accepted Cloud mode")
	}
}

func TestSelfHostedWorkerRejectsCloudCredentialsBeforeConnecting(t *testing.T) {
	cfg := config.Config{
		DeploymentMode: config.DeploymentModeSelfHosted,
		CommPeak:       config.CommPeakConfig{Authorization: "cloud-provider-secret"},
	}
	if _, err := NewSelfHosted(t.Context(), cfg); err == nil {
		t.Fatal("NewSelfHosted() accepted Cloud provider credentials")
	}
}

func TestRunComponentReportsUnexpectedStop(t *testing.T) {
	worker := &Worker{
		health: newHealthState("test-component"),
		logger: logging.NewWithHandler(slog.NewTextHandler(io.Discard, nil)),
	}
	errCh := make(chan error, 1)

	worker.runComponent(context.Background(), errCh, "test-component", func(context.Context) error { return nil })

	err := <-errCh
	if !strings.Contains(err.Error(), "component stopped unexpectedly") {
		t.Fatalf("expected unexpected stop error, got %v", err)
	}
	state := worker.health.snapshot()["test-component"]
	if state.Running || !strings.Contains(state.LastError, "component stopped unexpectedly") {
		t.Fatalf("unexpected component state: %+v", state)
	}
}

func TestRunComponentDoesNotReportContextCancellation(t *testing.T) {
	worker := &Worker{
		health: newHealthState("test-component"),
		logger: logging.NewWithHandler(slog.NewTextHandler(io.Discard, nil)),
	}
	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	worker.runComponent(ctx, errCh, "test-component", func(context.Context) error { return context.Canceled })

	select {
	case err := <-errCh:
		t.Fatalf("did not expect cancellation to be reported: %v", err)
	default:
	}
}
