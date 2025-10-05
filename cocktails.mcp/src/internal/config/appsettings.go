// Package config provides application configuration management for the Cezzi Cocktails MCP server.
// It handles loading environment variables from .env files and provides a singleton pattern
// for accessing application settings throughout the application lifecycle.
//
// The package supports configuration for:
//   - Cocktails API connection settings (host and subscription key)
//   - Auth0 authentication settings (domain, client id, audience, scopes)
//
// Configuration is loaded from environment variables and .env files located in the
// executable directory. The package uses a thread-safe singleton pattern to ensure
// configuration is loaded only once and shared across all goroutines.
package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"

	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

// AppSettings contains all application configuration settings loaded from environment variables.
// It provides a centralized way to access configuration values throughout the application.
type AppSettings struct {
	// CocktailsApiHost is the base URL for the Cocktails API service.
	// Example: "https://api.cocktails.com"
	CocktailsAPIHost string `env:"COCKTAILS_API_HOST"`

	// CocktailsApiSubscriptionKey is the subscription key required for authenticating
	// requests to the Cocktails API service.
	CocktailsAPISubscriptionKey string `env:"COCKTAILS_API_XKEY"`

	// Auth0Domain is your Auth0 domain or custom domain.
	// Example: "your-tenant.us.auth0.com" or "login.cezzis.com"
	Auth0Domain string `env:"AUTH0_DOMAIN"`

	// Auth0ClientID is the public client identifier for your Auth0 Application (Native/Public).
	Auth0ClientID string `env:"AUTH0_CLIENT_ID"`

	// Auth0Audience is the API Identifier to request an access token for.
	// Example: "https://api.cezzis.com/cocktails"
	Auth0Audience string `env:"AUTH0_AUDIENCE"`

	// Auth0Scopes is the list of scopes to request.
	// Example: "openid profile email offline_access cocktails:read cocktails:write"
	Auth0Scopes string `env:"AUTH0_SCOPES"`
}

// GetAppSettings returns a singleton instance of AppSettings loaded from environment variables.
// This function is thread-safe and ensures that configuration is loaded only once,
// even when called concurrently from multiple goroutines.
//
// The function performs the following operations:
//   - Loads environment variables from .env and .env.local files in the executable directory
//   - Parses the configuration into the AppSettings struct
//   - Logs warnings for missing required configuration values
//   - Returns the same instance on subsequent calls
//
// Returns a pointer to the AppSettings instance containing all application configuration.
func GetAppSettings() *AppSettings {
	instance := &AppSettings{}
	if err := env.Parse(instance); err != nil {
		l.Logger.Warn().Err(err).Msg("Failed to parse app settings")
	}

	if instance.CocktailsAPIHost == "" {
		l.Logger.Warn().Msg("Warning: COCKTAILS_API_HOST is not set")
	}
	if instance.CocktailsAPISubscriptionKey == "" {
		l.Logger.Warn().Msg("Warning: COCKTAILS_API_XKEY is not set")
	}
	// Warn if Auth0 is not configured
	if instance.Auth0Domain == "" {
		l.Logger.Warn().Msg("Warning: AUTH0_DOMAIN is not set; authentication will fail")
	}
	if instance.Auth0Domain == "" {
		l.Logger.Debug().Msg("Note: AUTH0_DOMAIN is not set (Auth0 disabled)")
	}
	if instance.Auth0ClientID == "" {
		l.Logger.Debug().Msg("Note: AUTH0_CLIENT_ID is not set (Auth0 disabled)")
	}

	return instance
}

// GetAuth0JWKSURI returns the JWKS URL for Auth0.
// Example: https://YOUR_DOMAIN/.well-known/jwks.json
func (a *AppSettings) GetAuth0JWKSURI() string {
	if a.Auth0Domain == "" {
		return ""
	}
	return fmt.Sprintf("https://%s/.well-known/jwks.json", a.Auth0Domain)
}
