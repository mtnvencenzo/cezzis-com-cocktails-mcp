package cocktailsapi

import (
	"context"
	"net/http"
)

// ICocktailsAPI defines the interface for interacting with the cocktails API.
// It provides methods for retrieving individual cocktails and listing cocktails.
type ICocktailsAPI interface {
	GetCocktail(ctx context.Context, id string, params *GetCocktailParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetCocktailsList(ctx context.Context, params *GetCocktailsListParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	RateCocktailWithApplicationJSONXAPIVersion10Body(ctx context.Context, params *RateCocktailParams, body RateCocktailApplicationJSONXAPIVersion10RequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}
