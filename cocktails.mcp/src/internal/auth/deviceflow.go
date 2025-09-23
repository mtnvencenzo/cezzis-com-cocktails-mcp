// Package auth provides OAuth authentication functionality for the MCP server.
// It implements device code flow for Azure AD B2C authentication.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cezzis.com/cezzis-mcp-server/internal/config"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	// Computed locally (not returned by the provider)
	ExpiresAt time.Time `json:"-"`
}

// PKCEChallenge represents PKCE challenge data for authorization code flow
type PKCEChallenge struct {
	CodeVerifier  string
	CodeChallenge string
}

// DeviceCodeResponse represents the device code response
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

// Manager handles OAuth authentication flows
type Manager struct {
	appSettings        *config.AppSettings
	currentTokens      *TokenResponse
	currentPKCE        *PKCEChallenge
	currentRedirectURI string
	httpClient         *http.Client
	storage            *TokenStorage
}

// NewManager creates a new authentication manager
func NewManager() *Manager {
	// Create storage in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		l.Logger.Warn().Err(err).Msg("Failed to get home directory, using temp storage")
		homeDir = "/tmp"
	}

	storage, err := NewTokenStorage(filepath.Join(homeDir, ".cezzis"))
	if err != nil {
		l.Logger.Error().Err(err).Msg("Failed to create token storage")
	}

	manager := &Manager{
		appSettings: config.GetAppSettings(),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		storage:     storage,
	}

	// Try to load existing tokens
	if storage != nil {
		if tokens, err := storage.LoadTokens(); err == nil && tokens != nil {
			manager.currentTokens = tokens
			l.Logger.Info().Msg("Loaded existing authentication tokens")
		}
	}

	return manager
}

// PollForTokens polls for tokens after user completes device authentication
//
//nolint:gocyclo
func (auth *Manager) PollForTokens(ctx context.Context, deviceCode *DeviceCodeResponse) (*TokenResponse, error) {
	tokenEndpoint := fmt.Sprintf("%s/%s/%s/oauth2/v2.0/token",
		auth.appSettings.AzureAdB2CInstance,
		auth.appSettings.AzureAdB2CDomain,
		auth.appSettings.AzureAdB2CUserFlow)

	pollInterval := time.Duration(deviceCode.Interval) * time.Second
	if pollInterval < 5*time.Second {
		pollInterval = 5 * time.Second // Minimum polling interval
	}

	expiresAt := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if time.Now().After(expiresAt) {
			return nil, fmt.Errorf("device code expired")
		}

		data := url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"client_id":   {auth.appSettings.AzureAdB2CClientID},
			"device_code": {deviceCode.DeviceCode},
		}

		req, err := http.NewRequestWithContext(ctx, "POST", tokenEndpoint, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create token request: %w", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := auth.httpClient.Do(req)
		if err != nil {
			l.Logger.Warn().Err(err).Msg("Token polling request failed")
			time.Sleep(pollInterval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			l.Logger.Warn().Err(closeErr).Msg("Failed to close response body")
		}
		if err != nil {
			l.Logger.Warn().Err(err).Msg("Failed to read token response")
			time.Sleep(pollInterval)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var tokenResp TokenResponse
			if err := json.Unmarshal(body, &tokenResp); err != nil {
				return nil, fmt.Errorf("failed to parse token response: %w", err)
			}

			// Compute local expiry with a safety margin
			if tokenResp.ExpiresIn > 0 {
				tokenResp.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
			}

			auth.currentTokens = &tokenResp

			// Save tokens to storage
			if auth.storage != nil {
				if err := auth.storage.SaveTokens(&tokenResp); err != nil {
					l.Logger.Warn().Err(err).Msg("Failed to save tokens to storage")
				}
			}

			l.Logger.Info().Msg("Successfully obtained access tokens")
			return &tokenResp, nil
		}

		// Check for authorization_pending (user hasn't completed auth yet)
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorCode, ok := errorResp["error"].(string); ok {
				if errorCode == "authorization_pending" {
					l.Logger.Debug().Msg("Authorization still pending, continuing to poll...")
					time.Sleep(pollInterval)
					continue
				}
				return nil, fmt.Errorf("authentication failed: %s", errorCode)
			}
		}

		l.Logger.Warn().Str("response", string(body)).Msg("Unexpected token response")
		time.Sleep(pollInterval)
	}
}

