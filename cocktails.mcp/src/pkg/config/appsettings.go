// Package config provides application configuration management for the Cezzi Cocktails MCP server.
// It handles loading environment variables from .env files and provides a singleton pattern
// for accessing application settings throughout the application lifecycle.
//
// The package supports configuration for:
//   - Cocktails API connection settings (host and subscription key)
//   - Azure AD B2C authentication settings (instance, domain, and user flow)
//
// Configuration is loaded from environment variables and .env files located in the
// executable directory. The package uses a thread-safe singleton pattern to ensure
// configuration is loaded only once and shared across all goroutines.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
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

	// AzureAdB2CInstance is the Azure AD B2C tenant instance URL.
	// Example: "https://your-tenant.b2clogin.com"
	AzureAdB2CInstance string `env:"AZUREAD_B2C_INSTANCE"`

	// AzureAdB2CDomain is the Azure AD B2C tenant domain name.
	// Example: "your-tenant.onmicrosoft.com"
	AzureAdB2CDomain string `env:"AZUREAD_B2C_DOMAIN"`

	// AzureAdB2CUserFlow is the name of the Azure AD B2C user flow for authentication.
	// Example: "B2C_1_signupsignin"
	AzureAdB2CUserFlow string `env:"AZUREAD_B2C_USERFLOW"`
}

// GetAzureAdB2CDiscoveryKeysURI constructs the Azure AD B2C discovery keys URI
// by combining the instance, domain, and user flow settings.
// This URI is used to fetch the JSON Web Key Set (JWKS) for JWT token validation.
// Returns a formatted string like: "https://your-tenant.b2clogin.com/your-tenant.onmicrosoft.com/B2C_1_signupsignin/discovery/v2.0/keys"
func (a *AppSettings) GetAzureAdB2CDiscoveryKeysURI() string {
	return fmt.Sprintf("%s/%s/%s/discovery/v2.0/keys", a.AzureAdB2CInstance, a.AzureAdB2CDomain, a.AzureAdB2CUserFlow)
}

var (
	instance *AppSettings
	once     sync.Once
)

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
	once.Do(func() {
		exePath, oserr := os.Executable()
		if oserr != nil {
			fmt.Printf("Server error - exe path: %v\n", oserr)
		}

		// Get the directory of the executable
		exeDir := filepath.Dir(exePath)

		_ = godotenv.Overload(
			fmt.Sprintf("%s/%s", exeDir, ".env"),
			fmt.Sprintf("%s/%s", exeDir, ".env.local"))

		instance = &AppSettings{}
		if err := env.Parse(instance); err != nil {
			fmt.Printf("Exe Path: %v\n", exeDir)
			fmt.Printf("Failed to parse app settings: %v\n", err)

			if instance.CocktailsAPIHost == "" {
				fmt.Printf("Warning: COCKTAILS_API_HOST is not set\n")
			}
			if instance.CocktailsAPISubscriptionKey == "" {
				fmt.Printf("Warning: COCKTAILS_API_XKEY is not set\n")
			}
		}
	})

	return instance
}
