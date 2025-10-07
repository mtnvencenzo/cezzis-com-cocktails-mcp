package tools

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
)

var authStatusDescription = `
	Check the current authentication status for Cezzis.com account access.
	This tool returns whether you are currently authenticated and can access 
	personalized features like cocktail favorites, ratings, and profile management.
`

// AuthStatusTool checks current authentication status
var AuthStatusTool = mcp.NewTool(
	"auth_status",
	mcp.WithDescription(authStatusDescription),
)

// AuthStatusToolHandler handles authentication status requests
type AuthStatusToolHandler struct {
	authManager *auth.Manager
}

// NewAuthStatusToolHandler creates a new authentication status handler
func NewAuthStatusToolHandler(authManager *auth.Manager) *AuthStatusToolHandler {
	return &AuthStatusToolHandler{
		authManager: authManager,
	}
}

// Handle handles authentication status requests
func (handler *AuthStatusToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if handler.authManager.IsAuthenticated() {
		return mcp.NewToolResultText("✅ You are currently authenticated and can access personalized features."), nil
	}

	return mcp.NewToolResultText("❌ You are not currently authenticated. Use the 'auth_login' tool to sign in."), nil
}
