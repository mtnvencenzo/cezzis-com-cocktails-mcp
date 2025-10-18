package tools

import (
	"context"
	"errors"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/mcpserver"
)

var authStatusDescription = `
	Use this tool to check the current authentication status for Cezzis.com account access.
	This tool returns whether you are currently authenticated and can access
	personalized features like cocktail favorites, ratings, and profile management for your Cezzis.com account.

	You must have a valid and active mcp session, the session identifier from the original initialization request must be present in the request
	to this tool via the Mcp-Session-Id header.
`

// AuthStatusTool checks current authentication status
var AuthStatusTool = mcp.NewTool(
	"auth_status",
	mcp.WithDescription(authStatusDescription),
)

// AuthStatusToolHandler handles authentication status requests
type AuthStatusToolHandler struct {
	authManager *auth.OAuthFlowManager
}

// NewAuthStatusToolHandler creates a new authentication status handler
func NewAuthStatusToolHandler(authManager *auth.OAuthFlowManager) *AuthStatusToolHandler {
	return &AuthStatusToolHandler{
		authManager: authManager,
	}
}

// Handle handles authentication status requests
func (handler *AuthStatusToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := ctx.Value(mcpserver.McpSessionIDKey)
	if sessionID == nil || sessionID == "" {
		err := errors.New("missing required Mcp-Session-Id header")
		return mcp.NewToolResultError(err.Error()), err
	}

	if handler.authManager.IsAuthenticated(ctx, sessionID.(string)) {
		return mcp.NewToolResultText("You are currently authenticated and can access personalized features."), nil
	}

	return mcp.NewToolResultText("You are not currently authenticated. Use the 'authentication_login_flow' tool to sign in."), nil
}
