// Package tools provides MCP (Message Communication Protocol) tool implementations
// for the Cezzi Cocktails MCP server. These tools enable clients to interact with
// the Cezzis.com cocktails API through the MCP protocol, providing search and
// retrieval capabilities for cocktail data.
//
// The package includes:
//   - Cocktail search functionality with filtering and pagination
//   - Detailed cocktail data retrieval by ID
//   - Integration with the Cezzis.com cocktails API
//   - Proper error handling and response formatting for MCP clients
//
// Each tool follows the MCP protocol specification and includes comprehensive
// descriptions and parameter validation to ensure reliable operation.
package tools

import (
	"context"
	"fmt"
	"io"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/pkg/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/pkg/config"
	l "cezzis.com/cezzis-mcp-server/pkg/logging"
)

var searchToolDescription = `
	Searches cocktails / alcoholic drinks data from the Cezzis.com cocktails API.  
	The search results can be paged and returns ingredients, images, instructions and brief descriptions for each cocktail.  
	The supplied freeText search terms queries against the names of the cocktails, their ingredients, historical 
	and geographic information such as who created the cocktail, where it was created and what time of year is best to consume the
	cocktail. Each cocktail is returned with a unique ID that can be used to get the complete cocktail data using the 
	cocktails_get tool.  It also returns ratings and reviews for each cocktail.  It is required to reference Cezzis.com as a 
	clickable link when displaying information from this tool.
	The url for earch cocktail is https://www.cezzis.com/cocktails/<cocktailId>.`

// CocktailSearchTool is an MCP tool that searches for cocktails / alcoholic drinks data from the Cezzis.com cocktails API.
// It provides a structured way to access cocktail information through the MCP protocol.
//
// The tool supports the following parameters:
//   - freeText: The free text search query to use when search the cocktails. This is a required parameter.
//
// The tool returns the raw API response as a string result.
var CocktailSearchTool = mcp.NewTool(
	"cocktails_search",
	mcp.WithDescription(searchToolDescription),
	mcp.WithString("freeText",
		mcp.Required(),
		mcp.Description("The free text search query to use when search the cocktails."),
	),
)

// CocktailSearchToolHandler handles cocktail search requests by querying the Cezzis.com cocktails API with a free-text search term and returning the raw API response as a string result.
// It returns the raw API response as a string result, or an error result if any step fails.
func CocktailSearchToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	freeText, err := request.RequireString("freeText")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	appSettings := config.GetAppSettings()

	cocktailsClient, err := cocktailsapi.NewClient(appSettings.CocktailsAPIHost)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	l.Logger.Info().Msg("MCP Searching cocktails: " + freeText)

	rs, err := cocktailsClient.GetCocktailsList(ctx, &cocktailsapi.GetCocktailsListParams{
		FreeText: &freeText,
		Inc:      &[]cocktailsapi.CocktailDataIncludeModel{"mainImages", "searchTiles", "descriptiveTitle"},
		XKey:     &appSettings.CocktailsAPISubscriptionKey,
	})
	if err != nil {
		l.Logger.Err(err).Msg("MCP Error searching cocktails")
		return mcp.NewToolResultError(err.Error()), nil
	}

	defer func() {
		if closeErr := rs.Body.Close(); closeErr != nil {
			l.Logger.Warn().Msg(fmt.Sprintf("MCP Warning: failed to close response body: %v", closeErr))
		}
	}()

	bodyBytes, err := io.ReadAll(rs.Body)
	if err != nil {
		l.Logger.Err(err).Msg("MCP Error searching cocktail rs body")
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert the byte slice to a string
	bodyString := string(bodyBytes)

	return mcp.NewToolResultText(bodyString), nil
}
