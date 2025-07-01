package tools

import (
	"context"
	"io"

	"cezzis.com/cezzis-mcp-server/pkg/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
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

var CocktailSearchTool = mcp.NewTool(
	"cocktails_search",
	mcp.WithDescription(searchToolDescription),
	mcp.WithString("freeText",
		mcp.Required(),
		mcp.Description("The free text search query to use when search the cocktails."),
	),
)

func CocktailSearchToolHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		Inc:      &[]cocktailsapi.CocktailDataIncludeModel{"mainImages", "searchTiles", "descriptiveTitle"},
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
