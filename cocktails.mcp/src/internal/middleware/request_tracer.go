// Package middleware provides HTTP middleware components for the Cezzi Cocktails MCP server.
// It includes middleware for request logging, correlation ID tracking, and response monitoring.
//
// The middleware package is designed to be used with standard http.Handler interfaces
// and provides instrumentation for HTTP request/response cycles. It implements common
// patterns like correlation ID propagation and structured logging using zerolog.
//
// Key features:
//   - Request logging with timing information
//   - Correlation ID generation and propagation
//   - Response status code tracking
//   - Panic recovery with proper status code setting
package middleware

import (
	"context"
	"net/http"

	"github.com/rs/xid"

	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// RequestTracer is a middleware that traces incoming HTTP requests using OpenTelemetry.
// It implements the following features:
//   - Starts a new trace span for each incoming request
//   - Adds a unique correlation ID to the request context
//   - Sets X-Correlation-ID header in the response
//   - Uses structured logging via zerolog
//
// Usage:
//
//	http.Handle("/api", middleware.RequestTracer(apiHandler))
func RequestTracer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, span := telemetry.Tracer.Start(req.Context(), req.Method+" "+req.URL.Path)
		defer span.End()

		lrw := newTracingResponseWriter(w)

		correlationID := xid.New().String()
		ctx = context.WithValue(ctx, CorrelationIDCtxKey("correlation_id"), correlationID)
		req = req.WithContext(ctx)

		w.Header().Add("X-Correlation-ID", correlationID)

		next.ServeHTTP(lrw, req)
	})
}

// CorrelationIDCtxKey is the context key type for correlation IDs.
type CorrelationIDCtxKey string

type tracingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	wrote      bool
}

func newTracingResponseWriter(w http.ResponseWriter) *tracingResponseWriter {
	return &tracingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (lrw *tracingResponseWriter) WriteHeader(code int) {
	if lrw.wrote {
		// update observed status but don't write twice
		lrw.statusCode = code
		return
	}
	lrw.statusCode = code
	lrw.wrote = true
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *tracingResponseWriter) Write(b []byte) (int, error) {
	if !lrw.wrote {
		lrw.WriteHeader(http.StatusOK)
	}
	return lrw.ResponseWriter.Write(b)
}
