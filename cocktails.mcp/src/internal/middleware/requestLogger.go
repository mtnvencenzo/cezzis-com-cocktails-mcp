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
	"time"

	"github.com/rs/xid"

	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

// RequestLogger is a middleware that logs incoming HTTP requests and their outcomes.
// It implements the following features:
//   - Generates a unique correlation ID for each request
//   - Adds the correlation ID to the request context
//   - Sets X-Correlation-ID header in the response
//   - Logs request method, URL, status code, and timing
//   - Recovers from panics and sets 500 status code
//   - Uses structured logging via zerolog
//
// Usage:
//
//	http.Handle("/api", middleware.RequestLogger(apiHandler))
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lg := l.Logger

		correlationID := xid.New().String()

		ctx := context.WithValue(r.Context(), ctxKey("correlation_id"), correlationID)

		r = r.WithContext(ctx)

		w.Header().Add("X-Correlation-ID", correlationID)

		lrw := newLoggingResponseWriter(w)

		reqLogger := lg.With().Str("correlation_id", correlationID).Logger()
		r = r.WithContext(reqLogger.WithContext(r.Context()))

		defer func() {
			if rec := recover(); rec != nil {
				lrw.statusCode = http.StatusInternalServerError
				reqLogger.Error().
					Str("method", r.Method).
					Str("url", r.URL.RequestURI()).
					Int("status_code", lrw.statusCode).
					Int64("elapsed_ms", time.Since(start).Milliseconds()).
					Interface("panic", rec).
					Msg("Recovered panic in HTTP handler")
				http.Error(lrw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			reqLogger.
				Info().
				Str("method", r.Method).
				Str("url", r.URL.RequestURI()).
				Int("status_code", lrw.statusCode).
				Int64("elapsed_ms", time.Since(start).Milliseconds()).
				Msgf("MCP: %s %s %d", r.Method, r.URL.RequestURI(), lrw.statusCode)
		}()

		next.ServeHTTP(lrw, r)
	})
}

type ctxKey string

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
