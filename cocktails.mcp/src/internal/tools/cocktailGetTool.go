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
	"errors"
	"fmt"
	"io"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/config"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

var getToolDescription = `
	Gets the complete cocktail / alcoholic drink data from the Cezzis.com cocktails API.  
	The cocktail data includes ingredients images, and instructions, historical and geographic information, 
	descriptions and instructions for each cocktail.  It also returns ratings and reviews for each cocktail.
	It is required to reference Cezzis.com as a clickable link when displaying information from this tool.
	The url for the cocktail is https://www.cezzis.com/cocktails/<cocktailId>.`

// CocktailGetTool is an MCP tool that retrieves detailed cocktail data from the Cezzis.com cocktails API.
// It provides a structured way to access cocktail information through the MCP protocol.
//
// The tool supports the following parameters:
//   - cocktailId: The ID of the cocktail to retrieve. This is a required parameter.
//
// The tool returns the complete cocktail data as a string result.
var CocktailGetTool = mcp.NewTool(
	"cocktails_get",
	mcp.WithDescription(getToolDescription),
	mcp.WithString("cocktailId",
		mcp.Required(),
		mcp.Description("The ID of the cocktail to get.  This can typically be found for each cocktail in the cocktails_search tool results for each cocktail by the 'id' field.  The ID is a unique identifier for each cocktail and is used to get the complete cocktail data."),
	),
)

// CocktailGetToolHandler handles cocktail retrieval requests through the MCP protocol.
// It maintains a reference to the cocktails API factory for making API calls.
type CocktailGetToolHandler struct {
	cocktailsAPIFactory cocktailsapi.ICocktailsAPIFactory
}

// NewCocktailGetToolHandler creates a new instance of CocktailGetToolHandler with the provided API factory.
func NewCocktailGetToolHandler(cocktailsAPIFactory cocktailsapi.ICocktailsAPIFactory) *CocktailGetToolHandler {
	return &CocktailGetToolHandler{
		cocktailsAPIFactory,
	}
}

// Handle handles requests to retrieve detailed cocktail data from the Cezzis.com cocktails API using a provided cocktail ID.
// It returns the full cocktail information as a string result, or an error result if any step fails.
func (handler CocktailGetToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cocktailID, err := request.RequireString("cocktailId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if cocktailID == "" {
		err := errors.New("required argument \"cocktailId\" is empty")
		return mcp.NewToolResultError(err.Error()), err
	}

	appSettings := config.GetAppSettings()

	l.Logger.Info().Msg("MCP Getting cocktail: " + cocktailID)

	cocktailsAPI, cliErr := handler.cocktailsAPIFactory.GetClient()
	if cliErr != nil {
		return mcp.NewToolResultError(cliErr.Error()), cliErr // already logged upstream
	}

	rs, callErr := cocktailsAPI.GetCocktail(ctx, cocktailID, &cocktailsapi.GetCocktailParams{
		XKey: &appSettings.CocktailsAPISubscriptionKey,
	})

	if callErr != nil {
		l.Logger.Err(callErr).Msg("MCP Error getting cocktail: " + cocktailID)
		return mcp.NewToolResultError(callErr.Error()), callErr
	}

	defer func() {
		if closeErr := rs.Body.Close(); closeErr != nil {
			l.Logger.Warn().Msg(fmt.Sprintf("MCP Warning: failed to close response body: %v", closeErr))
		}
	}()

	bodyBytes, readErr := io.ReadAll(rs.Body)
	if readErr != nil {
		l.Logger.Err(readErr).Msg("MCP Error getting cocktail rs body: " + cocktailID)
		return mcp.NewToolResultError(readErr.Error()), readErr
	}

	// Convert the byte slice to a string
	bodyString := string(bodyBytes)

	return mcp.NewToolResultText(bodyString), nil
}
