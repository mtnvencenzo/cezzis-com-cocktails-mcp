package main

import (
	"context"
	"fmt"
	"io"

	"cezzis.com/cezzis-mcp-server/pkg/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	mcpServer := server.NewMCPServer(
		"Cezzi Cocktails Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	searchTool := mcp.NewTool(
		"cocktails_search",
		mcp.WithDescription("Search cocktails data from the cocktails API"),
		mcp.WithString("freeText",
			mcp.Required(),
			mcp.Description("The free text search query"),
		),
	)

	mcpServer.AddTool(searchTool, server.ToolHandlerFunc(searchToolHandler))

	// Start the stdio server
	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func searchToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	freeText, err := request.RequireString("freeText")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	config := config.GetAppSettings()

	cocktailsClient, err := cocktailsapi.NewClient(config.CocktailsApiHost)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	rs, err := cocktailsClient.GetCocktailsList(ctx, &cocktailsapi.GetCocktailsListParams{
		FreeText: &freeText,
		XKey:     &config.CocktailsApiSubscriptionKey,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert the byte slice to a string
	bodyString := string(bodyBytes)

	return mcp.NewToolResultError(bodyString), nil
}
