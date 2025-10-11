package cocktailsapi

import (
	"errors"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/logging"
)

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
