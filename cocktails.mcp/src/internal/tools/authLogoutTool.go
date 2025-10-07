package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

var authLogoutDescription = `
	Logs you out by clearing stored authentication tokens.
	Use this if you changed scopes, audiences, or want to switch accounts.
`

// AuthLogoutTool clears stored tokens and in-memory session
var AuthLogoutTool = mcp.NewTool(
	"auth_logout",
	mcp.WithDescription(authLogoutDescription),
)

// AuthLogoutToolHandler handles logout requests
type AuthLogoutToolHandler struct {
	authManager *auth.Manager
}

// NewAuthLogoutToolHandler creates a new logout handler
func NewAuthLogoutToolHandler(authManager *auth.Manager) *AuthLogoutToolHandler {
	return &AuthLogoutToolHandler{authManager: authManager}
}

// Handle processes logout by clearing tokens from memory and disk
func (handler *AuthLogoutToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := handler.authManager.Logout(); err != nil {
		l.Logger.Error().Err(err).Msg("Failed to logout and clear tokens")
		return mcp.NewToolResultError(fmt.Sprintf("Logout failed: %v", err)), nil
	}
	return mcp.NewToolResultText("✅ Logged out. Tokens cleared. Use 'auth_login' to sign in again."), nil
}
