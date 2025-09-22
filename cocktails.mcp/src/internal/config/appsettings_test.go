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
	expectedAzureAdB2CDomain := "cezzis.onmicrosoft.com"
	expectedAzureAdB2CInstance := "https://testlogin.cezzis.com"
	expectedAzureAdB2CUserFlow := "B2C_1_SignInSignUp_Policy"
	expectedCocktailsAPISubscriptionKey := "00000000-0000-0000-0000-000000000000"

	// act
	var appSettings *config.AppSettings = config.GetAppSettings()

	// asert
	if appSettings.CocktailsAPIHost != expectedCocktailsAPIHost {
		t.Errorf("Expected CocktailsAPIHost to be %s, got %s", expectedCocktailsAPIHost, appSettings.CocktailsAPIHost)
	}

	if appSettings.AzureAdB2CDomain != expectedAzureAdB2CDomain {
		t.Errorf("Expected AzureAdB2CDomain to be %s, got %s", expectedAzureAdB2CDomain, appSettings.AzureAdB2CDomain)
	}

	if appSettings.AzureAdB2CInstance != expectedAzureAdB2CInstance {
		t.Errorf("Expected AzureAdB2CInstance to be %s, got %s", expectedAzureAdB2CInstance, appSettings.AzureAdB2CInstance)
	}

	if appSettings.AzureAdB2CUserFlow != expectedAzureAdB2CUserFlow {
		t.Errorf("Expected AzureAdB2CUserFlow to be %s, got %s", expectedAzureAdB2CUserFlow, appSettings.AzureAdB2CUserFlow)
	}

	if appSettings.CocktailsAPISubscriptionKey != expectedCocktailsAPISubscriptionKey {
		t.Errorf("Expected CocktailsAPISubscriptionKey to be %s, got %s", expectedCocktailsAPISubscriptionKey, appSettings.CocktailsAPISubscriptionKey)
	}
}

func Test_appsettings_produces_correct_discovery_keys_url(t *testing.T) {
	testutils.LoadEnvironment("..", "..")

	// arrange
	expectedAzureAdB2CDiscoveryKeysURI := "https://testlogin.cezzis.com/cezzis.onmicrosoft.com/B2C_1_SignInSignUp_Policy/discovery/v2.0/keys"

	// act
	appSettings := config.GetAppSettings()
	azureAdB2CDiscoveryKeysURI := appSettings.GetAzureAdB2CDiscoveryKeysURI()

	// asert
	if azureAdB2CDiscoveryKeysURI != expectedAzureAdB2CDiscoveryKeysURI {
		t.Errorf("Expected GetAzureAdB2CDiscoveryKeysURI() to be %s, got %s", expectedAzureAdB2CDiscoveryKeysURI, azureAdB2CDiscoveryKeysURI)
	}
}
