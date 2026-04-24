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

	"github.com/huantt/plaintext-extractor/html"
	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var removeHTMLTagsToolDescription = `Removes HTML tags from the provided content and returns plain text.

	This tool is designed to process HTML content, plain text, or markdown content that contains HTML tags, such as cocktail descriptions 
	or recipes, and remove all HTML tags to return clean, plain text.

	It can be used to extract readable content from HTML-formatted data retrieved from the Cezzis.com cocktails API.

	This tool does not require authentication and can be used without an account.`

// RemoveHTMLTagsTool is an MCP tool that removes HTML tags from the provided content and returns plain text.
// It provides a structured way to access cleaned HTML content through the MCP protocol.
//
// The tool supports the following parameters:
//   - content: The HTML content to clean. This is a required parameter.
//
// The tool returns the cleaned HTML content as a string result.
var RemoveHTMLTagsTool = mcp.NewTool(
	"remove_html_tags",
	mcp.WithDescription(removeHTMLTagsToolDescription),
	mcp.WithString("content",
		mcp.Required(),
		mcp.Description("The content to clean. This can include cocktail information in HTML format, plain text or markdown."),
	),
)

// RemoveHTMLTagsToolHandler handles requests to remove HTML tags from content through the MCP protocol.
type RemoveHTMLTagsToolHandler struct{}

// NewRemoveHTMLTagsToolHandler creates a new instance of RemoveHTMLTagsToolHandler.
func NewRemoveHTMLTagsToolHandler() *RemoveHTMLTagsToolHandler {
	return &RemoveHTMLTagsToolHandler{}
}

// Handle handles requests to remove HTML tags from content through the MCP protocol.
// It returns the cleaned HTML content as a string result, or an error result if any step fails.
func (handler RemoveHTMLTagsToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	telemetry.Logger.Info().Ctx(ctx).Msg("MCP cleaning HTML content")

	// Clean the HTML content
	cleanedContent, err := cleanHTML(content)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	return mcp.NewToolResultText(cleanedContent), nil
}

func cleanHTML(content string) (string, error) {
	extractor := html.NewExtractor()
	cleaned, err := extractor.PlainText(content)
	if err != nil {
		return "", err
	}
	if cleaned == nil {
		return "", nil
	}
	return *cleaned, nil
}
