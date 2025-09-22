// Package cocktailsapi provides HTTP client functionality and authentication middleware
// for interacting with the Cocktails API service. It includes generated API client code
// and custom authentication middleware for Azure AD B2C integration.
//
// The package features:
//   - Auto-generated HTTP client code from OpenAPI specifications
//   - Azure AD B2C JWT token validation middleware
//   - JSON Web Key Set (JWKS) integration for secure token verification
//   - Scope-based authorization with configurable required permissions
//
// The authentication middleware supports graceful degradation - if Azure AD B2C
// configuration is missing or invalid, the middleware will allow all requests
// through without authentication, making it suitable for development environments.
package cocktailsapi

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"

	"cezzis.com/cezzis-mcp-server/internal/config"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// userContextKey is the key used to store user claims in the request context
const userContextKey contextKey = "user"

var jwks *keyfunc.JWKS

func init() {
	appSettings := config.GetAppSettings()
	uri := appSettings.GetAzureAdB2CDiscoveryKeysURI()

	if uri == "" {
		l.Logger.Warn().Msg("Warning: Azure AD B2C discovery URI not configured\n")
		if strings.ToLower(os.Getenv("ENV")) != "local" && strings.ToLower(os.Getenv("ENV")) != "test" {
			// In non-local, do not disable auth silently
			return
		}
		// local/test: leave jwks nil to permit fail-open
		return
	}

	if uri == "" {
		// Don't panic, just log a warning and continue without auth
		l.Logger.Warn().Msg("Warning: Azure AD B2C discovery URI not configured, auth middleware will be disabled\n")
		return
	}

	var err error
	jwks, err = keyfunc.Get(uri, keyfunc.Options{})
	if err != nil {
		// Don't panic, just log a warning and continue without auth
		l.Logger.Warn().Err(err).Msg("Warning: Failed to get JWKS: auth middleware will be disabled")
		return
	}
}

// AuthMiddleware creates an HTTP middleware function that validates Azure AD B2C JWT tokens
// and enforces scope-based authorization. The middleware can be configured to require
// specific scopes for access to protected endpoints.
//
// Parameters:
//   - requiredScopes: A slice of scope strings that must be present in the JWT token.
//     If empty, no authentication is required and all requests are allowed through.
//
// Behavior:
//   - If Azure AD B2C is not configured (JWKS is nil), all requests are allowed through
//   - If no scopes are required, all requests are allowed through
//   - Validates the Authorization header format (must start with "Bearer ")
//   - Verifies JWT token signature using the configured JWKS
//   - Checks that all required scopes are present in the token claims
//   - Adds user claims to the request context for downstream handlers
//
// Returns an HTTP middleware function that can be used with any HTTP handler.
// The middleware returns appropriate HTTP error responses (401 Unauthorized, 403 Forbidden)
// for authentication and authorization failures.
func AuthMiddleware(requiredScopes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow-through only in local/test; otherwise fail closed
			if jwks == nil && (strings.ToLower(os.Getenv("ENV")) == "local" || strings.ToLower(os.Getenv("ENV")) == "test") {
				next.ServeHTTP(w, r)
				return
			} else if jwks == nil {
				http.Error(w, "Authentication service unavailable", http.StatusServiceUnavailable)
				return
			}

			// No need to check auth if no scopes are required
			if len(requiredScopes) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			token, err := jwt.Parse(tokenString, jwks.Keyfunc)
			if err != nil || !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Check scopes
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			scopes, ok := claims["scp"].(string) // or "roles" or "scope" depending on your config
			if !ok {
				http.Error(w, "Insufficient scope", http.StatusForbidden)
				return
			}

			scopesList := strings.Split(scopes, " ")

			// Check for all required scopes
			for _, required := range requiredScopes {
				found := false
				for _, s := range scopesList {
					if s == required {
						found = true
						break
					}
				}
				if !found {
					http.Error(w, "Insufficient scope: "+required, http.StatusForbidden)
					return
				}
			}

			// Optionally, set user info in context
			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
