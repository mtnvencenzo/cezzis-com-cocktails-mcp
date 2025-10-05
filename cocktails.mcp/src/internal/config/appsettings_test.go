package config_test

import (
	"testing"

	"cezzis.com/cezzis-mcp-server/internal/config"
	"cezzis.com/cezzis-mcp-server/internal/testutils"
)

func Test_appsettings_retreives_correct_env_file_settings(t *testing.T) {
	testutils.LoadEnvironment("..", "..")

	// arrange
	expectedCocktailsAPIHost := "https://testapi.cezzis.com/prd/cocktails"
	expectedAuth0Domain := "login.test.cezzis.com"
	expectedAuth0ClientID := "00000000-0000-0000-0000-000000000000"
	expectedCocktailsAPISubscriptionKey := "00000000-0000-0000-0000-000000000000"

	// act
	var appSettings *config.AppSettings = config.GetAppSettings()

	// asert
	if appSettings.CocktailsAPIHost != expectedCocktailsAPIHost {
		t.Errorf("Expected CocktailsAPIHost to be %s, got %s", expectedCocktailsAPIHost, appSettings.CocktailsAPIHost)
	}

	if appSettings.Auth0Domain != expectedAuth0Domain {
		t.Errorf("Expected Auth0Domain to be %s, got %s", expectedAuth0Domain, appSettings.Auth0Domain)
	}
	if appSettings.Auth0ClientID != expectedAuth0ClientID {
		t.Errorf("Expected Auth0ClientID to be %s, got %s", expectedAuth0ClientID, appSettings.Auth0ClientID)
	}

	if appSettings.CocktailsAPISubscriptionKey != expectedCocktailsAPISubscriptionKey {
		t.Errorf("Expected CocktailsAPISubscriptionKey to be %s, got %s", expectedCocktailsAPISubscriptionKey, appSettings.CocktailsAPISubscriptionKey)
	}
}

func Test_appsettings_produces_correct_jwks_url(t *testing.T) {
	testutils.LoadEnvironment("..", "..")

	// arrange
	expectedJWKS := "https://login.test.cezzis.com/.well-known/jwks.json"

	// act
	appSettings := config.GetAppSettings()
	jwks := appSettings.GetAuth0JWKSURI()

	// asert
	if jwks != expectedJWKS {
		t.Errorf("Expected GetAuth0JWKSURI() to be %s, got %s", expectedJWKS, jwks)
	}
}
