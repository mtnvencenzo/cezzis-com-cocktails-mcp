package config_test

import (
	"net/url"
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
	if appSettings.Auth0NativeClientID != expectedAuth0ClientID {
		t.Errorf("Expected Auth0NativeClientID to be %s, got %s", expectedAuth0ClientID, appSettings.Auth0NativeClientID)
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

	// assert
	if jwks != expectedJWKS {
		t.Errorf("Expected GetAuth0JWKSURI() to be %s, got %s", expectedJWKS, jwks)
	}
}

func Test_appsettings_escapes_postgres_userinfo(t *testing.T) {
	settings := &config.AppSettings{
		PostgresHost:     "psqlfs-vec-eus2-glo-shared-001.postgres.database.azure.com",
		PostgresPort:     5432,
		PostgresDBName:   "cezzis-cocktailsmcp-db-loc",
		PostgresUser:     "admin",
		PostgresPassword: "pa:ss@wo/rd?x#y%z",
	}

	connString := settings.PostgresConnString()
	parsedURL, err := url.Parse(connString)
	if err != nil {
		t.Fatalf("expected PostgresConnString() to produce a valid URL, got error: %v", err)
	}

	password, hasPassword := parsedURL.User.Password()
	if !hasPassword {
		t.Fatal("expected PostgresConnString() to include a password")
	}

	if parsedURL.User.Username() != settings.PostgresUser {
		t.Errorf("expected username %q, got %q", settings.PostgresUser, parsedURL.User.Username())
	}

	if password != settings.PostgresPassword {
		t.Errorf("expected password %q, got %q", settings.PostgresPassword, password)
	}

	if parsedURL.Path != "/"+settings.PostgresDBName {
		t.Errorf("expected database path %q, got %q", "/"+settings.PostgresDBName, parsedURL.Path)
	}

	adminConnString := settings.PostgresAdminConnString()
	adminURL, err := url.Parse(adminConnString)
	if err != nil {
		t.Fatalf("expected PostgresAdminConnString() to produce a valid URL, got error: %v", err)
	}

	if adminURL.Path != "/postgres" {
		t.Errorf("expected admin database path %q, got %q", "/postgres", adminURL.Path)
	}
	if adminPassword, _ := adminURL.User.Password(); adminPassword != settings.PostgresPassword {
		t.Errorf("expected admin password %q, got %q", settings.PostgresPassword, adminPassword)
	}
}
