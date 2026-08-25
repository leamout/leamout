package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Tracer provides Leamout's application-facing tracing API.
type Tracer struct {
	tracer trace.Tracer
}

// New creates a tracer for the supplied instrumentation name.
func New(name string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(name),
	}
}

// Start starts a new span and returns the derived context and span.
func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// Tracer returns the underlying OpenTelemetry tracer.
func (t *Tracer) Tracer() trace.Tracer {
	return t.tracer
}
