package integration

import (
	"testing"

	"github.com/willibrandon/lazynuget/internal/bootstrap"
)

// TestParseFlagsVersion tests --version flag
func TestParseFlagsVersion(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags, shouldExit, err := app.ParseFlags([]string{"-version"})
	if err != nil {
		t.Fatalf("ParseFlags() failed: %v", err)
	}

	if !shouldExit {
		t.Error("Expected shouldExit=true for version flag")
	}

	if !flags.ShowVersion {
		t.Error("Expected ShowVersion=true")
	}
}

// TestParseFlagsHelp tests --help flag
func TestParseFlagsHelp(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags, shouldExit, err := app.ParseFlags([]string{"-help"})
	if err != nil {
		t.Fatalf("ParseFlags() failed: %v", err)
	}

	if !shouldExit {
		t.Error("Expected shouldExit=true for help flag")
	}

	if !flags.ShowHelp {
		t.Error("Expected ShowHelp=true")
	}
}

// TestParseFlagsMultiple tests multiple flags
func TestParseFlagsMultiple(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags, shouldExit, err := app.ParseFlags([]string{
		"-log-level", "debug",
		"-non-interactive",
		"-config", "/custom/config.yml",
	})
	if err != nil {
		t.Fatalf("ParseFlags() failed: %v", err)
	}

	if shouldExit {
		t.Error("Expected shouldExit=false for non-exit flags")
	}

	if flags.LogLevel != "debug" {
		t.Errorf("Expected LogLevel=debug, got %s", flags.LogLevel)
	}

	if !flags.NonInteractive {
		t.Error("Expected NonInteractive=true")
	}

	if flags.ConfigPath != "/custom/config.yml" {
		t.Errorf("Expected ConfigPath=/custom/config.yml, got %s", flags.ConfigPath)
	}
}

// TestShowVersion tests the ShowVersion function
func TestShowVersion(_ *testing.T) {
	version := bootstrap.VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2025-01-01",
	}

	// Should not panic
	bootstrap.ShowVersion(version)
}

// TestShowHelp tests the ShowHelp function
func TestShowHelp(_ *testing.T) {
	// Should not panic
	bootstrap.ShowHelp()
}

// TestVersionInfoString tests VersionInfo.String() method
func TestVersionInfoString(t *testing.T) {
	version := bootstrap.VersionInfo{
		Version: "1.0.0",
		Commit:  "abc123",
		Date:    "2025-01-01",
	}

	str := version.String()
	if str == "" {
		t.Error("VersionInfo.String() returned empty string")
	}

	// Should contain version, commit, and date
	if !containsStr(str, "1.0.0") {
		t.Errorf("Expected string to contain version, got: %s", str)
	}
}

// TestGetConfig tests the GetConfig() method
func TestGetConfig(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags := &bootstrap.Flags{
		NonInteractive: true,
	}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	config := app.GetConfig()
	if config == nil {
		t.Error("GetConfig() returned nil")
	}

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

// Helper function (renamed to avoid conflict with config_cli_test.go)
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
