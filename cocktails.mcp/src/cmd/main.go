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
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/server"

	l "cezzis.com/cezzis-mcp-server/internal/logging"
	internalServer "cezzis.com/cezzis-mcp-server/internal/server"
	"cezzis.com/cezzis-mcp-server/internal/tools"
)

// Version uses build linkers to set this value at build time
var Version string = "0.0.0"

// main initializes and runs the Cezzi Cocktails MCP server, registering cocktail search and retrieval tools and serving requests over standard input/output or HTTP.
func main() {

	// Set up environment variables from .env files in the executable directory
	// This allows configuration settings to be loaded at runtime.
	exePath, oserr := os.Executable()
	if oserr != nil {
		fmt.Printf("Server error - exe path: %v\n", oserr)
	}

	e := os.Getenv("ENV")

	_ = godotenv.Overload(
		fmt.Sprintf("%s/%s", exePath, ".env"),
		fmt.Sprintf("%s/%s.%s", exePath, ".env", e))

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

	// Add the carious tools to the MCP server
	// Each tool is registered with its corresponding handler function.
	// This allows clients to invoke the tools via the MCP protocol.
	mcpServer.AddTool(tools.CocktailSearchTool, server.ToolHandlerFunc(tools.CocktailSearchToolHandler))
	mcpServer.AddTool(tools.CocktailGetTool, server.ToolHandlerFunc(tools.CocktailGetToolHandler))

	// Finally, start the server in the chosen mode
	// If --http is provided, start the HTTP server with logging middleware and a health check endpoint.
	// Otherwise, serve requests over stdio.
	// Proper error handling ensures that any issues during startup are logged.
	// The server will run until it is manually stopped or encounters a fatal error.
	if *httpAddr != "" {
		// HTTP mode
		httpServer := internalServer.NewMCPHTTPServer(*httpAddr, mcpServer, Version)
		l.Logger.Fatal().Err(httpServer.Start()).Msg("MCP Server Closed")
	} else {
		// Stdio mode (default)
		if err := server.ServeStdio(mcpServer); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}
}
