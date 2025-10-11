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
	"fmt"
	"log"
	"strconv"

	"github.com/mark3labs/mcp-go/server"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/environment"
	"cezzis.com/cezzis-mcp-server/internal/logging"
	"cezzis.com/cezzis-mcp-server/internal/mcpserver"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

// Version uses build linkers to set this value at build time
var Version string = "0.0.0"

// main initializes and runs the Cezzi Cocktails MCP server, registering cocktail search and retrieval
// tools and serving requests over standard input/output or HTTP.
func main() {
	environment.LoadEnv()

	// Initialize the logger
	_, err := logging.InitLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	settings := config.GetAppSettings()

	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Initialize authentication manager
	authManager := auth.NewOAuthFlowManager()

	// Initialize API factories
	cocktailsAPIFactory := cocktailsapi.NewCocktailsAPIFactory()
	authCocktailsAPIFactory := cocktailsapi.NewAuthenticatedCocktailsAPIFactory(authManager)

	// Add the various tools to the MCP server
	// Each tool is registered with its corresponding handler function.
	// This allows clients to invoke the tools via the MCP protocol.

	// Basic cocktail tools (no authentication required)
	mcpServer.AddTool(tools.CocktailGetTool, server.ToolHandlerFunc(tools.NewCocktailGetToolHandler(cocktailsAPIFactory).Handle))
	mcpServer.AddTool(tools.CocktailSearchTool, server.ToolHandlerFunc(tools.NewCocktailSearchToolHandler(cocktailsAPIFactory).Handle))

	// Authentication tools
	mcpServer.AddTool(tools.AuthLoginTool, server.ToolHandlerFunc(tools.NewAuthLoginToolHandler(authManager).Handle))
	mcpServer.AddTool(tools.AuthStatusTool, server.ToolHandlerFunc(tools.NewAuthStatusToolHandler(authManager).Handle))
	mcpServer.AddTool(tools.AuthLogoutTool, server.ToolHandlerFunc(tools.NewAuthLogoutToolHandler(authManager).Handle))

	// Account Authenticated tools (require user login)
	mcpServer.AddTool(tools.RateCocktailTool, server.ToolHandlerFunc(tools.NewRateCocktailToolHandler(authManager, authCocktailsAPIFactory).Handle))

	// Finally, start the server in the chosen mode
	// Proper error handling ensures that any issues during startup are logged.
	// The server will run until it is manually stopped or encounters a fatal error.
	httpServer := mcpserver.NewMCPHTTPServer(fmt.Sprintf(":%d", settings.Port), mcpServer, Version)

	logging.Logger.Info().
		Str("version", Version).
		Str("port", strconv.Itoa(settings.Port)).
		Msg("Starting Cezzi Cocktails MCP Server")

	if err := httpServer.Start(); err != nil {
		logging.Logger.Fatal().Err(err).Msg("MCP HTTP server failed")
	}

	logging.Logger.Info().Msg("MCP HTTP server stopped")
}
