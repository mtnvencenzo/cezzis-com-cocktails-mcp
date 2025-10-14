// Package auth provides authentication related structures and functions for OAuth flows.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/logging"
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

// DeviceCodeResponse represents the device code response
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	Message         string `json:"message"`
}

// OAuthFlowManager handles OAuth authentication flows
type OAuthFlowManager struct {
	appSettings *config.AppSettings
	httpClient  *http.Client
	storage     *TokenStorage
}

// NewOAuthFlowManager creates a new OAuth flow manager
func NewOAuthFlowManager() *OAuthFlowManager {
	storage, err := NewTokenStorage()
	if err != nil {
		logging.Logger.Error().Err(err).Msg("Failed to create token storage")
	}

	manager := &OAuthFlowManager{
		appSettings: config.GetAppSettings(),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		storage:     storage,
	}

	return manager
}

// StartDeviceFlow initiates the device code authentication flow
func (auth *OAuthFlowManager) StartDeviceFlow(ctx context.Context) (*DeviceCodeResponse, error) {
	if strings.TrimSpace(auth.appSettings.Auth0Domain) == "" || strings.TrimSpace(auth.appSettings.Auth0ClientID) == "" {
		return nil, fmt.Errorf("Auth0 not configured: set AUTH0_DOMAIN and AUTH0_CLIENT_ID")
	}
	deviceEndpoint := fmt.Sprintf("https://%s/oauth/device/code", strings.TrimRight(auth.appSettings.Auth0Domain, "/"))

	data := url.Values{
		"client_id": {auth.appSettings.Auth0ClientID},
		"scope":     {auth.appSettings.Auth0Scopes},
	}
	audience := strings.TrimSpace(auth.appSettings.Auth0Audience)
	if audience != "" {
		data.Set("audience", audience)
	}

	logging.Logger.Info().
		Str("scopes_requested", auth.appSettings.Auth0Scopes).
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
			logging.Logger.Warn().Err(err).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read device code response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logging.Logger.Error().
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

	logging.Logger.Info().
		Str("user_code", deviceResp.UserCode).
		Str("verification_uri", deviceResp.VerificationURI).
		Msg("Device code flow started")

	return &deviceResp, nil
}

// PollForTokens polls for tokens after user completes device authentication
//
//nolint:gocyclo
func (auth *OAuthFlowManager) PollForTokens(ctx context.Context, deviceCode *DeviceCodeResponse, sessionID string) (*TokenResponse, error) {
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

		logging.Logger.Debug().
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
			logging.Logger.Warn().Err(err).Msg("Token polling request failed")
			time.Sleep(pollInterval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if closeErr := resp.Body.Close(); closeErr != nil {
			logging.Logger.Warn().Err(closeErr).Msg("Failed to close response body")
		}
		if err != nil {
			logging.Logger.Warn().Err(err).Msg("Failed to read token response")
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

			// Save tokens to storage
			if err := auth.storage.SaveToken(sessionID, &tokenResp); err != nil {
				logging.Logger.Warn().Err(err).Msg("Failed to save tokens to storage")
			}

			logging.Logger.Info().
				Str("scopes_granted", tokenResp.Scope).
				Msg("Successfully obtained access tokens")

			return &tokenResp, nil
		}

		// Check for authorization_pending (user hasn't completed auth yet)
		var errorResp map[string]interface{}
		if json.Unmarshal(body, &errorResp) == nil {
			if errorCode, ok := errorResp["error"].(string); ok {
				if errorCode == "authorization_pending" {
					logging.Logger.Debug().Msg("Authorization still pending, continuing to poll...")
					time.Sleep(pollInterval)
					continue
				}
				return nil, fmt.Errorf("authentication failed: %s", errorCode)
			}
		}

		logging.Logger.Warn().Str("response", string(body)).Msg("Unexpected token response")
		time.Sleep(pollInterval)
	}
}

