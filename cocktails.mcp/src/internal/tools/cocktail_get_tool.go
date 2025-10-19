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
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var getToolDescription = `
	Gets the complete cocktail recipe data from the Cezzis.com cocktails API for a given cocktailId.

	The cocktail data includes ingredients images, and instructions, historical and geographic information, 
	descriptions and instructions for each cocktail.  It also returns ratings and reviews for each cocktail.

	It is required to reference Cezzis.com as a clickable link when displaying cocktail information from this tool.
	The url for each cocktail is formatted as https://www.cezzis.com/cocktails/<cocktailId>.

	This tool does not require authentication and can be used without an account.`

// CocktailGetTool is an MCP tool that retrieves detailed cocktail data from the Cezzis.com cocktails API.
// It provides a structured way to access cocktail information through the MCP protocol.
//
// The tool supports the following parameters:
//   - cocktailId: The ID of the cocktail to retrieve. This is a required parameter.
//
// The tool returns the complete cocktail data as a string result.
var CocktailGetTool = mcp.NewTool(
	"get_cocktail",
	mcp.WithDescription(getToolDescription),
	mcp.WithString("cocktailId",
		mcp.Required(),
		mcp.Description("The ID of the cocktail to get.  This can typically be found for each cocktail in the search_cocktails tool results for each cocktail by the 'id' field.  The ID is a unique identifier for each cocktail and is used to get the complete cocktail data."),
	),
)

// CocktailGetToolHandler handles cocktail retrieval requests through the MCP protocol.
// It maintains a reference to the cocktails API factory for making API calls.
type CocktailGetToolHandler struct {
	client *cocktailsapi.Client
}

// NewCocktailGetToolHandler creates a new instance of CocktailGetToolHandler with the provided API factory.
func NewCocktailGetToolHandler(client *cocktailsapi.Client) *CocktailGetToolHandler {
	return &CocktailGetToolHandler{
		client: client,
	}
}

// Handle handles requests to retrieve detailed cocktail data from the Cezzis.com cocktails API using a provided cocktail ID.
// It returns the full cocktail information as a string result, or an error result if any step fails.
func (handler CocktailGetToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := ctx.Value(middleware.McpSessionIDKey)
	if sessionID == nil || sessionID == "" {
		err := errors.New("missing required Mcp-Session-Id header")
		return mcp.NewToolResultError(err.Error()), err
	}

	// Validate and extract the cocktailId parameter
	cocktailID, err := request.RequireString("cocktailId")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if strings.TrimSpace(cocktailID) == "" {
		err := errors.New("required argument \"cocktailId\" is empty")
		return mcp.NewToolResultError(err.Error()), err
	}

	appSettings := config.GetAppSettings()

	telemetry.Logger.Info().Ctx(ctx).Msg("MCP Getting cocktail: " + cocktailID)

	// default to a safe deadline if none present
	callCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		callCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	rs, callErr := handler.client.GetCocktail(callCtx, cocktailID, &cocktailsapi.GetCocktailParams{
		XKey: &appSettings.CocktailsAPISubscriptionKey,
	})

	if callErr != nil {
		telemetry.Logger.Error().Ctx(ctx).Err(callErr).Msg("MCP Error getting cocktail: " + cocktailID)
		return mcp.NewToolResultError(callErr.Error()), callErr
	}

	defer func() {
		if closeErr := rs.Body.Close(); closeErr != nil {
			telemetry.Logger.Error().Ctx(ctx).Msg(fmt.Sprintf("MCP Warning: failed to close response body: %v", closeErr))
		}
	}()

	bodyBytes, readErr := io.ReadAll(rs.Body)
	if readErr != nil {
		telemetry.Logger.Error().Err(readErr).Ctx(ctx).Msg("MCP Error getting cocktail rs body: " + cocktailID)
		return mcp.NewToolResultError(readErr.Error()), readErr
	}

	// Convert the byte slice to a string
	bodyString := string(bodyBytes)

	return mcp.NewToolResultText(bodyString), nil
}
