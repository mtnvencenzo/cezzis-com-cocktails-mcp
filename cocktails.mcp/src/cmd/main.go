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

	"cezzis.com/cezzis-mcp-server/internal/api/accountsapi"
	"cezzis.com/cezzis-mcp-server/internal/api/aisearch"
	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/background"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/db"
	"cezzis.com/cezzis-mcp-server/internal/environment"
	"cezzis.com/cezzis-mcp-server/internal/mcpserver"
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

	// Initialize PostgreSQL connection pool
	pool, err := db.NewPool(context.Background(), settings)
	if err != nil {
		log.Fatalf("Failed to create PostgreSQL connection pool: %v", err)
	}
	defer pool.Close()

	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithLogging(),
	)

	// Initialize authentication manager
	authManager := auth.NewOAuthFlowManager(pool)
	cocktailsClient, err := cocktailsapi.GetClient()
	if err != nil {
		panic(err)
	}

	aiSearchClient, err := aisearch.GetClient()
	if err != nil {
		panic(err)
	}

	accountsClient, err := accountsapi.GetClient()
	if err != nil {
		panic(err)
	}

	// Add the various tools to the MCP server
	// Each tool is registered with its corresponding handler function.
	// This allows clients to invoke the tools via the MCP protocol.

	// Basic cocktail tools (no authentication required)
	mcpServer.AddTool(tools.CocktailGetTool, server.ToolHandlerFunc(tools.NewCocktailGetToolHandler(cocktailsClient).Handle))
	mcpServer.AddTool(tools.CocktailSearchTool, server.ToolHandlerFunc(tools.NewCocktailSearchToolHandler(aiSearchClient).Handle))

	// Simple formating and cleaning tools (no authentication required)
	mcpServer.AddTool(tools.ConvertToPlainTextTool, server.ToolHandlerFunc(tools.NewConvertToPlainTextToolHandler().Handle))

	// Authentication tools
	mcpServer.AddTool(tools.AuthLoginTool, server.ToolHandlerFunc(tools.NewAuthLoginToolHandler(authManager).Handle))
	mcpServer.AddTool(tools.AuthStatusTool, server.ToolHandlerFunc(tools.NewAuthStatusToolHandler(authManager).Handle))
	mcpServer.AddTool(tools.AuthLogoutTool, server.ToolHandlerFunc(tools.NewAuthLogoutToolHandler(authManager).Handle))

	// Account Authenticated tools (require user login)
	mcpServer.AddTool(tools.RateCocktailTool, server.ToolHandlerFunc(tools.NewRateCocktailToolHandler(authManager, accountsClient).Handle))

	// Finally, start the server in the chosen mode
	// Proper error handling ensures that any issues during startup are logged.
	// The server will run until it is manually stopped or encounters a fatal error.
	httpServer := mcpserver.NewMCPHTTPServer(
		fmt.Sprintf(":%d", settings.Port),
		mcpServer,
		Version,
	)

	// Run background init job (ensures database and tables exist after a delay)
	go background.RunInitJob(context.Background(), pool, settings)

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

	if settings.AccountsAPIHost == "" {
		telemetry.Logger.Warn().Msg("Warning: ACCOUNTS_API_HOST is not set")
	}

	if settings.AccountsAPISubscriptionKey == "" {
		telemetry.Logger.Warn().Msg("Warning: ACCOUNTS_API_XKEY is not set")
	}

	if settings.CezzisBaseURL == "" {
		telemetry.Logger.Warn().Msg("Warning: CEZZIS_BASE_URL is not set")
	}

	assertAuth0Settings(settings)
	assertPostgresSettings(settings)
	assertInitJobSettings(settings)
	assertOtlpSettings(settings)
}

func assertAuth0Settings(settings *config.AppSettings) {
	if settings.Auth0Domain == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_DOMAIN is not set; authentication will fail")
	}

	if settings.Auth0NativeClientID == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_NATIVE_CLIENT_ID is not set; authentication will fail")
	}

	if settings.Auth0AccountsAPIAudience == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_ACCOUNTS_API_AUDIENCE is not set; authentication will fail")
	}

	if settings.Auth0Scopes == "" {
		telemetry.Logger.Warn().Msg("Warning: AUTH0_SCOPES is not set; authentication will fail")
	}
}

func assertPostgresSettings(settings *config.AppSettings) {
	if settings.PostgresHost == "" {
		telemetry.Logger.Warn().Msg("Warning: POSTGRES_HOST is not set; database access will fail")
	}

	if settings.PostgresPort == 0 {
		telemetry.Logger.Warn().Msg("Warning: POSTGRES_PORT is not set; database access will fail")
	}

	if settings.PostgresDBName == "" {
		telemetry.Logger.Warn().Msg("Warning: POSTGRES_DB is not set; database access will fail")
	}

	if settings.PostgresUser == "" {
		telemetry.Logger.Warn().Msg("Warning: POSTGRES_USER is not set; database access will fail")
	}

	if settings.PostgresPassword == "" {
		telemetry.Logger.Warn().Msg("Warning: POSTGRES_PASSWORD is not set; database access will fail")
	}
}

func assertInitJobSettings(settings *config.AppSettings) {
	if settings.InitJobEnabled {
		telemetry.Logger.Info().
			Int("delay_seconds", settings.InitDelaySeconds).
			Msg("Background init job is enabled")
	} else {
		telemetry.Logger.Info().Msg("Background init job is disabled")
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