// GetAccessToken returns the current access token, refreshing if necessary
func (auth *Manager) GetAccessToken(ctx context.Context) (string, error) {
	if auth.currentTokens == nil {
		return "", fmt.Errorf("no tokens available, authentication required")
	}

	if !auth.currentTokens.ExpiresAt.IsZero() && time.Until(auth.currentTokens.ExpiresAt) < 2*time.Minute {
		if _, err := auth.refreshAccessToken(ctx); err != nil {
			l.Logger.Warn().Err(err).Msg("Access token refresh failed; user may need to re-authenticate")
		}
	}

	return auth.currentTokens.AccessToken, nil
}

// refreshAccessToken refreshes tokens using the current refresh_token.
func (auth *Manager) refreshAccessToken(ctx context.Context) (*TokenResponse, error) {
	if auth.currentTokens == nil || auth.currentTokens.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}
	tokenURL := fmt.Sprintf("%s/%s/%s/oauth2/v2.0/token",
		auth.appSettings.AzureAdB2CInstance,
		auth.appSettings.AzureAdB2CDomain,
		auth.appSettings.AzureAdB2CUserFlow)

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {auth.appSettings.AzureAdB2CClientID},
		"refresh_token": {auth.currentTokens.RefreshToken},
		// Let the server infer scopes; if needed, reuse previously granted scopes:
		// "scope": {auth.currentTokens.Scope},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := auth.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Logger.Warn().Err(err).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse refreshed tokens: %w", err)
	}
	if tokens.ExpiresIn > 0 {
		tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn-60) * time.Second)
	}
	auth.currentTokens = &tokens
	if auth.storage != nil {
		_ = auth.storage.SaveTokens(&tokens)
	}
	return &tokens, nil
}

