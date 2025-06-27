package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cezzis.com/cezzis-mcp-server/api/cocktailsapi"
	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Server error - exe path: %v\n", err)
	}

	// Get the directory of the executable
	exeDir := filepath.Dir(exePath)

	_ = godotenv.Overload(
		fmt.Sprintf("%s/%s", exeDir, ".env"),
		fmt.Sprintf("%s/%s", exeDir, ".env.local"))

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

	apiHost := os.Getenv("COCKTAILS_API_HOST")
	xKey := os.Getenv("COCKTAILS_API_XKEY")

	cocktailsClient, err := cocktailsapi.NewClient(apiHost)
	if err != nil {
		return nil, err
	}

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
