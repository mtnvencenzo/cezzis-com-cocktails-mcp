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
	"net/url"

	"github.com/caarlos0/env/v11"
)

// DefaultAuth0Scopes defines the default OAuth2 scopes to request when no AUTH0_SCOPES
// environment variable is configured. These scopes provide basic authentication and
// account access capabilities.
const DefaultAuth0Scopes = "openid offline_access profile email read:owned-account write:owned-account"

// AppSettings contains all application configuration settings loaded from environment variables.
// It provides a centralized way to access configuration values throughout the application.
type AppSettings struct {
	// Port is the port number on which the server will listen for HTTP requests.
	// Default is 7999 if not set.
	// Example: "7999"
	Port int `env:"PORT" envDefault:"7999"`

	// CocktailsApiHost is the base URL for the Cocktails API service.
	// Example: "https://api.cocktails.com"
	CocktailsAPIHost string `env:"COCKTAILS_API_HOST"`

	// CocktailsApiSubscriptionKey is the subscription key required for authenticating
	// requests to the Cocktails API service.
	CocktailsAPISubscriptionKey string `env:"COCKTAILS_API_XKEY"`

	// AccountsApiHost is the base URL for the Accounts API service.
	// Example: "https://api.accounts.com"
	AccountsAPIHost string `env:"ACCOUNTS_API_HOST"`

	// AccountsApiSubscriptionKey is the subscription key required for authenticating
	// requests to the Accounts API service.
	AccountsAPISubscriptionKey string `env:"ACCOUNTS_API_XKEY"`

	// AISearchAPIHost is the base URL for the AI Search API service.
	// Example: "https://api.aisearch.com"
	AISearchAPIHost string `env:"AISEARCH_API_HOST"`

	// AISearchAPISubscriptionKey is the subscription key required for authenticating
	// requests to the AI Search API service.
	AISearchAPISubscriptionKey string `env:"AISEARCH_API_XKEY"`

	// Auth0Domain is your Auth0 domain or custom domain.
	// Example: "your-tenant.us.auth0.com" or "login.cezzis.com"
	Auth0Domain string `env:"AUTH0_DOMAIN"`

	// Auth0NativeClientID is the public client identifier for your Auth0 Application (Native/Public).
	Auth0NativeClientID string `env:"AUTH0_NATIVE_CLIENT_ID"`

	// Auth0Audience is the API Identifier to request an access token for.
	// Example: "https://api.cezzis.com/cocktails"
	Auth0AccountsAPIAudience string `env:"AUTH0_ACCOUNTS_API_AUDIENCE"`

	// Auth0Scopes is the list of scopes to request.
	// Example: "openid profile email offline_access cocktails:read cocktails:write"
	Auth0Scopes string `env:"AUTH0_SCOPES"`

	// PostgresHost is the hostname for the PostgreSQL server.
	PostgresHost string `env:"POSTGRES_HOST" envDefault:"localhost"`

	// PostgresPort is the port number for the PostgreSQL server.
	PostgresPort int `env:"POSTGRES_PORT" envDefault:"5432"`

	// PostgresDBName is the name of the PostgreSQL database.
	PostgresDBName string `env:"POSTGRES_DB"`

	// PostgresUser is the username for the PostgreSQL connection.
	PostgresUser string `env:"POSTGRES_USER"`

	// PostgresPassword is the password for the PostgreSQL connection.
	PostgresPassword string `env:"POSTGRES_PASSWORD"`

	// PostgresUseTLS controls whether PostgreSQL connections use TLS.
	PostgresUseTLS bool `env:"POSTGRES_USE_TLS" envDefault:"false"`

	// InitJobEnabled controls whether the background init job runs at startup.
	InitJobEnabled bool `env:"INIT_JOB_ENABLED" envDefault:"true"`

	// InitDelaySeconds is the delay in seconds before the background init job runs.
	InitDelaySeconds int `env:"INIT_DELAY_SECONDS" envDefault:"30"`

	// OTLPEndpoint is the OTLP collector endpoint to send telemetry data to.
	// Example: "localhost:4317"
	OTLPEndpoint string `env:"OTLP_ENDPOINT" envDefault:"localhost:4317"`

	// OTLPHeaders are the headers to include in OTLP requests, formatted as key=value pairs separated by commas.
	// Example: "api-key=your_api_key,env=production"
	OTLPHeaders string `env:"OTLP_HEADERS" envDefault:""`

	// OTLPInsecure indicates whether to use an insecure connection (no TLS) for OTLP communication.
	// Default is false (use secure connection).
	OTLPInsecure bool `env:"OTLP_INSECURE" envDefault:"false"`

	// OTLPLogEnabled indicates whether to send logs to the OTLP collector.
	// Default is true.
	OTLPLogEnabled bool `env:"OTLP_LOG_ENABLED" envDefault:"true"`

	// OTLPTraceEnabled indicates whether to send traces to the OTLP collector.
	// Default is true.
	OTLPTraceEnabled bool `env:"OTLP_TRACE_ENABLED" envDefault:"true"`

	// OTLPMetricsEnabled indicates whether to send metrics to the OTLP collector.
	// Default is true.
	OTLPMetricsEnabled bool `env:"OTLP_METRICS_ENABLED" envDefault:"true"`

	// CezzisBaseURL is the base URL for Cezzis.com website.
	// Example: "https://www.cezzis.com"
	CezzisBaseURL string `env:"CEZZIS_BASE_URL" envDefault:""`
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
		panic(err)
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

// PostgresConnString assembles a PostgreSQL connection string from decomposed settings.
func (a *AppSettings) PostgresConnString() string {
	return a.postgresConnString(a.PostgresDBName)
}

// PostgresAdminConnString returns a connection string targeting the default 'postgres' database.
func (a *AppSettings) PostgresAdminConnString() string {
	return a.postgresConnString("postgres")
}

func (a *AppSettings) postgresConnString(databaseName string) string {
	connURL := &url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(a.PostgresUser, a.PostgresPassword),
		Host:   fmt.Sprintf("%s:%d", a.PostgresHost, a.PostgresPort),
		Path:   databaseName,
	}

	return connURL.String()
}
