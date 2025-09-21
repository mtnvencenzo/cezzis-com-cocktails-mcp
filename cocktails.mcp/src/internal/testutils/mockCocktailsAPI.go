package testutils

import (
	"context"
	"net/http"

	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
)

// MockCocktailsAPI is a lightweight in-memory implementation of the
// cocktails API client used for tests. It returns simple placeholder
// responses and is intended to be wired into tests that need a
// cocktailsapi-compatible client without performing real HTTP calls.
type MockCocktailsAPI struct{}

// NewMockCocktailsAPI constructs and returns a ready-to-use
// MockCocktailsAPI value. The returned value has no internal state and
// can be copied freely.
func NewMockCocktailsAPI() MockCocktailsAPI {
	return MockCocktailsAPI{}
}

// GetCocktail implements the cocktailsapi client's GetCocktail method.
//
// In this mock it returns an empty HTTP response and a nil error. Tests
// can replace or extend this method if they need to return specific
// payloads or simulate error cases.
func (api MockCocktailsAPI) GetCocktail(ctx context.Context, id string, params *cocktailsapi.GetCocktailParams, reqEditors ...cocktailsapi.RequestEditorFn) (*http.Response, error) {
	return &http.Response{}, nil
}

// GetCocktailsList implements the cocktailsapi client's GetCocktailsList
// method.
//
// The mock returns an empty HTTP response and a nil error by default.
// Tests that require specific list data should either replace this mock
// or inspect/modify the response returned here.
func (api MockCocktailsAPI) GetCocktailsList(ctx context.Context, params *cocktailsapi.GetCocktailsListParams, reqEditors ...cocktailsapi.RequestEditorFn) (*http.Response, error) {
	return &http.Response{}, nil
}
