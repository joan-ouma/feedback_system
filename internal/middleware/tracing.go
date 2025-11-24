package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// TracingMiddleware wraps handlers with OpenTelemetry tracing
func TracingMiddleware(handler http.Handler) http.Handler {
	return otelhttp.NewHandler(handler, "http.server")
}

