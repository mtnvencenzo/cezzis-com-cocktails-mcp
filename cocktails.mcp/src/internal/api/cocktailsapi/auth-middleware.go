// Package cocktailsapi provides HTTP client functionality and authentication middleware
// for interacting with the Cocktails API service. It includes generated API client code
// and custom authentication middleware for Auth0 integration.
//
// The package features:
//   - Auto-generated HTTP client code from OpenAPI specifications
//   - Auth0 JWT token validation middleware
//   - JSON Web Key Set (JWKS) integration for secure token verification
//   - Scope-based authorization with configurable required permissions
//
// The authentication middleware supports graceful degradation - if Auth0
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
	uri := appSettings.GetAuth0JWKSURI()

	if uri == "" {
		l.Logger.Warn().Msg("Warning: Auth0 JWKS URI not configured\n")
		if strings.ToLower(os.Getenv("ENV")) != "local" && strings.ToLower(os.Getenv("ENV")) != "test" {
			// In non-local, do not disable auth silently
			return
		}
		// local/test: leave jwks nil to permit fail-open
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

// AuthMiddleware creates an HTTP middleware function that validates Auth0 JWT tokens
// and enforces scope-based authorization. The middleware can be configured to require
// specific scopes for access to protected endpoints.
//
// Parameters:
//   - requiredScopes: A slice of scope strings that must be present in the JWT token.
//     If empty, no authentication is required and all requests are allowed through.
//
// Behavior:
//   - If Auth0 is not configured (JWKS is nil), all requests are allowed through in local/test
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

			// Auth0 typically uses "scope" (space-separated). Some providers use "scp".
			scopes, ok := claims["scope"].(string)
			if !ok {
				scopes, _ = claims["scp"].(string)
			}
			if !ok {
				http.Error(w, "Insufficient scope", http.StatusForbidden)
				return
			}

			scopesList := strings.Split(scopes, " ")

			if !hasAllScopes(scopesList, requiredScopes) {
				http.Error(w, "Insufficient scope", http.StatusForbidden)
				return
			}

			// Optionally, set user info in context
			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// hasAllScopes returns true if all required scopes are present in the provided list
func hasAllScopes(have []string, required []string) bool {
	for _, req := range required {
		found := false
		for _, s := range have {
			if s == req {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
