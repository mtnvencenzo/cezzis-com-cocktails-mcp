//coverage:ignore file

// Package testutils provides shared test utilities for the Cezzi Cocktails MCP server.
// This file is only compiled when the 'test' build tag is used.
package testutils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"cezzis.com/cezzis-mcp-server/internal/telemetry"
)

// LoadEnvironment loads environment variables from .env files for tests.
//
// It determines the base path from the current working directory and then
// appends any provided wdOffsets to form the directory where it looks for
// the files. It attempts to overload environment variables from two files
// (in this order):
//   - .env
//   - .env.test
//
// Overload is used so values in these files will replace any existing
// environment variables for the duration of the test run.
//
// This function is intended to be used only in test builds (see package
// comment) to set up environment values required by tests.
func LoadEnvironment(wdOffsets ...string) {
	wd, wderr := os.Getwd()
	if wderr != nil {
		telemetry.Logger.Warn().Err(wderr).Msg("Server error - wd path: %v\n")
	}

	envPath := filepath.Join(append([]string{wd}, wdOffsets...)...)

	_ = godotenv.Overload(
		fmt.Sprintf("%s/%s", envPath, ".env"),
		fmt.Sprintf("%s/%s", envPath, ".env.test"))
}
