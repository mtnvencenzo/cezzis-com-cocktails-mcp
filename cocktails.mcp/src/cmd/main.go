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
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/mark3labs/mcp-go/server"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/environment"
	"cezzis.com/cezzis-mcp-server/internal/mcpserver"
	"cezzis.com/cezzis-mcp-server/internal/repos"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

// Version uses build linkers to set this value at build time
var Version string = "0.0.0"

// main initializes and runs the Cezzi Cocktails MCP server, registering cocktail search and retrieval
// tools and serving requests over standard input/output or HTTP.
func main() {
	environment.LoadEnv()

	// Initialize OpenTelemetry SDK
	otelShutdown, err := telemetry.SetupOTelSDK(context.Background(), Version)
	if err != nil {
		telemetry.Logger.Error().Err(err).Msg("Failed to initialize logger")
	}

	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	// Initialize open telelmetry
	err = telemetry.InitTelemetry()
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}

	settings := config.GetAppSettings()
	assertAppSettings(settings)

	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Initialize authentication manager
	authManager := auth.NewOAuthFlowManager()
	cocktailsClient, err := cocktailsapi.GetClient()
	if err != nil {
		panic(err)
	}

	// Add the various tools to the MCP server
	// Each tool is registered with its corresponding handler function.
	// This allows clients to invoke the tools via the MCP protocol.

	// Basic cocktail tools (no authentication required)
	mcpServer.AddTool(tools.CocktailGetTool, server.ToolHandlerFunc(tools.NewCocktailGetToolHandler(cocktailsClient).Handle))
	mcpServer.AddTool(tools.CocktailSearchTool, server.ToolHandlerFunc(tools.NewCocktailSearchToolHandler(cocktailsClient).Handle))

	// Authentication tools
	mcpServer.AddTool(tools.AuthLoginTool, server.ToolHandlerFunc(tools.NewAuthLoginToolHandler(authManager).Handle))
	mcpServer.AddTool(tools.AuthStatusTool, server.ToolHandlerFunc(tools.NewAuthStatusToolHandler(authManager).Handle))
	mcpServer.AddTool(tools.AuthLogoutTool, server.ToolHandlerFunc(tools.NewAuthLogoutToolHandler(authManager).Handle))

	// Account Authenticated tools (require user login)
	mcpServer.AddTool(tools.RateCocktailTool, server.ToolHandlerFunc(tools.NewRateCocktailToolHandler(authManager, cocktailsClient).Handle))

	// Initialize the Cosmos DB (if not already initialized)
	err = repos.InitializeDatabase(context.Background())
	if err != nil {
		telemetry.Logger.Err(err).Msg("Failed to initialize database")
	} else {
		telemetry.Logger.Info().Msg("Database initialized")
	}

	// Finally, start the server in the chosen mode
	// Proper error handling ensures that any issues during startup are logged.
	// The server will run until it is manually stopped or encounters a fatal error.
	httpServer := mcpserver.NewMCPHTTPServer(
		fmt.Sprintf(":%d", settings.Port),
		mcpServer,
		Version,
		settings.TLSCertFile,
		settings.TLSKeyFile,
	)

	telemetry.Logger.Info().
		Str("version", Version).
		Str("port", strconv.Itoa(settings.Port)).
		Msg("Starting Cezzi Cocktails MCP Server")

	if err := httpServer.Start(); err != nil {
		telemetry.Logger.Fatal().Err(err).Msg("MCP HTTP server failed")
	}

	telemetry.Logger.Info().Msg("MCP HTTP server stopped")
}

func assertAppSettings(settings *config.AppSettings) {
	if settings.Port == 0 {
		telemetry.Logger.Warn().Msg("Warning: PORT is not set")
	}

	if settings.CocktailsAPIHost == "" {
		telemetry.Logger.Warn().Msg("Warning: COCKTAILS_API_HOST is not set")
	}

	if settings.CocktailsAPISubscriptionKey == "" {
		telemetry.Logger.Warn().Msg("Warning: COCKTAILS_API_XKEY is not set")
	}

	assertAuth0Settings(settings)
	assertCosmosSettings(settings)
	assertOtlpSettings(settings)
	assertTLSSettings(settings)
}

func assertAuth0Settings(settings *config.AppSettings) {
	if settings.Auth0Domain == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_DOMAIN is not set; authentication will fail")
	}

	if settings.Auth0ClientID == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_CLIENT_ID is not set; authentication will fail")
	}

	if settings.Auth0Audience == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_AUDIENCE is not set; authentication will fail")
	}

	if settings.Auth0Scopes == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_SCOPES is not set; authentication will fail")
	}
}

func assertCosmosSettings(settings *config.AppSettings) {
	if settings.CosmosConnectionString == "" {
		telemetry.Logger.Warn().Msg("Warning: COSMOS_CONNECTION_STRING is not set; database access will fail")
	}

	if settings.CosmosAccountEndpoint == "" {
		telemetry.Logger.Warn().Msg("Warning: COSMOS_ACCOUNT_ENDPOINT is not set; database access will fail")
	}

	if settings.CosmosDatabaseName == "" {
		telemetry.Logger.Warn().Msg("Warning: COSMOS_DATABASE_NAME is not set; database access will fail")
	}

	if settings.CosmosContainerName == "" {
		telemetry.Logger.Warn().Msg("Warning: COSMOS_CONTAINER_NAME is not set; database access will fail")
	}
}

func assertOtlpSettings(settings *config.AppSettings) {
	if settings.OTLPEndpoint == "" {
		telemetry.Logger.Warn().Msg("Warning: OTLP_ENDPOINT is not set; telemetry will not be exported")
	}

	if settings.OTLPInsecure {
		telemetry.Logger.Info().Msg("OTLP Insecure mode is enabled")
	}

	if settings.OTLPTraceEnabled {
		telemetry.Logger.Info().Msg("OTLP Trace Exporter is enabled")
	} else {
		telemetry.Logger.Info().Msg("OTLP Trace Exporter is disabled")
	}

	if settings.OTLPMetricsEnabled {
		telemetry.Logger.Info().Msg("OTLP Metrics Exporter is enabled")
	} else {
		telemetry.Logger.Info().Msg("OTLP Metrics Exporter is disabled")
	}

	if settings.OTLPLogEnabled {
		telemetry.Logger.Info().Msg("OTLP Log Exporter is enabled")
	} else {
		telemetry.Logger.Info().Msg("OTLP Log Exporter is disabled")
	}
}

func assertTLSSettings(settings *config.AppSettings) {
	if settings.TLSCertFile != "" && settings.TLSKeyFile != "" {
		telemetry.Logger.Info().
			Str("cert", settings.TLSCertFile).
			Str("key", settings.TLSKeyFile).
			Msg("TLS/HTTPS is enabled")
	} else if settings.TLSCertFile != "" || settings.TLSKeyFile != "" {
		telemetry.Logger.Warn().Msg("Warning: Both TLS_CERT_FILE and TLS_KEY_FILE must be set for HTTPS; falling back to HTTP")
	} else {
		telemetry.Logger.Info().Msg("TLS/HTTPS is not configured; using HTTP")
	}
}
