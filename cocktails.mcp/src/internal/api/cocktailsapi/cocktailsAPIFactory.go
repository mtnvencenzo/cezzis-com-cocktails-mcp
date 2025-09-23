package cocktailsapi

import (
	"context"
	"errors"
	"net/http"

	"cezzis.com/cezzis-mcp-server/internal/config"
	l "cezzis.com/cezzis-mcp-server/internal/logging"
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
		l.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	client, err := NewClient(appSettings.CocktailsAPIHost)
	if err != nil {
		l.Logger.Error().Err(err).Msg(err.Error())
		return nil, err
	}

	return client, nil
}
