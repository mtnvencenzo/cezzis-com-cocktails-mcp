package config

import (
	"testing"
)

func Test_Appsettings(t *testing.T) {

	// arrange, act
	var appSettings *AppSettings = GetAppSettings()

	// asert
	if appSettings.CocktailsAPIHost == "" {
		t.Errorf("Expected CocktailsAPIHost to be set, but it was empty")
	}
}
