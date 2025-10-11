package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

var authLogoutDescription = `
	This tool logs out a session by clearing stored authentication tokens associated with the mcp session.
	Use this when the user closes the chat, wants to switch accounts, or to ensure
	that no authentication tokens remain stored on disk.

	You must have a valid and active mcp session, the session identifier from the original initialization request must be present in the request
	to this tool via the Mcp-Session-Id header.
	
	After calling this tool, you will need to run the 'authentication_login_flow' tool
	to sign in again and obtain new authentication tokens prior to using any authenticated tools.`

// AuthLogoutTool clears stored tokens and in-memory session
var AuthLogoutTool = mcp.NewTool(
	"authentication_logout_flow",
	mcp.WithDescription(authLogoutDescription),
)

// AuthLogoutToolHandler handles logout requests
type AuthLogoutToolHandler struct {
	authManager *auth.OAuthFlowManager
}

// NewAuthLogoutToolHandler creates a new logout handler
func NewAuthLogoutToolHandler(authManager *auth.OAuthFlowManager) *AuthLogoutToolHandler {
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
