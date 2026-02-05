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
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/api/aisearch"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var searchToolDescription = fmt.Sprintf(`
	Searches cocktail recipe data from the Cezzis.com cocktails API. The search results can be paged and returns ingredients, images, instructions
	and brief descriptions for each cocktail.

	The supplied freeText search terms queries against the names and descriptions of the cocktails, their ingredients, historical 
	and geographic information such as who created the cocktail, where it was created and what time of year is best to consume the
	cocktail.

	Each cocktail is returned with a unique ID commonly called the cocktailId that can be used to get the complete cocktail data using the
	get_cocktail tool.

	It is required to reference Cezzis.com as a clickable link when displaying cocktail information from this tool.
	The url for each cocktail is formatted as %[1]s/cocktails/<cocktailId>.

	This tool does not require authentication and can be used without an account.

	Examples of free text search terms include:
	- "cocktail with gin and lime"
	- "whiskey sour"
	- "cocktail with rum and mint"
	- "modern cocktails"`, config.GetAppSettings().CezzisBaseURL)

// CocktailSearchTool is an MCP tool that searches for cocktails / alcoholic drinks data from the Cezzis.com cocktails API.
// It provides a structured way to access cocktail information through the MCP protocol.
//
// The tool supports the following parameters:
//   - freeText: The free text search query to use when search the cocktails. This is a required parameter.
//
// The tool returns the raw API response as a string result.
var CocktailSearchTool = mcp.NewTool(
	"search_cocktails",
	mcp.WithDescription(searchToolDescription),
	mcp.WithString("freeText",
		mcp.Required(),
		mcp.Description("The free text search query to use when searching the cocktails."),
	),
)

// CocktailSearchToolHandler implements the MCP tool handler for searching cocktails.
// It maintains a reference to the cocktails API factory for making API calls.
type CocktailSearchToolHandler struct {
	client *aisearch.Client
}

// NewCocktailSearchToolHandler creates a new instance of CocktailSearchToolHandler with the provided API factory.
// The handler uses the factory to create API clients for searching cocktails.
func NewCocktailSearchToolHandler(client *aisearch.Client) *CocktailSearchToolHandler {
	return &CocktailSearchToolHandler{
		client: client,
	}
}

// Handle handles cocktail search requests by querying the Cezzis.com cocktails API with a free-text search term and returning the raw API response as a string result.
// It returns the raw API response as a string result, or an error result if any step fails.
func (handler CocktailSearchToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := ctx.Value(middleware.McpSessionIDKey)
	if sessionID == nil || sessionID == "" {
		err := errors.New("missing required Mcp-Session-Id header")
		return mcp.NewToolResultError(err.Error()), err
	}

	freeText, err := request.RequireString("freeText")
	ft := freeText
	if len(ft) > 120 {
		ft = ft[:120] + "…"
	}

	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	telemetry.Logger.Info().Ctx(ctx).Msg("MCP Searching cocktails: " + ft)

	// default to a safe deadline if none present
	callCtx := ctx
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		callCtx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	rs, callErr := handler.client.GetV1CocktailsSearch(callCtx, &aisearch.GetV1CocktailsSearchParams{
		Freetext: &freeText,
	}, aisearch.RequestEditor())

	if callErr != nil {
		telemetry.Logger.Err(callErr).Ctx(ctx).Msg("MCP Error searching cocktails")
		return mcp.NewToolResultError(callErr.Error()), callErr
	}

	defer func() {
		if closeErr := rs.Body.Close(); closeErr != nil {
			telemetry.Logger.Warn().Ctx(ctx).Msg(fmt.Sprintf("MCP Warning: failed to close response body: %v", closeErr))
		}
	}()

	bodyBytes, readErr := io.ReadAll(rs.Body)
	if readErr != nil {
		telemetry.Logger.Err(readErr).Ctx(ctx).Msg("MCP Error searching cocktail rs body")
		return mcp.NewToolResultError(readErr.Error()), readErr
	}

	// Convert the byte slice to a string
	bodyString := string(bodyBytes)

	return mcp.NewToolResultText(bodyString), nil
}
