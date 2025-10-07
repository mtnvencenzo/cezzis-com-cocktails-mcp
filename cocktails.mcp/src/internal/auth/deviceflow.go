// Package auth provides OAuth authentication functionality for the MCP server.
// It implements Authorization Code with PKCE and Device Code flows for Auth0.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	appSettings   *config.AppSettings
	currentTokens *TokenResponse
	httpClient    *http.Client
	storage       *TokenStorage
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

// StartDeviceFlow initiates the device code authentication flow
func (auth *Manager) StartDeviceFlow(ctx context.Context) (*DeviceCodeResponse, error) {
	if strings.TrimSpace(auth.appSettings.Auth0Domain) == "" || strings.TrimSpace(auth.appSettings.Auth0ClientID) == "" {
		return nil, fmt.Errorf("Auth0 not configured: set AUTH0_DOMAIN and AUTH0_CLIENT_ID")
	}
	deviceEndpoint := fmt.Sprintf("https://%s/oauth/device/code", strings.TrimRight(auth.appSettings.Auth0Domain, "/"))

	requestedScopes := firstNonEmpty(auth.appSettings.Auth0Scopes, config.DefaultAuth0Scopes)
	data := url.Values{
		"client_id": {auth.appSettings.Auth0ClientID},
		"scope":     {requestedScopes},
	}
	audience := strings.TrimSpace(auth.appSettings.Auth0Audience)
	if audience != "" {
		data.Set("audience", audience)
	}

	l.Logger.Info().
		Str("scopes_requested", requestedScopes).
		Str("audience", audience).
		Bool("audience_included", audience != "").
		Str("request_params", data.Encode()).
		Msg("Starting device flow authentication")

	req, err := http.NewRequestWithContext(ctx, "POST", deviceEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create device code request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := auth.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			l.Logger.Warn().Err(err).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read device code response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		l.Logger.Error().
			Int("status_code", resp.StatusCode).
			Str("response_body", string(body)).
			Str("device_endpoint", deviceEndpoint).
			Msg("Device code request failed")

		return nil, fmt.Errorf("device code request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var deviceResp DeviceCodeResponse
	if err := json.Unmarshal(body, &deviceResp); err != nil {
		return nil, fmt.Errorf("failed to parse device code response: %w", err)
	}

	l.Logger.Info().
		Str("user_code", deviceResp.UserCode).
		Str("verification_uri", deviceResp.VerificationURI).
		Msg("Device code flow started")

	return &deviceResp, nil
}

// PollForTokens polls for tokens after user completes device authentication
//
//nolint:gocyclo
func (auth *Manager) PollForTokens(ctx context.Context, deviceCode *DeviceCodeResponse) (*TokenResponse, error) {
	tokenEndpoint := fmt.Sprintf("https://%s/oauth/token", strings.TrimRight(auth.appSettings.Auth0Domain, "/"))

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
			"client_id":   {auth.appSettings.Auth0ClientID},
			"device_code": {deviceCode.DeviceCode},
		}
		audience := strings.TrimSpace(auth.appSettings.Auth0Audience)
		if audience != "" {
			data.Set("audience", audience)
		}

		l.Logger.Debug().
			Str("audience", audience).
			Bool("audience_included", audience != "").
			Msg("Polling for tokens")

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

			l.Logger.Info().
				Str("scopes_granted", tokenResp.Scope).
				Msg("Successfully obtained access tokens")
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

// Authenticate initiates the appropriate authentication flow based on the environment
func (auth *Manager) Authenticate(ctx context.Context) (*TokenResponse, error) {
	l.Logger.Info().Msg("Container environment detected, using device code flow")
	return auth.authenticateDeviceCode(ctx)
}

// authenticateDeviceCode handles device code flow for container environments
func (auth *Manager) authenticateDeviceCode(ctx context.Context) (*TokenResponse, error) {
	// Start device code flow
	deviceCode, err := auth.StartDeviceFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start device code flow: %w", err)
	}

	// Return instructions to the user instead of trying to open browser
	l.Logger.Info().
		Str("user_code", deviceCode.UserCode).
		Str("verification_uri", deviceCode.VerificationURI).
		Str("message", deviceCode.Message).
		Msg("Device authentication required")

	// For HTTP mode, you might want to return this information to the client
	// instead of polling automatically
	return auth.PollForTokens(ctx, deviceCode)
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
	if strings.TrimSpace(auth.appSettings.Auth0Domain) == "" || strings.TrimSpace(auth.appSettings.Auth0ClientID) == "" {
		return nil, fmt.Errorf("Auth0 not configured: set AUTH0_DOMAIN and AUTH0_CLIENT_ID")
	}
	tokenURL := fmt.Sprintf("https://%s/oauth/token", strings.TrimRight(auth.appSettings.Auth0Domain, "/"))

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {auth.appSettings.Auth0ClientID},
		"refresh_token": {auth.currentTokens.RefreshToken},
		"scope":         {auth.currentTokens.Scope},
	}
	audience := strings.TrimSpace(auth.appSettings.Auth0Audience)
	if audience != "" {
		data.Set("audience", audience)
	}

	l.Logger.Info().
		Str("audience", audience).
		Bool("audience_included", audience != "").
		Str("scope", auth.currentTokens.Scope).
		Msg("Refreshing access token")

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

	l.Logger.Info().
		Str("scopes_granted", tokens.Scope).
		Msg("Successfully refreshed access tokens")

	return &tokens, nil
}

// firstNonEmpty returns the first non-empty string
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
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

// BeginDeviceAuth starts the device code flow and begins polling in the background.
// It returns the verification details for the user to complete in their browser.
func (auth *Manager) BeginDeviceAuth(ctx context.Context) (*DeviceCodeResponse, error) {
	deviceCode, err := auth.StartDeviceFlow(ctx)
	if err != nil {
		return nil, err
	}
	// Start polling in the background; tokens will be saved to storage on success
	go func() {
		// Use a generous timeout independent of the tool request context
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()
		if _, err := auth.PollForTokens(bgCtx, deviceCode); err != nil {
			l.Logger.Warn().Err(err).Msg("Device code polling ended without tokens")
		}
	}()
	return deviceCode, nil
}
