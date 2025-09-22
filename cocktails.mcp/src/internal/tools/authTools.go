package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

var authLoginDescription = `
	Initiates OAuth authentication flow for Cezzis.com account access.
	This tool starts a device code authentication flow that allows you to authenticate 
	with your Cezzis.com account to access personalized features like favorites, ratings,
	and profile management.
	
	The tool will provide you with a user code and verification URL. Open the URL in your 
	browser and enter the provided code to complete authentication.
`

// AuthLoginTool handles OAuth authentication using device code flow
var AuthLoginTool = mcp.NewTool(
	"auth_login",
	mcp.WithDescription(authLoginDescription),
)

// AuthLoginToolHandler handles authentication login requests
type AuthLoginToolHandler struct {
	authManager *auth.AuthManager
}

// NewAuthLoginToolHandler creates a new authentication login handler
func NewAuthLoginToolHandler(authManager *auth.AuthManager) *AuthLoginToolHandler {
	return &AuthLoginToolHandler{
		authManager: authManager,
	}
}

// Handle handles authentication login requests
func (handler *AuthLoginToolHandler) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Start device code flow
	deviceCode, err := handler.authManager.StartDeviceFlow(ctx)
	if err != nil {
		l.Logger.Error().Err(err).Msg("Failed to start device flow")
		return mcp.NewToolResultError(fmt.Sprintf("Failed to start authentication: %v", err)), nil
	}

	// Format the response for the user
	result := fmt.Sprintf(`Authentication started successfully!

Please follow these steps to complete authentication:

1. Open your web browser and navigate to: %s
2. Enter the following user code: %s
3. Complete the sign-in process with your Cezzis.com account

The authentication will be active for %d minutes. The system will automatically check for completion.

Waiting for you to complete authentication...`,
		deviceCode.VerificationURI,
		deviceCode.UserCode,
		deviceCode.ExpiresIn/60)

	// Start polling for tokens in background
	go func() {
		pollCtx, cancel := context.WithTimeout(context.Background(), time.Duration(deviceCode.ExpiresIn)*time.Second)
		defer cancel()

		tokens, err := handler.authManager.PollForTokens(pollCtx, deviceCode)
		if err != nil {
			l.Logger.Error().Err(err).Msg("Failed to obtain tokens")
			return
		}

		l.Logger.Info().
			Str("token_type", tokens.TokenType).
			Int("expires_in", tokens.ExpiresIn).
			Msg("Authentication completed successfully")
	}()

	return mcp.NewToolResultText(result), nil
}

var authStatusDescription = `
	Check the current authentication status for Cezzis.com account access.
	This tool returns whether you are currently authenticated and can access 
	personalized features like favorites, ratings, and profile management.
`

// AuthStatusTool checks current authentication status
var AuthStatusTool = mcp.NewTool(
	"auth_status",
	mcp.WithDescription(authStatusDescription),
)

// AuthStatusToolHandler handles authentication status requests
type AuthStatusToolHandler struct {
	authManager *auth.AuthManager
}

// NewAuthStatusToolHandler creates a new authentication status handler
func NewAuthStatusToolHandler(authManager *auth.AuthManager) *AuthStatusToolHandler {
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
