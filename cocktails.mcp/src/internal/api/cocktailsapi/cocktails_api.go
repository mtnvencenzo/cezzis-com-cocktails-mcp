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
)

// ICocktailsAPI defines the interface for interacting with the cocktails API.
// It provides methods for retrieving individual cocktails and listing cocktails.
type ICocktailsAPI interface {
	GetCocktail(ctx context.Context, id string, params *GetCocktailParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetCocktailsList(ctx context.Context, params *GetCocktailsListParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	RateCocktailWithApplicationJSONXAPIVersion10Body(ctx context.Context, params *RateCocktailParams, body RateCocktailApplicationJSONXAPIVersion10RequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

// ICocktailsAPIFactory defines the interface for creating cocktails API clients.
type ICocktailsAPIFactory interface {
	GetClient() (ICocktailsAPI, error)
}

// AuthenticatedRequestEditor creates a request editor that adds OAuth bearer token
func AuthenticatedRequestEditor(authManager *auth.OAuthFlowManager) RequestEditorFn {
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
				logging.Logger.Warn().Err(err).Msg("Failed to get access token")
				return err
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			logging.Logger.Debug().Msg("Added OAuth bearer token to request")
		}

		return nil
	}
}

// AuthenticatedCocktailsAPIFactory extends the base factory with authentication support
type AuthenticatedCocktailsAPIFactory struct {
	CocktailsAPIFactory
	authManager *auth.OAuthFlowManager
}

// NewAuthenticatedCocktailsAPIFactory creates a new authenticated API factory
func NewAuthenticatedCocktailsAPIFactory(authManager *auth.OAuthFlowManager) *AuthenticatedCocktailsAPIFactory {
	return &AuthenticatedCocktailsAPIFactory{
		CocktailsAPIFactory: NewCocktailsAPIFactory(),
		authManager:         authManager,
	}
}

// CocktailsAPIFactory implements the ICocktailsAPIFactory interface and
// provides functionality to create new cocktails API clients.
type CocktailsAPIFactory struct { //nolint:revive
}

// NewCocktailsAPIFactory creates a new instance of CocktailsAPIFactory.
func NewCocktailsAPIFactory() CocktailsAPIFactory {
	return CocktailsAPIFactory{}
}

// GetClient creates and returns a new cocktails API client using application settings.
// Returns an error if the client creation fails.
func (factory CocktailsAPIFactory) GetClient() (ICocktailsAPI, error) {
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
