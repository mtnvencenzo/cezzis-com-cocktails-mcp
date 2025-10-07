package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

var authLoginDescription = `
	Initiates OAuth authentication flow for Cezzis.com account access.
	This tool starts an authorization code flow that opens your web browser 
	to authenticate with your Cezzis.com account and access personalized features 
	like favorites, ratings, and profile management.
	
	The tool will provide you with an authorization URL that will open in your browser.
	After completing the authentication, you can return to this application.
`

// AuthLoginTool handles OAuth authentication using device code flow
var AuthLoginTool = mcp.NewTool(
	"auth_login",
	mcp.WithDescription(authLoginDescription),
)

// AuthLoginToolHandler handles authentication login requests
type AuthLoginToolHandler struct {
	authManager *auth.Manager
}

// NewAuthLoginToolHandler creates a new authentication login handler
func NewAuthLoginToolHandler(authManager *auth.Manager) *AuthLoginToolHandler {
	return &AuthLoginToolHandler{
		authManager: authManager,
	}
}

// Handle handles authentication login requests
func (handler *AuthLoginToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// If HTTP mode, use device code flow and return instructions
	if handler.authManager.IsHTTPMode() {
		dc, err := handler.authManager.BeginDeviceAuth(ctx)
		if err != nil {
			l.Logger.Error().Err(err).Msg("Failed to start device code auth")
			return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
		}
		msg := fmt.Sprintf(`Continue in your browser to sign in:

1) Visit: %s
2) Enter code: %s

This window can stay open; we'll finish automatically once you complete sign in.`, dc.VerificationURI, dc.UserCode)
		return mcp.NewToolResultText(msg), nil
	}

	// stdio/local: run browser-based PKCE and return token details
	tokens, err := handler.authManager.Authenticate(ctx)
	if err != nil {
		l.Logger.Error().Err(err).Msg("Failed to complete authentication")
		return mcp.NewToolResultError(fmt.Sprintf("Authentication failed: %v", err)), nil
	}
	result := fmt.Sprintf(`✅ Authentication successful!

Token Details:
- Token Type: %s
- Expires In: %d seconds
- Scopes: %s

You are now authenticated and can use personalized features like:
- Rating cocktails
- Managing favorites
- Accessing your profile

Authentication will be automatically saved for future use.`,
		tokens.TokenType,
		tokens.ExpiresIn,
		tokens.Scope)
	return mcp.NewToolResultText(result), nil
}
