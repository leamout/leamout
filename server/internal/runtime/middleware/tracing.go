package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/leamout/leamout/internal/platform/tracing"
)

// Tracing instruments inbound HTTP requests with OpenTelemetry server spans.
// Trace context is extracted from incoming requests and propagated through the
// request context for downstream application and integration instrumentation.
func Tracing(tracer *tracing.Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(
			next,
			"HTTP "+tracingOperation(next),
			otelhttp.WithTracer(tracer.Tracer()),
		)
	}
}

func tracingOperation(http.Handler) string {
	return "request"
}
