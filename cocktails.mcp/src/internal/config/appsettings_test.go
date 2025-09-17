package config

import (
	"testing"
)

func Test_Appsettings(t *testing.T) {

	// arrange
	expectedApiHost := "https://testapi.cezzis.com/prd/cocktails"

	// act
	var appSettings *AppSettings = GetAppSettings()

	// asert
	if appSettings.CocktailsAPIHost != expectedApiHost {
		t.Errorf("Expected CocktailsAPIHost to be %s, got %s", expectedApiHost, appSettings.CocktailsAPIHost)
	}
}
