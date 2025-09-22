package config_test

import (
	"os"
	"testing"

	"cezzis.com/cezzis-mcp-server/internal/testutils"
)

func TestMain(m *testing.M) {
	testutils.LoadEnvironment("..", "..")

	// Run all tests in the package.
	exitCode := m.Run()

	// Exit with the appropriate code from the test runner.
	os.Exit(exitCode)
}
