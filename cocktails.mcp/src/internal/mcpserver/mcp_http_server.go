// Package mcpserver provides HTTP server implementations for the Cezzi Cocktails MCP server.
// It includes the core HTTP server functionality with support for health checks,
// version information, and MCP protocol streaming.
//
// The mcpserver package implements standard http.Handler interfaces and provides
// integration with the MCP protocol server. It supports both health monitoring
// and version reporting endpoints.
//
// Key features:
//   - HTTP server implementation for MCP protocol
//   - Health check endpoint for monitoring
//   - Version information endpoint
//   - Structured logging integration
//   - Request middleware support
package mcpserver

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mark3labs/mcp-go/server"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/db"
	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// MCPHTTPServer wraps the MCP server HTTP functionality.
// It provides HTTP endpoints for health monitoring, version information,
// and MCP protocol communication.
type MCPHTTPServer struct {
	addr        string
	mcpServer   *server.MCPServer
	version     string
	tlsCertFile string
	tlsKeyFile  string
	pool        *pgxpool.Pool
	settings    *config.AppSettings
}

// NewMCPHTTPServer creates a new MCPHTTPServer instance.
// It requires:
//   - addr: The address to listen on (e.g., ":8080")
//   - mcpServer: An initialized MCP protocol server
//   - version: The server version string
//   - tlsCertFile: Path to TLS certificate file (optional, empty string for HTTP)
//   - tlsKeyFile: Path to TLS private key file (optional, empty string for HTTP)
//
// Returns an MCPHTTPServer instance configured with the provided parameters.
func NewMCPHTTPServer(addr string, mcpServer *server.MCPServer, version string, tlsCertFile string, tlsKeyFile string, pool *pgxpool.Pool, settings *config.AppSettings) *MCPHTTPServer {
	return &MCPHTTPServer{
		addr:        addr,
		mcpServer:   mcpServer,
		version:     version,
		tlsCertFile: tlsCertFile,
		tlsKeyFile:  tlsKeyFile,
		pool:        pool,
		settings:    settings,
	}
}

// Start initializes and runs the HTTP server.
// It sets up the following endpoints:
//   - /healthz: Health check endpoint
//   - /version: Server version information
//   - /mcp: MCP protocol endpoint with request logging
//
// If TLS certificate and key files are configured, the server will start with HTTPS.
// Otherwise, it will start with HTTP.
//
// Returns an error if the server fails to start or encounters an error while running.
func (s *MCPHTTPServer) Start() error {
	// Register health and version endpoints
	http.HandleFunc("/healthz", s.healthCheckHandler())
	http.HandleFunc("/version", s.versionHandler())

	// Register Dapr job endpoint for application initialization
	http.HandleFunc("/job/initialize-app", s.initializeAppHandler())

	// Use the official streamable HTTP server for MCP
	streamableHTTP := server.NewStreamableHTTPServer(s.mcpServer)

	// Wrap the MCP route to support GET probes and POST for MCP
	mcpMiddleware := middleware.McpRequestHandler(nil, streamableHTTP)

	loggingMiddleware := middleware.RequestLogger(mcpMiddleware)

	tracingMiddleware := middleware.RequestTracer(loggingMiddleware)

	opts := []otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return "MCP " + r.Method
		}),
	}

	http.Handle("/mcp", otelhttp.NewHandler(tracingMiddleware, "mcp", opts...))

	// Check if TLS is configured
	if s.tlsCertFile != "" && s.tlsKeyFile != "" {
		telemetry.Logger.Info().
			Str("port", s.addr).
			Str("protocol", "https").
			Msgf("Starting MCP Server with HTTPS on port '%s'", s.addr)

		return http.ListenAndServeTLS(s.addr, s.tlsCertFile, s.tlsKeyFile, nil)
	}

	telemetry.Logger.Info().
		Str("port", s.addr).
		Str("protocol", "http").
		Msgf("Starting MCP Server with HTTP on port '%s'", s.addr)

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

func (s *MCPHTTPServer) initializeAppHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Validate Dapr app token if configured
		if s.settings.AppAPIToken != "" {
			token := r.Header.Get("dapr-api-token")
			if token != s.settings.AppAPIToken {
				telemetry.Logger.Warn().Msg("Unauthorized init job request: invalid dapr-api-token")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		telemetry.Logger.Info().Msg("Received initialize-app job request")

		ctx := r.Context()

		// Ensure database exists
		if err := db.EnsureDatabaseExists(ctx, s.settings); err != nil {
			telemetry.Logger.Error().Err(err).Msg("Failed to ensure database exists")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			resp, _ := json.Marshal(map[string]string{"error": "failed to ensure database exists"})
			_, _ = w.Write(resp)
			return
		}

		// Ensure tables exist
		if err := db.EnsureTablesExist(ctx, s.pool); err != nil {
			telemetry.Logger.Error().Err(err).Msg("Failed to ensure tables exist")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			resp, _ := json.Marshal(map[string]string{"error": "failed to ensure tables exist"})
			_, _ = w.Write(resp)
			return
		}

		telemetry.Logger.Info().Msg("Application initialization completed successfully")
		w.WriteHeader(http.StatusOK)
	}
}
