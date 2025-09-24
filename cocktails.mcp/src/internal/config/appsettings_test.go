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
	expectedAzureCIAMDomain := "cezzis.onmicrosoft.com"
	expectedAzureCIAMInstance := "https://testlogin.cezzis.com"
	expectedAzureCIAMUserFlow := "sisu-p"
	expectedCocktailsAPISubscriptionKey := "00000000-0000-0000-0000-000000000000"

	// act
	var appSettings *config.AppSettings = config.GetAppSettings()

	// asert
	if appSettings.CocktailsAPIHost != expectedCocktailsAPIHost {
		t.Errorf("Expected CocktailsAPIHost to be %s, got %s", expectedCocktailsAPIHost, appSettings.CocktailsAPIHost)
	}

	if appSettings.AzureCIAMDomain != expectedAzureCIAMDomain {
		t.Errorf("Expected AzureCIAMDomain to be %s, got %s", expectedAzureCIAMDomain, appSettings.AzureCIAMDomain)
	}

	if appSettings.AzureCIAMInstance != expectedAzureCIAMInstance {
		t.Errorf("Expected AzureCIAMInstance to be %s, got %s", expectedAzureCIAMInstance, appSettings.AzureCIAMInstance)
	}

	if appSettings.AzureCIAMUserFlow != expectedAzureCIAMUserFlow {
		t.Errorf("Expected AzureCIAMUserFlow to be %s, got %s", expectedAzureCIAMUserFlow, appSettings.AzureCIAMUserFlow)
	}

	if appSettings.CocktailsAPISubscriptionKey != expectedCocktailsAPISubscriptionKey {
		t.Errorf("Expected CocktailsAPISubscriptionKey to be %s, got %s", expectedCocktailsAPISubscriptionKey, appSettings.CocktailsAPISubscriptionKey)
	}
}

func Test_appsettings_produces_correct_discovery_keys_url(t *testing.T) {
	testutils.LoadEnvironment("..", "..")

	// arrange
	expectedAzureCIAMDiscoveryKeysURI := "https://testlogin.cezzis.com/cezzis.onmicrosoft.com/sisu-p/discovery/v2.0/keys"

	// act
	appSettings := config.GetAppSettings()
	azureCIAMDiscoveryKeysURI := appSettings.GetAzureCIAMDiscoveryKeysURI()

	// asert
	if azureCIAMDiscoveryKeysURI != expectedAzureCIAMDiscoveryKeysURI {
		t.Errorf("Expected GetAzureCIAMDiscoveryKeysURI() to be %s, got %s", expectedAzureCIAMDiscoveryKeysURI, azureCIAMDiscoveryKeysURI)
	}
}