// StartBrowserAuth initiates browser-based OAuth authentication using Authorization Code Flow with PKCE
//
//nolint:gocyclo
func (auth *Manager) StartBrowserAuth(ctx context.Context) (*TokenResponse, error) {
	// Generate PKCE challenge
	pkce, err := generatePKCEChallenge()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}
	auth.currentPKCE = pkce

	// Use fixed port 6097 for Azure AD B2C configuration
	port := 6097

	// Test if port is available
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, fmt.Errorf("port %d is not available: %w", port, err)
	}
	if err := listener.Close(); err != nil {
		l.Logger.Warn().Err(err).Msg("Failed to close port listener")
	}

	redirectURI := fmt.Sprintf("http://localhost:%d/callback", port)
	auth.currentRedirectURI = redirectURI

	// Generate state parameter for CSRF protection
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		return nil, fmt.Errorf("failed to generate state parameter: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(stateBytes)

	// Build authorization URL
	authURL := fmt.Sprintf("%s/%s/%s/oauth2/v2.0/authorize",
		auth.appSettings.AzureAdB2CInstance,
		auth.appSettings.AzureAdB2CDomain,
		auth.appSettings.AzureAdB2CUserFlow)

	params := url.Values{
		"client_id":             {auth.appSettings.AzureAdB2CClientID},
		"response_type":         {"code"},
		"redirect_uri":          {redirectURI},
		"scope":                 {"openid offline_access https://cezzis.onmicrosoft.com/cocktailsapi/Account.Read https://cezzis.onmicrosoft.com/cocktailsapi/Account.Write openid"},
		"state":                 {state},
		"code_challenge":        {pkce.CodeChallenge},
		"code_challenge_method": {"S256"},
	}

	fullAuthURL := authURL + "?" + params.Encode()

	// Channel to receive callback result
	resultChan := make(chan CallbackResult, 1)

	// Setup callback server
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		returnedState := r.URL.Query().Get("state")
		errorParam := r.URL.Query().Get("error")

		result := CallbackResult{
			Code:  code,
			State: returnedState,
			Error: errorParam,
		}

		if errorParam != "" {
			http.Error(w, fmt.Sprintf("Authorization failed: %s", errorParam), http.StatusBadRequest)
		} else if code == "" {
			http.Error(w, "No authorization code received", http.StatusBadRequest)
			result.Error = "no_code"
		} else if returnedState != state {
			http.Error(w, "Invalid state parameter", http.StatusBadRequest)
			result.Error = "invalid_state"
		} else {
			// Success page
			w.Header().Set("Content-Type", "text/html")
			if _, err := fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>Authorization Complete</title>
<style>body{font-family:Arial,sans-serif;text-align:center;margin:50px;}
.success{color:#28a745;}</style></head>
<body><div class="success">
<h2>✅ Authorization Successful!</h2>
<p>You can now close this window and return to your application.</p>
</div></body></html>`); err != nil {
				l.Logger.Warn().Err(err).Msg("Failed to write success page")
			}
		}

		// Send result to waiting goroutine
		select {
		case resultChan <- result:
		default:
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Logger.Error().Err(err).Msg("Callback server failed")
		}
	}()

	// Try to open browser
	l.Logger.Info().
		Str("auth_url", fullAuthURL).
		Int("port", port).
		Msg("Opening browser for authentication")

	if err := openBrowser(fullAuthURL); err != nil {
		l.Logger.Warn().Err(err).Msg("Failed to open browser automatically")
	}

	// Wait for callback with timeout
	authCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	var result CallbackResult
	select {
	case result = <-resultChan:
	case <-authCtx.Done():
		if err := server.Shutdown(context.Background()); err != nil {
			l.Logger.Warn().Err(err).Msg("Failed to shutdown server")
		}
		return nil, fmt.Errorf("authentication timed out")
	}

	// Shutdown server
	if err := server.Shutdown(context.Background()); err != nil {
		l.Logger.Warn().Err(err).Msg("Failed to shutdown server")
	}

	if result.Error != "" {
		return nil, fmt.Errorf("authentication failed: %s", result.Error)
	}

	// Exchange code for tokens
	return auth.exchangeCodeForTokens(ctx, result.Code)
}

// CallbackResult represents the OAuth callback result
type CallbackResult struct {
	Code  string
	State string
	Error string
}

// generatePKCEChallenge generates PKCE code verifier and challenge
func generatePKCEChallenge() (*PKCEChallenge, error) {
	verifierBytes := make([]byte, 96)
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, err
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEChallenge{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
	}, nil
}

// exchangeCodeForTokens exchanges authorization code for tokens
func (auth *Manager) exchangeCodeForTokens(ctx context.Context, code string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/%s/%s/oauth2/v2.0/token",
		auth.appSettings.AzureAdB2CInstance,
		auth.appSettings.AzureAdB2CDomain,
		auth.appSettings.AzureAdB2CUserFlow)

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {auth.appSettings.AzureAdB2CClientID},
		"code":          {code},
		"redirect_uri":  {auth.currentRedirectURI},
		"code_verifier": {auth.currentPKCE.CodeVerifier},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := auth.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Logger.Warn().Err(err).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		l.Logger.Error().
			Int("status", resp.StatusCode).
			Str("response", string(body)).
			Msg("Token exchange failed")
		return nil, fmt.Errorf("token exchange failed (%d): %s", resp.StatusCode, string(body))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse tokens: %w", err)
	}

	// Save tokens
	auth.currentTokens = &tokens
	if auth.storage != nil {
		if err := auth.storage.SaveTokens(&tokens); err != nil {
			l.Logger.Warn().Err(err).Msg("Failed to save tokens")
		}
	}

	l.Logger.Info().Msg("Successfully obtained tokens via browser authentication")
	return &tokens, nil
}

// openBrowser opens URL in default browser
func openBrowser(targetURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", targetURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	case "darwin":
		cmd = exec.Command("open", targetURL)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// IsAuthenticated returns true if the user is currently authenticated
func (auth *Manager) IsAuthenticated() bool {
	return auth.currentTokens != nil && auth.currentTokens.AccessToken != ""
}

// Logout clears current authentication
func (auth *Manager) Logout() error {
	auth.currentTokens = nil

	if auth.storage != nil {
		if err := auth.storage.ClearTokens(); err != nil {
			return fmt.Errorf("failed to clear stored tokens: %w", err)
		}
	}

	l.Logger.Info().Msg("Successfully logged out")
	return nil
}
