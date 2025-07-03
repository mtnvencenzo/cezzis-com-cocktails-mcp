package cocktailsapi

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"

	"cezzis.com/cezzis-mcp-server/pkg/config"
)

var jwks *keyfunc.JWKS

func init() {
	appSettings := config.GetAppSettings()
	uri := appSettings.GetAzureAdB2CDiscoveryKeysUri()
	if uri == "" {
		// Don't panic, just log a warning and continue without auth
		fmt.Printf("Warning: Azure AD B2C discovery URI not configured, auth middleware will be disabled\n")
		return
	}

	var err error
	jwks, err = keyfunc.Get(uri, keyfunc.Options{})
	if err != nil {
		// Don't panic, just log a warning and continue without auth
		fmt.Printf("Warning: Failed to get JWKS: %v, auth middleware will be disabled\n", err)
		return
	}
}

func AuthMiddleware(requiredScopes []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If JWKS is not initialized, allow all requests through
			if jwks == nil {
				next.ServeHTTP(w, r)
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
			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