// GetAccessToken returns the current access token, refreshing if necessary
func (auth *OAuthFlowManager) GetAccessToken(ctx context.Context, sessionID string) (string, error) {
	token, err := auth.storage.GetToken(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to load tokens from storage: %w", err)
	}

	if token == nil {
		return "", fmt.Errorf("no tokens available, authentication required")
	}

	if !token.ExpiresAt.IsZero() && time.Until(token.ExpiresAt) < 2*time.Minute {
		newToken, err := auth.refreshAccessToken(ctx, sessionID)
		if err != nil {
			logging.Logger.Warn().Err(err).Msg("Access token refresh failed; user may need to re-authenticate")

			if err := auth.storage.ClearTokens(sessionID); err != nil {
				logging.Logger.Warn().Err(err).Msg("Failed to clear tokens from storage")
			}

			return "", fmt.Errorf("failed to refresh access token: %w", err)
		}

		return newToken.AccessToken, nil
	}

	return "", nil
}

// refreshAccessToken refreshes tokens using the current refresh_token.
func (auth *OAuthFlowManager) refreshAccessToken(ctx context.Context, sessionID string) (*TokenResponse, error) {
	token, err := auth.storage.GetToken(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tokens from storage: %w", err)
	}

	if token == nil || token.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	if strings.TrimSpace(auth.appSettings.Auth0Domain) == "" || strings.TrimSpace(auth.appSettings.Auth0ClientID) == "" {
		return nil, fmt.Errorf("Auth0 not configured: set AUTH0_DOMAIN and AUTH0_CLIENT_ID")
	}

	tokenURL := fmt.Sprintf("https://%s/oauth/token", strings.TrimRight(auth.appSettings.Auth0Domain, "/"))

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {auth.appSettings.Auth0ClientID},
		"refresh_token": {token.RefreshToken},
		"scope":         {token.Scope},
	}
	audience := strings.TrimSpace(auth.appSettings.Auth0Audience)
	if audience != "" {
		data.Set("audience", audience)
	}

	logging.Logger.Info().
		Str("audience", audience).
		Bool("audience_included", audience != "").
		Str("scope", token.Scope).
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
			logging.Logger.Warn().Err(err).Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed (%d): %s", resp.StatusCode, string(body))
	}

	var newToken TokenResponse
	if err := json.Unmarshal(body, &newToken); err != nil {
		return nil, fmt.Errorf("failed to parse refreshed tokens: %w", err)
	}
	if newToken.ExpiresIn > 0 {
		newToken.ExpiresAt = time.Now().Add(time.Duration(newToken.ExpiresIn-60) * time.Second)
	}

	if auth.storage != nil {
		_ = auth.storage.SaveToken(sessionID, &newToken)
	}

	logging.Logger.Info().
		Str("scopes_granted", newToken.Scope).
		Msg("Successfully refreshed access tokens")

	return &newToken, nil
}

// IsAuthenticated returns true if the user is currently authenticated
func (auth *OAuthFlowManager) IsAuthenticated(sessionID string) bool {
	token, err := auth.storage.GetToken(sessionID)
	if err != nil || token == nil {
		return false
	}

	// Check if token is expired
	if time.Now().After(token.ExpiresAt) {
		logging.Logger.Warn().Msg("Stored tokens are expired")
		return false
	}

	return true
}

// Logout clears current authentication
func (auth *OAuthFlowManager) Logout(sessionID string) error {

	if err := auth.storage.ClearTokens(sessionID); err != nil {
		return fmt.Errorf("failed to clear stored tokens: %w", err)
	}

	logging.Logger.Info().Msg("Successfully logged out")
	return nil
}

// BeginDeviceAuth starts the device code flow and begins polling in the background.
// It returns the verification details for the user to complete in their browser.
func (auth *OAuthFlowManager) BeginDeviceAuth(ctx context.Context, sessionID string) (*DeviceCodeResponse, error) {
	deviceCode, err := auth.StartDeviceFlow(ctx)
	if err != nil {
		return nil, err
	}
	// Start polling in the background; tokens will be saved to storage on success
	go func() {
		// Use a generous timeout independent of the tool request context
		bgCtx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if _, err := auth.PollForTokens(bgCtx, deviceCode, sessionID); err != nil {
			logging.Logger.Warn().Err(err).Msg("Device code polling ended without tokens")
		}
	}()

	return deviceCode, nil
}
