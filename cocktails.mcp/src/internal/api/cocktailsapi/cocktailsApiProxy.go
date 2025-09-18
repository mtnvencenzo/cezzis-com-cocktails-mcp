package cocktailsapi

import (
	"context"
	"net/http"
)

type ICocktailsApiProxy interface {
	GetCocktail(ctx context.Context, id string, params *GetCocktailParams, reqEditors ...RequestEditorFn) (*http.Response, error)
	GetCocktailsList(ctx context.Context, params *GetCocktailsListParams, reqEditors ...RequestEditorFn) (*http.Response, error)
}

type CocktailsApiProxy struct {
	apiClient Client
}

func NewCocktailsApiProxy(apiClient Client) CocktailsApiProxy {
	return CocktailsApiProxy{
		apiClient,
	}
}

func (proxy CocktailsApiProxy) GetCocktail(ctx context.Context, id string, params *GetCocktailParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return proxy.apiClient.GetCocktail(ctx, id, params)
}

func (proxy CocktailsApiProxy) GetCocktailsList(ctx context.Context, params *GetCocktailsListParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return proxy.apiClient.GetCocktailsList(ctx, params)
}
