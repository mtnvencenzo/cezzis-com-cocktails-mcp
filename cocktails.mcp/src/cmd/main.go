package main

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"

	"cezzis.com/cezzis-mcp-server/pkg/tools"
)

// main initializes and runs the Cezzi Cocktails MCP server, registering cocktail search and retrieval tools and serving requests over standard input/output.
func main() {
	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	mcpServer.AddTool(tools.CocktailSearchTool, server.ToolHandlerFunc(tools.CocktailSearchToolHandler))
	mcpServer.AddTool(tools.CocktailGetTool, server.ToolHandlerFunc(tools.CocktailGetToolHandler))

	// Start the stdio server
	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
