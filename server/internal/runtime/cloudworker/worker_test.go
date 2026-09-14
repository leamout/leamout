package worker

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/leamout/leamout/internal/platform/logging"
)

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
