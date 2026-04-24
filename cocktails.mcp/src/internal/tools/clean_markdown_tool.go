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

	"github.com/huantt/plaintext-extractor/markdown"
	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var cleanMarkdownToolDescription = `Cleans markdown content and returns plain text. This tool is designed to process markdown content, such as cocktail descriptions 
	or recipes, and remove all markdown formatting to return clean, plain text.
	
	It can be used to extract readable content from markdown-formatted data retrieved from the Cezzis.com cocktails API.

	This tool does not require authentication and can be used without an account.`

// CleanMarkdownTool is an MCP tool that cleans markdown content and returns plain text.
// It provides a structured way to access cleaned markdown content through the MCP protocol.
//
// The tool supports the following parameters:
//   - content: The markdown content to clean. This is a required parameter.
//
// The tool returns the cleaned markdown content as a string result.
var CleanMarkdownTool = mcp.NewTool(
	"clean_markdown",
	mcp.WithDescription(cleanMarkdownToolDescription),
	mcp.WithString("content",
		mcp.Required(),
		mcp.Description("The markdown content to clean. This can include cocktail information in markdown format."),
	),
)

// CleanMarkdownToolHandler handles requests to clean markdown content through the MCP protocol.
// It validates the input parameters, processes the markdown content to remove formatting, and returns the cleaned text as a result.
// The handler ensures that the required "content" parameter is provided and is not empty before performing the cleaning operation.
type CleanMarkdownToolHandler struct{}

// NewCleanMarkdownToolHandler creates a new instance of CleanMarkdownToolHandler with the provided API factory.
func NewCleanMarkdownToolHandler() *CleanMarkdownToolHandler {
	return &CleanMarkdownToolHandler{}
}

// Handle processes incoming MCP requests for the CleanMarkdownTool. It validates the request parameters,
// cleans the provided markdown content, and returns the cleaned text as a result.
func (handler CleanMarkdownToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	telemetry.Logger.Info().Ctx(ctx).Msg("MCP cleaning markdown content")

	// Clean the markdown content
	cleanedContent, err := cleanMarkdown(content)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), err
	}

	return mcp.NewToolResultText(cleanedContent), nil
}

func cleanMarkdown(content string) (string, error) {
	extractor := markdown.NewExtractor()
	cleaned, err := extractor.PlainText(content)
	if err != nil {
		return "", err
	}
	return *cleaned, nil
}
