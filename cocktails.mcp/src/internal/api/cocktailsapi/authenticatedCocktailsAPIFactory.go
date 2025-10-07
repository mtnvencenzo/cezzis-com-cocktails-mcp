// Package cocktailsapi provides primitives to interact with the openapi HTTP API.
package cocktailsapi

import (
	"context"
	"fmt"
	"net/http"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/config"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
)

// AuthenticatedRequestEditor creates a request editor that adds OAuth bearer token
func AuthenticatedRequestEditor(authManager *auth.Manager) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		// Add subscription key header
		appSettings := config.GetAppSettings()
		if appSettings.CocktailsAPISubscriptionKey != "" {
			req.Header.Set("X-Key", appSettings.CocktailsAPISubscriptionKey)
		}

		// Add OAuth bearer token if authenticated
		if authManager.IsAuthenticated() {
			token, err := authManager.GetAccessToken(ctx)
			if err != nil {
				l.Logger.Warn().Err(err).Msg("Failed to get access token")
				return err
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			l.Logger.Debug().Msg("Added OAuth bearer token to request")
		}

		return nil
	}
}

// AuthenticatedCocktailsAPIFactory extends the base factory with authentication support
type AuthenticatedCocktailsAPIFactory struct {
	CocktailsAPIFactory
	authManager *auth.Manager
}

// NewAuthenticatedCocktailsAPIFactory creates a new authenticated API factory
func NewAuthenticatedCocktailsAPIFactory(authManager *auth.Manager) *AuthenticatedCocktailsAPIFactory {
	return &AuthenticatedCocktailsAPIFactory{
		CocktailsAPIFactory: NewCocktailsAPIFactory(),
		authManager:         authManager,
	}
}

// GetAuthenticatedRequestEditor returns a request editor with authentication
func (factory *AuthenticatedCocktailsAPIFactory) GetAuthenticatedRequestEditor() RequestEditorFn {
	return AuthenticatedRequestEditor(factory.authManager)
}
