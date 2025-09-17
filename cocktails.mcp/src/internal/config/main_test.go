package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	// Get the current working directory before changing it.
	// originalDir, err := os.Getwd()
	// if err != nil {
	// 	log.Fatalf("Failed to get current working directory: %v", err)
	// }

	wd, wderr := os.Getwd()
	if wderr != nil {
		fmt.Printf("Server error - wd path: %v\n", wderr)
	}

	// exePath, oserr := os.Executable()
	// if oserr != nil {
	// 	fmt.Printf("Server error - exe path: %v\n", oserr)
	// }

	// var finalEnvPath string = envPath

	// if finalEnvPath == "" {
	// 	finalEnvPath = filepath.Dir(exePath)
	// }

	envPath := filepath.Join(wd, "..", "..")

	_ = godotenv.Overload(
		fmt.Sprintf("%s/%s", envPath, ".env"),
		fmt.Sprintf("%s/%s", envPath, ".env.local"))

	// Change the current working directory.
	// if err := os.Chdir("../../"); err != nil {
	// 	log.Fatalf("Failed to change directory: %v", err)
	// }

	// Run all tests in the package.
	exitCode := m.Run()

	// Restore the original working directory.
	// if err := os.Chdir(originalDir); err != nil {
	// 	log.Fatalf("Failed to restore original directory to %s: %v", originalDir, err)
	// }

	// Exit with the appropriate code from the test runner.
	os.Exit(exitCode)
}

// func init() {
//     _, filename, _, _ := runtime.Caller(0)
//     // The ".." may change depending on you folder structure
//     dir := path.Join(path.Dir(filename), "..")
//     err := os.Chdir(dir)
//     if err != nil {
//         panic(err)
//     }
// }
