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
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var removeSpecialJSONCharactersToolDescription = `Removes special JSON characters from the provided content and returns plain text.

	This tool is designed to process content that may contain special JSON characters, such as cocktail descriptions 
	or recipes, and remove all special JSON characters to return clean, plain text.

	It can be used to extract readable content from data retrieved from the Cezzis.com cocktails API.

	This tool does not require authentication and can be used without an account.`

// RemoveSpecialJSONCharactersTool is an MCP tool that removes special JSON characters from the provided content and returns plain text.
// It provides a structured way to access cleaned content through the MCP protocol.
//
// The tool supports the following parameters:
//   - content: The content to clean. This is a required parameter.
//
// The tool returns the cleaned content as a string result.
var RemoveSpecialJSONCharactersTool = mcp.NewTool(
	"remove_special_json_characters",
	mcp.WithDescription(removeSpecialJSONCharactersToolDescription),
	mcp.WithString("content",
		mcp.Required(),
		mcp.Description("The content to clean. This can include cocktail information in HTML format, markdown format, or just plain text that may contain special JSON characters."),
	),
)

// RemoveSpecialJSONCharactersToolHandler handles requests to remove special JSON characters from content through the MCP protocol.
type RemoveSpecialJSONCharactersToolHandler struct{}

// NewRemoveSpecialJSONCharactersToolHandler creates a new instance of RemoveSpecialJSONCharactersToolHandler.
func NewRemoveSpecialJSONCharactersToolHandler() *RemoveSpecialJSONCharactersToolHandler {
	return &RemoveSpecialJSONCharactersToolHandler{}
}

// Handle handles requests to remove special JSON characters from content through the MCP protocol.
// It returns the cleaned content as a string result, or an error result if any step fails.
func (handler RemoveSpecialJSONCharactersToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := ctx.Value(middleware.McpSessionIDKey)
	if sessionID == nil || sessionID == "" {
		err := errors.New("missing required Mcp-Session-Id header")
		return mcp.NewToolResultError(err.Error()), err
	}

	// Validate and extract the content parameter
	content, err := request.RequireString("content")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	if strings.TrimSpace(content) == "" {
		err := errors.New("required argument \"content\" is empty")
		return mcp.NewToolResultError(err.Error()), err
	}

	telemetry.Logger.Info().Ctx(ctx).Msg("MCP Removing special JSON characters from content")

	// Clean the content by removing special JSON characters
	cleanedContent := cleanSpecialJSONCharacters(content)

	return mcp.NewToolResultText(cleanedContent), nil
}

func cleanSpecialJSONCharacters(content string) string {
	r := strings.NewReplacer("\"", "", "'", "", "\\", "")
	cleanedText := r.Replace(content)
	return cleanedText
}
