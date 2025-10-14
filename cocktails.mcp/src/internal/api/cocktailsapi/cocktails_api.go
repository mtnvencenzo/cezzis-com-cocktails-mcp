// Package cocktailsapi provides a client for interacting with the Cocktails API.
// It includes methods for retrieving cocktail details, searching for cocktails,
// and rating cocktails. The package handles authentication, request construction,
// and response parsing.
package cocktailsapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"cezzis.com/cezzis-mcp-server/internal/auth"
	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/logging"
	"cezzis.com/cezzis-mcp-server/internal/mcpserver"
)

// AuthenticatedRequestEditor creates a request editor that adds OAuth bearer token
func AuthenticatedRequestEditor(authManager *auth.OAuthFlowManager) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		sessionID := ctx.Value(mcpserver.McpSessionIDKey)
		if sessionID == nil || sessionID == "" {
			return errors.New("missing required Mcp-Session-Id header")
		}

		// Add subscription key header
		appSettings := config.GetAppSettings()
		if appSettings.CocktailsAPISubscriptionKey != "" {
			req.Header.Set("X-Key", appSettings.CocktailsAPISubscriptionKey)
		}

		// Add OAuth bearer token if authenticated
		if authManager.IsAuthenticated(sessionID.(string)) {
			token, err := authManager.GetAccessToken(ctx, sessionID.(string))
			if err != nil {
				logging.Logger.Warn().Err(err).Msg("Failed to get access token")
				return err
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			logging.Logger.Debug().Msg("Added OAuth bearer token to request")
		}

		return nil
	}
}

// GetClient creates and returns a new cocktails API client using application settings.
// Returns an error if the client creation fails.
func GetClient() (*Client, error) {
	appSettings := config.GetAppSettings()

	if appSettings.CocktailsAPIHost == "" {
		err := errors.New("CocktailsAPIHost has not been configured")
		logging.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	client, err := NewClient(appSettings.CocktailsAPIHost)
	if err != nil {
		logging.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	return client, nil
}
