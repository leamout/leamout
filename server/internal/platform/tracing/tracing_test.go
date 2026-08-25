package tracing

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestNewUsesDefaultInstrumentationName(t *testing.T) {
	tracer := New("")
	if tracer == nil {
		t.Fatal("expected tracer")
	}

	if tracer.Tracer() == nil {
		t.Fatal("expected OpenTelemetry tracer")
	}
}

func TestStart(t *testing.T) {
	tracer := New("leamout.test")

	ctx, span := tracer.Start(context.Background(), "test")
	if ctx == nil {
		t.Fatal("expected context")
	}
	if span == nil {
		t.Fatal("expected span")
	}

	span.End()
}

var _ trace.Tracer = (*Tracer)(nil)
