//coverage:ignore file

package testutils

import (
	"cezzis.com/cezzis-mcp-server/internal/api/cocktailsapi"
)

// MockCocktailsAPIFactory is a simple factory that returns a preconfigured
// cocktails API client for use in tests. It satisfies any factory or
// dependency injection points that expect a provider of
// cocktailsapi.ICocktailsAPI.
//
// The factory holds a reference to an ICocktailsAPI implementation which
// is returned by GetClient. This allows tests to inject mock clients and
// control the behavior of downstream code that requests a client from the
// factory.
type MockCocktailsAPIFactory struct {
	api cocktailsapi.ICocktailsAPI
}

// NewMockCocktailsAPIFactory constructs a MockCocktailsAPIFactory that will
// always return the provided ICocktailsAPI when GetClient is called. Use
// this in tests to inject a prepared mock or stub implementation.
func NewMockCocktailsAPIFactory(api cocktailsapi.ICocktailsAPI) MockCocktailsAPIFactory {
	return MockCocktailsAPIFactory{
		api,
	}
}

// GetClient returns the preconfigured ICocktailsAPI held by the factory.
// The function returns an error to match the factory interface used in
// production; the mock factory always returns a nil error.
func (factory MockCocktailsAPIFactory) GetClient() (cocktailsapi.ICocktailsAPI, error) {
	return factory.api, nil
}
