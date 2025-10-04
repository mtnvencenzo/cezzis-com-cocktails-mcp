// Package config provides application configuration management for the Cezzi Cocktails MCP server.
// It handles loading environment variables from .env files and provides a singleton pattern
// for accessing application settings throughout the application lifecycle.
//
// The package supports configuration for:
//   - Cocktails API connection settings (host and subscription key)
//   - Azure Entra External Id Tenant authentication settings (instance, domain, and user flow)
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

	// AzureCIAMInstance is the Azure Entra External Id Tenant instance URL.
	// Example: "https://your-tenant.b2clogin.com"
	AzureCIAMInstance string `env:"AZURE_CIAM_INSTANCE"`

	// AzureCIAMDomain is the Azure CIAM tenant domain name.
	// Example: "your-tenant.onmicrosoft.com"
	AzureCIAMDomain string `env:"AZURE_CIAM_DOMAIN"`

	// AzureCIAMUserFlow is the name of the Azure CIAM user flow for authentication.
	// Example: "sisu-p"
	AzureCIAMUserFlow string `env:"AZURE_CIAM_USERFLOW"`

	// AzureCIAMClientID is the client ID of the Azure External Id Tenant app registration.
	// Example: "84744194-da27-410f-ae0e-74f5589d4c96"
	AzureCIAMClientID string `env:"AZURE_CIAM_CLIENT_ID"`
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
	if instance.AzureCIAMInstance == "" {
		l.Logger.Warn().Msg("Warning: AZURE_CIAM_INSTANCE is not set")
	}
	if instance.AzureCIAMDomain == "" {
		l.Logger.Warn().Msg("Warning: AZURE_CIAM_DOMAIN is not set")
	}
	if instance.AzureCIAMUserFlow == "" {
		l.Logger.Warn().Msg("Warning: AZURE_CIAM_USERFLOW is not set")
	}
	if instance.AzureCIAMClientID == "" {
		l.Logger.Warn().Msg("Warning: AZURE_CIAM_CLIENT_ID is not set")
	}

	return instance
}

// GetAzureCIAMDiscoveryKeysURI constructs the Azure CIAM discovery keys URI
// by combining the instance, domain, and user flow settings.
// This URI is used to fetch the JSON Web Key Set (JWKS) for JWT token validation.
// Returns a formatted string like: "https://your-tenant.b2clogin.com/your-tenant.onmicrosoft.com/sisu-p/discovery/v2.0/keys"
func (a *AppSettings) GetAzureCIAMDiscoveryKeysURI() string {
	return fmt.Sprintf("%s/%s/%s/discovery/v2.0/keys", a.AzureCIAMInstance, a.AzureCIAMDomain, a.AzureCIAMUserFlow)
}
