package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AppSettings struct {
	CocktailsApiHost            string `env:"COCKTAILS_API_HOST"`
	CocktailsApiSubscriptionKey string `env:"COCKTAILS_API_XKEY"`

	AzureAdB2CInstance string `env:"AZUREAD_B2C_INSTANCE"`
	AzureAdB2CDomain   string `env:"AZUREAD_B2C_DOMAIN"`
	AzureAdB2CUserFlow string `env:"AZUREAD_B2C_USERFLOW"`
}

func (a *AppSettings) GetAzureAdB2CDiscoveryKeysUri() string {
	return fmt.Sprintf("%s/%s/%s/discovery/v2.0/keys", a.AzureAdB2CInstance, a.AzureAdB2CDomain, a.AzureAdB2CUserFlow)
}

var (
	instance *AppSettings
	once     sync.Once
)

func GetAppSettings() *AppSettings {
	once.Do(func() {
		exePath, oserr := os.Executable()
		if oserr != nil {
			fmt.Printf("Server error - exe path: %v\n", oserr)
		}

		// Get the directory of the executable
		exeDir := filepath.Dir(exePath)

		_ = godotenv.Overload(
			fmt.Sprintf("%s/%s", exeDir, ".env"),
			fmt.Sprintf("%s/%s", exeDir, ".env.local"))

		instance = &AppSettings{}
		if err := env.Parse(instance); err != nil {
			fmt.Printf("Exe Path: %v\n", exeDir)
			fmt.Printf("Failed to parse app settings: %v\n", err)
		}
	})

	return instance
}
