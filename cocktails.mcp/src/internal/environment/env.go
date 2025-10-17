// Package environment provides functionality to load environment variables from .env files.
package environment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from .env files located in the executable directory.
// It supports loading multiple .env files based on the ENV environment variable.
// For example, if ENV=development, it will load .env and .env.development if they exist.
// An optional ENV_DIR_OVERRIDE environment variable can be set to specify a different directory for the .env files.
func LoadEnv() {
	// Set up environment variables from .env files in the executable directory
	// This allows configuration settings to be loaded at runtime.
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	envFileDir := exeDir
	fmt.Println("Exe dir:", exeDir)

	if os.Getenv("ENV_DIR_OVERRIDE") != "" {
		envFileDir = os.Getenv("ENV_DIR_OVERRIDE")
	}

	env := os.Getenv("ENV")
	baseEnvFile := filepath.Join(envFileDir, ".env")
	candidates := []string{baseEnvFile}

	if env != "" {
		candidates = append(candidates, baseEnvFile+"."+env)
	}

	toLoad := make([]string, 0, len(candidates))
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			fmt.Println("Loading env file:", f)
			toLoad = append(toLoad, f)
		}
	}

	if len(toLoad) > 0 {
		_ = godotenv.Overload(toLoad...)
	}
}

// IsLocalEnv returns true if the ENV environment variable is set to "local"
func IsLocalEnv() bool {
	return strings.ToLower(os.Getenv("ENV")) == "local"
}
