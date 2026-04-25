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
	"github.com/huantt/plaintext-extractor/html"
	"github.com/huantt/plaintext-extractor/markdown"
	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var convertToPlainTextToolDescription = `Cleans textual context removing markdown syntax, HTML tags, special JSON characters, and emojis, and returns plain text. This tool is designed to process markdown content, such as cocktail descriptions 
	or recipes, and remove all formatting to return clean, plain text.
	
	It can be used to extract readable content from markdown-formatted data retrieved from the Cezzis.com cocktails API.

	This tool does not require authentication and can be used without an account.`

// ConvertToPlainTextTool is an MCP tool that cleans textual content removing markdown syntax, HTML tags, special JSON characters, and emojis, and returns plain text.
// It provides a structured way to access clean cocktail descriptions through the MCP protocol.
//
// The tool supports the following parameters:
//   - content: The content to clean. This is a required parameter.
//
// The tool returns the clean plaintext content as a string.
var ConvertToPlainTextTool = mcp.NewTool(
	"convert_to_plaintext",
	mcp.WithDescription(convertToPlainTextToolDescription),
	mcp.WithString("content",
		mcp.Required(),
		mcp.Description("The content to convert to plain text. This can include cocktail information in markdown format, HTML, or other text formats."),
	),
)

// ConvertToPlainTextToolHandler handles requests to convert textual content to plain text through the MCP protocol.
// It validates the input parameters, processes the content to remove formatting, HTML tags, special JSON characters, and emojis, and returns the plain text as a result.
// The handler ensures that the required "content" parameter is provided and is not empty before performing the conversion.
type ConvertToPlainTextToolHandler struct{}

// NewConvertToPlainTextToolHandler creates a new instance of ConvertToPlainTextToolHandler with the provided API factory.
func NewConvertToPlainTextToolHandler() *ConvertToPlainTextToolHandler {
	return &ConvertToPlainTextToolHandler{}
}

// Handle processes incoming MCP requests for the ConvertToPlainTextTool. It validates the request parameters,
// converts the provided content to plain text, and returns the plain text as a result.
func (handler ConvertToPlainTextToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	telemetry.Logger.Info().Ctx(ctx).Msg("ConvertToPlainTextTool: Converting content to plain text")

	// Convert the markdown content to plain text
	plainTextContent, err := cleanMarkdown(content, ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	// Clean the HTML content
	plainTextContent, err = cleanHTML(plainTextContent, ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	// Clean the content by removing special JSON characters
	plainTextContent = cleanSpecialJSONCharacters(plainTextContent, ctx)

	// Clean the content by removing emojis
	plainTextContent = cleanEmojis(plainTextContent, ctx)

	return mcp.NewToolResultText(plainTextContent), nil
}

func cleanMarkdown(content string, ctx context.Context) (string, error) {
	telemetry.Logger.Info().Ctx(ctx).Msg("Cleaning markdown syntax from content")

	extractor := markdown.NewExtractor()
	cleaned, err := extractor.PlainText(content)
	if err != nil {
		return "", err
	}
	if cleaned == nil {
		return "", nil
	}
	return *cleaned, nil
}

func cleanEmojis(content string, ctx context.Context) string {
	telemetry.Logger.Info().Ctx(ctx).Msg("Removing emojis from content")

	return gomoji.RemoveEmojis(content)
}

func cleanHTML(content string, ctx context.Context) (string, error) {
	telemetry.Logger.Info().Ctx(ctx).Msg("Cleaning HTML tags from content")

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

func cleanSpecialJSONCharacters(content string, ctx context.Context) string {
	telemetry.Logger.Info().Ctx(ctx).Msg("Cleaning special JSON characters from content")

	r := strings.NewReplacer("\"", "", "'", "", "\\", "")
	cleanedText := r.Replace(content)
	return cleanedText
}
