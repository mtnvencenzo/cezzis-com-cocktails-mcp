// Package auth provides OAuth authentication functionality for the MCP server.
// It implements device code flow for Azure AD B2C authentication.
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

// AuthManager handles OAuth authentication flows
type AuthManager struct {
	appSettings   *config.AppSettings
	currentTokens *TokenResponse
	httpClient    *http.Client
	storage       *TokenStorage
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
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

	manager := &AuthManager{
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
func (auth *AuthManager) StartDeviceFlow(ctx context.Context) (*DeviceCodeResponse, error) {
	// Azure AD B2C device code endpoint
	deviceEndpoint := fmt.Sprintf("%s/%s/%s/oauth2/v2.0/devicecode",
		auth.appSettings.AzureAdB2CInstance,
		auth.appSettings.AzureAdB2CDomain,
		auth.appSettings.AzureAdB2CUserFlow)

	data := url.Values{
		"client_id": {"84744194-da27-410f-ae0e-74f5589d4c96"}, // From your OpenAPI spec
		"scope":     {"https://cezzis.onmicrosoft.com/cocktailsapi/Account.Read https://cezzis.onmicrosoft.com/cocktailsapi/Account.Write"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", deviceEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create device code request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := auth.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read device code response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed: %s", string(body))
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
func (auth *AuthManager) PollForTokens(ctx context.Context, deviceCode *DeviceCodeResponse) (*TokenResponse, error) {
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
			"client_id":   {"84744194-da27-410f-ae0e-74f5589d4c96"},
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
		resp.Body.Close()
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
func (auth *AuthManager) GetAccessToken(ctx context.Context) (string, error) {
	if auth.currentTokens == nil {
		return "", fmt.Errorf("no tokens available, authentication required")
	}

	// TODO: Implement token refresh logic here if needed
	// For now, just return the current token
	return auth.currentTokens.AccessToken, nil
}

// IsAuthenticated returns true if the user is currently authenticated
func (auth *AuthManager) IsAuthenticated() bool {
	return auth.currentTokens != nil && auth.currentTokens.AccessToken != ""
}

// Logout clears current authentication
func (auth *AuthManager) Logout() error {
	auth.currentTokens = nil

	if auth.storage != nil {
		if err := auth.storage.ClearTokens(); err != nil {
			return fmt.Errorf("failed to clear stored tokens: %w", err)
		}
	}

	l.Logger.Info().Msg("Successfully logged out")
	return nil
}
