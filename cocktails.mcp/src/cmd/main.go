// Package main is the entry point for the Cezzi Cocktails MCP server.
// It initializes and runs the MCP server, registering tools and serving requests
// over standard input/output or HTTP.
//
// The server supports two modes:
//   - Standard input/output (stdio) - default mode
//   - HTTP mode - if --http flag is provided
//
// The server includes:
//   - MCP server initialization with tool registration
//   - HTTP server setup with health check endpoint
//   - Logging middleware for request tracking
//   - Proper error handling and response formatting
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/xid"
	"github.com/rs/zerolog"

	l "cezzis.com/cezzis-mcp-server/pkg/logging"
	"cezzis.com/cezzis-mcp-server/pkg/tools"
)

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

type ctxKey string

// main initializes and runs the Cezzi Cocktails MCP server, registering cocktail search and retrieval tools and serving requests over standard input/output or HTTP.
func main() {
	_, err := l.InitLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Add a flag to choose between stdio and HTTP
	httpAddr := flag.String("http", "", "If set, serve HTTP on this address (e.g., :8080). Otherwise, use stdio.")
	flag.Parse()

	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	mcpServer.AddTool(tools.CocktailSearchTool, server.ToolHandlerFunc(tools.CocktailSearchToolHandler))
	mcpServer.AddTool(tools.CocktailGetTool, server.ToolHandlerFunc(tools.CocktailGetToolHandler))

	requestLogger := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lg := l.Logger

			correlationID := xid.New().String()

			ctx := context.WithValue(r.Context(), ctxKey("correlation_id"), correlationID)

			r = r.WithContext(ctx)

			lg.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("correlation_id", correlationID)
			})

			w.Header().Add("X-Correlation-ID", correlationID)

			lrw := newLoggingResponseWriter(w)

			r = r.WithContext(lg.WithContext(r.Context()))

			defer func() {
				panicVal := recover()
				if panicVal != nil {
					lrw.statusCode = http.StatusInternalServerError
					panic(panicVal)
				}

				lg.
					Info().
					Str("method", r.Method).
					Str("url", r.URL.RequestURI()).
					Int("status_code", lrw.statusCode).
					Dur("elapsed_ms", time.Since(start)).
					Msgf("MCP: %s %s %d %s", r.Method, r.URL.RequestURI(), lrw.statusCode, time.Since(start))
			}()

			next.ServeHTTP(lrw, r)
		})
	}

	if *httpAddr != "" {
		// HTTP mode
		http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
				l.Logger.Err(err).Msg(fmt.Sprintf("Error writing health check response: %v", err))
			}
		})

		// Use the official streamable HTTP server for MCP
		streamableHTTP := server.NewStreamableHTTPServer(mcpServer)
		http.Handle("/mcp", requestLogger(streamableHTTP))

		l.Logger.Info().
			Str("port", *httpAddr).
			Msgf("Starting MCP Server on port '%s'", *httpAddr)

		l.Logger.Fatal().
			Err(http.ListenAndServe(*httpAddr, nil)).
			Msg("MCP Server Closed")

	} else {
		// Stdio mode (default)
		if err := server.ServeStdio(mcpServer); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}
}
