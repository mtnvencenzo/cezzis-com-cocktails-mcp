// Package server provides HTTP server implementations for the Cezzi Cocktails MCP server.
// It includes the core HTTP server functionality with support for health checks,
// version information, and MCP protocol streaming.
//
// The server package implements standard http.Handler interfaces and provides
// integration with the MCP protocol server. It supports both health monitoring
// and version reporting endpoints.
//
// Key features:
//   - HTTP server implementation for MCP protocol
//   - Health check endpoint for monitoring
//   - Version information endpoint
//   - Structured logging integration
//   - Request middleware support
package server

import (
	"net/http"

	"github.com/mark3labs/mcp-go/server"

	l "cezzis.com/cezzis-mcp-server/internal/logging"
	"cezzis.com/cezzis-mcp-server/internal/middleware"
)

// MCPHTTPServer wraps the MCP server HTTP functionality.
// It provides HTTP endpoints for health monitoring, version information,
// and MCP protocol communication.
type MCPHTTPServer struct {
	addr      string
	mcpServer *server.MCPServer
	version   string
}

// NewMCPHTTPServer creates a new MCPHTTPServer instance.
// It requires:
//   - addr: The address to listen on (e.g., ":8080")
//   - mcpServer: An initialized MCP protocol server
//   - version: The server version string
//
// Returns an MCPHTTPServer instance configured with the provided parameters.
func NewMCPHTTPServer(addr string, mcpServer *server.MCPServer, version string) *MCPHTTPServer {
	return &MCPHTTPServer{
		addr:      addr,
		mcpServer: mcpServer,
		version:   version,
	}
}

// Start initializes and runs the HTTP server.
// It sets up the following endpoints:
//   - /healthz: Health check endpoint
//   - /version: Server version information
//   - /mcp: MCP protocol endpoint with request logging
//
// Returns an error if the server fails to start or encounters an error while running.
func (s *MCPHTTPServer) Start() error {
	// Register health and version endpoints
	http.HandleFunc("/healthz", s.healthCheckHandler())
	http.HandleFunc("/version", s.versionHandler())

	// Use the official streamable HTTP server for MCP
	streamableHTTP := server.NewStreamableHTTPServer(s.mcpServer)

	// Wrap the MCP route to support GET probes and POST for MCP
	mcpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"ok", "sse":false}`))
			return
		case http.MethodPost:
			streamableHTTP.ServeHTTP(w, r)
			return
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
	})

	http.Handle("/mcp", middleware.RequestLogger(mcpHandler))

	l.Logger.Info().
		Str("port", s.addr).
		Msgf("Starting MCP Server on port '%s'", s.addr)

	return http.ListenAndServe(s.addr, nil)
}

func (s *MCPHTTPServer) healthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}
}

func (s *MCPHTTPServer) versionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"version": "` + s.version + `"}`))
	}
}
