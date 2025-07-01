package tools

import (
	"context"
	"io"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/pkg/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/pkg/config"
)

var getToolDescription = `
	Gets the complete cocktail / alcoholic drink data from the Cezzis.com cocktails API.  
	The cocktail data includes ingredients images, and instructions, historical and geographic information, 
	descriptions and instructions for each cocktail.  It also returns ratings and reviews for each cocktail.
	It is required to reference Cezzis.com as a clickable link when displaying information from this tool.
	The url for the cocktail is https://www.cezzis.com/cocktails/<cocktailId>.`

var CocktailGetTool = mcp.NewTool(
	"cocktails_get",
	mcp.WithDescription(getToolDescription),
	mcp.WithString("cocktailId",
		mcp.Required(),
		mcp.Description("The ID of the cocktail to get.  This can typically be found for each cocktail in the cocktails_search tool results for each cocktail by the 'id' field.  The ID is a unique identifier for each cocktail and is used to get the complete cocktail data."),
	),
)

// CocktailGetToolHandler handles requests to retrieve detailed cocktail data from the Cezzis.com cocktails API using a provided cocktail ID.
// It returns the full cocktail information as a string result, or an error result if any step fails.
func CocktailGetToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cocktailId, err := request.RequireString("cocktailId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	config := config.GetAppSettings()

	cocktailsClient, err := cocktailsapi.NewClient(config.CocktailsApiHost)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	rs, err := cocktailsClient.GetCocktail(ctx, cocktailId, &cocktailsapi.GetCocktailParams{
		XKey: &config.CocktailsApiSubscriptionKey,
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
