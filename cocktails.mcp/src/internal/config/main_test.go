package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	wd, wderr := os.Getwd()
	if wderr != nil {
		fmt.Printf("Server error - wd path: %v\n", wderr)
	}

	envPath := filepath.Join(wd, "..", "..")

	_ = godotenv.Overload(
		fmt.Sprintf("%s/%s", envPath, ".env"),
		fmt.Sprintf("%s/%s", envPath, ".env.test"))

	// Run all tests in the package.
	exitCode := m.Run()

	// Exit with the appropriate code from the test runner.
	os.Exit(exitCode)
}
