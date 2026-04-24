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

	"github.com/forPelevin/gomoji"
	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var removeEmojisToolDescription = `Removes emojis from the provided content and returns plain text.

	This tool is designed to process content that may contain emojis, such as cocktail descriptions 
	or recipes, and remove all emojis to return clean, plain text.

	It can be used to extract readable content from data retrieved from the Cezzis.com cocktails API.

	This tool does not require authentication and can be used without an account.`

// RemoveEmojisTool is an MCP tool that removes emojis from the provided content and returns plain text.
// It provides a structured way to access cleaned content through the MCP protocol.
//
// The tool supports the following parameters:
//   - content: The content to clean. This is a required parameter.
//
// The tool returns the cleaned content as a string result.
var RemoveEmojisTool = mcp.NewTool(
	"remove_emojis",
	mcp.WithDescription(removeEmojisToolDescription),
	mcp.WithString("content",
		mcp.Required(),
		mcp.Description("The content to clean. This can include cocktail information in HTML format, plain text or markdown."),
	),
)

// RemoveEmojisToolHandler handles requests to remove emojis from content through the MCP protocol.
type RemoveEmojisToolHandler struct{}

// NewRemoveEmojisToolHandler creates a new instance of RemoveEmojisToolHandler.
func NewRemoveEmojisToolHandler() *RemoveEmojisToolHandler {
	return &RemoveEmojisToolHandler{}
}

// Handle handles requests to remove emojis from content through the MCP protocol.
// It returns the cleaned content as a string result, or an error result if any step fails.
func (handler RemoveEmojisToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	telemetry.Logger.Info().Ctx(ctx).Msg("MCP Removing emojis from content")

	// Clean the content by removing emojis
	cleanedContent := cleanEmojis(content)

	return mcp.NewToolResultText(cleanedContent), nil
}

func cleanEmojis(content string) string {
	return gomoji.RemoveEmojis(content)
}
