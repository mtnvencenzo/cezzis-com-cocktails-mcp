package main

import (
	"context"
	"fmt"
	"io"

	"cezzis.com/cezzis-mcp-server/api/cocktailsapi"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	//http.Handle("/", cocktailsapi.AuthMiddleware([]string{})(searchHandler))

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

	//log.Println("Starting MCP server on :9191")

	// if err := http.ListenAndServe(":9191", mcpServer.Handler()); err != nil {
	// 	log.Fatalf("could not start server: %s\n", err)
	// }

	// Start the stdio server
	//fmt.Printf("Starting MCP server on stdio")
	if err := server.ServeStdio(mcpServer); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func searchToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	freeText, err := request.RequireString("freeText")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cocktailsClient, err := cocktailsapi.NewClient("https://api.cezzis.com/prd/cocktails")
	if err != nil {
		return nil, err
	}
	xKey := "---"
	rs, err := cocktailsClient.GetCocktailsList(ctx, &cocktailsapi.GetCocktailsListParams{
		FreeText: &freeText,
		XKey:     &xKey,
	})
	if err != nil {
		return nil, err
	}
	defer rs.Body.Close()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert the byte slice to a string
	bodyString := string(bodyBytes)

	return mcp.NewToolResultError(bodyString), nil
}
