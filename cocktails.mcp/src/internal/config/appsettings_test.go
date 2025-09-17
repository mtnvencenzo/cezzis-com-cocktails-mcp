package config

import (
	"testing"
)

func Test_Appsettings(t *testing.T) {

	// arrange
	expectedAPIHost := "https://testapi.cezzis.com/prd/cocktails"

	// act
	var appSettings *AppSettings = GetAppSettings()

	// asert
	if appSettings.CocktailsAPIHost != expectedAPIHost {
		t.Errorf("Expected CocktailsAPIHost to be %s, got %s", expectedAPIHost, appSettings.CocktailsAPIHost)
	}
}
