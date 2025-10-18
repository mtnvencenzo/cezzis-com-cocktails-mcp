package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/middleware"
	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

var authLoginDescription = `
	This tool initiates the OAuth device login flow for accessing your Cezzis.com account
	and personalized features.

	Calling this tool will start the device code authentication process, which will return
	an authorization URL and code to enter in your web browser. After entering the code
	and signing in, you can return to this application and let us know you have completed
	the sign-in process.

	Once authenticated, the tool will store your access token securely for future requests
	that require authentication.
`

// AuthLoginTool handles OAuth authentication using device code flow
var AuthLoginTool = mcp.NewTool(
	"authentication_login_flow",
	mcp.WithDescription(authLoginDescription),
)

// AuthLoginToolHandler handles authentication login requests
type AuthLoginToolHandler struct {
	authManager *auth.OAuthFlowManager
}

// NewAuthLoginToolHandler creates a new authentication login handler
func NewAuthLoginToolHandler(authManager *auth.OAuthFlowManager) *AuthLoginToolHandler {
	return &AuthLoginToolHandler{
		authManager: authManager,
	}
}

// Handle handles authentication login requests
func (handler *AuthLoginToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sessionID := ctx.Value(middleware.McpSessionIDKey)
	if sessionID == nil || sessionID == "" {
		err := errors.New("missing required Mcp-Session-Id header")
		return mcp.NewToolResultError(err.Error()), err
	}

	// If HTTP mode, use device code flow and return instructions
	dc, err := handler.authManager.BeginDeviceAuth(ctx, sessionID.(string))
	if err != nil {
		telemetry.Logger.Error().Err(err).Msg("Failed to start device code auth")
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	msg := fmt.Sprintf(`Continue in your browser to sign in:

1) Visit: %s
2) Enter code: %s

This window can stay open; we'll finish automatically once you complete sign in.`, dc.VerificationURI, dc.UserCode)
	return mcp.NewToolResultText(msg), nil
}
