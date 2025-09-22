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
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/auth"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
	internalServer "cezzis.com/cezzis-mcp-server/internal/server"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

// Version uses build linkers to set this value at build time
var Version string = "0.0.0"

// main initializes and runs the Cezzi Cocktails MCP server, registering cocktail search and retrieval tools and serving requests over standard input/output or HTTP.
func main() {
	loadEnv()

	// Initialize the logger
	_, err := l.InitLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Add a flag to choose between stdio and HTTP
	// If --http is provided, the server will run in HTTP mode on the specified address.
	// Otherwise, it will default to stdio mode.
	httpAddr := flag.String("http", "", "If set, serve HTTP on this address (e.g., :8080). Otherwise, use stdio.")
	flag.Parse()

	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// Initialize authentication manager
	authManager := auth.NewManager()

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

	// Authenticated tools (require user login)
	mcpServer.AddTool(tools.RateCocktailTool, server.ToolHandlerFunc(tools.NewRateCocktailToolHandler(authManager, authCocktailsAPIFactory).Handle))
	mcpServer.AddTool(tools.GetFavoritesTool, server.ToolHandlerFunc(tools.NewGetFavoritesToolHandler(authManager, authCocktailsAPIFactory).Handle))

	// Finally, start the server in the chosen mode
	// If --http is provided, start the HTTP server with logging middleware and a health check endpoint.
	// Otherwise, serve requests over stdio.
	// Proper error handling ensures that any issues during startup are logged.
	// The server will run until it is manually stopped or encounters a fatal error.
	if *httpAddr != "" {
		// HTTP mode
		httpServer := internalServer.NewMCPHTTPServer(*httpAddr, mcpServer, Version)

		if err := httpServer.Start(); err != nil {
			l.Logger.Fatal().Err(err).Msg("MCP HTTP server failed")
		}

		l.Logger.Info().Msg("MCP HTTP server stopped")
	} else {
		// Stdio mode (default)
		if err := server.ServeStdio(mcpServer); err != nil {
			l.Logger.Err(err).Msg("Server error: %v\n")
		}
	}
}

func loadEnv() {
	// Set up environment variables from .env files in the executable directory
	// This allows configuration settings to be loaded at runtime.
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	envFileDir := exeDir

	if os.Getenv("ENV_DIR_OVERRIDE") != "" {
		envFileDir = os.Getenv("ENV_DIR_OVERRIDE")
	}

	env := os.Getenv("ENV")
	baseEnvFile := filepath.Join(envFileDir, ".env")
	candidates := []string{baseEnvFile}

	if env != "" {
		candidates = append(candidates, baseEnvFile+"."+env)
	}

	toLoad := make([]string, 0, len(candidates))
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			toLoad = append(toLoad, f)
		}
	}

	if len(toLoad) > 0 {
		_ = godotenv.Overload(toLoad...)
	}
}
