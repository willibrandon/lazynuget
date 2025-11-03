package integration

import (
	"os"
	"testing"

	"github.com/willibrandon/lazynuget/internal/bootstrap"
)

// TestWithCIEnvironment tests behavior in CI environment
func TestWithCIEnvironment(t *testing.T) {
	// Save original env
	origCI := os.Getenv("CI")
	defer os.Setenv("CI", origCI)

	// Set CI environment variable
	os.Setenv("CI", "true")

	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Don't set NonInteractive flag - should auto-detect from CI env
	flags := &bootstrap.Flags{}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	platform := app.GetPlatform()
	if platform == nil {
		t.Fatal("GetPlatform() returned nil")
	}

	// In CI environment, should be non-interactive
	runMode := app.GetRunMode()
	if runMode.IsInteractive() {
		t.Error("Expected non-interactive mode in CI environment")
	}

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

// TestWithGitHubActionsEnvironment tests GitHub Actions detection
func TestWithGitHubActionsEnvironment(t *testing.T) {
	// Save original env
	origGHA := os.Getenv("GITHUB_ACTIONS")
	defer os.Setenv("GITHUB_ACTIONS", origGHA)

	// Set GitHub Actions environment variable
	os.Setenv("GITHUB_ACTIONS", "true")

	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags := &bootstrap.Flags{}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	platform := app.GetPlatform()
	if platform == nil {
		t.Fatal("GetPlatform() returned nil")
	}

	// In GitHub Actions, should be non-interactive
	runMode := app.GetRunMode()
	if runMode.IsInteractive() {
		t.Error("Expected non-interactive mode in GitHub Actions")
	}

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

// TestWithNoColorEnvironment tests NO_COLOR detection
func TestWithNoColorEnvironment(t *testing.T) {
	// Save original env
	origNoColor := os.Getenv("NO_COLOR")
	defer os.Setenv("NO_COLOR", origNoColor)

	// Set NO_COLOR environment variable
	os.Setenv("NO_COLOR", "1")

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

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

// TestWithDumbTerminal tests TERM=dumb detection
func TestWithDumbTerminal(t *testing.T) {
	// Save original env
	origTerm := os.Getenv("TERM")
	defer os.Setenv("TERM", origTerm)

	// Set TERM to dumb
	os.Setenv("TERM", "dumb")

	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags := &bootstrap.Flags{}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	platform := app.GetPlatform()
	if platform == nil {
		t.Fatal("GetPlatform() returned nil")
	}

	// With TERM=dumb, should be non-interactive
	runMode := app.GetRunMode()
	if runMode.IsInteractive() {
		t.Error("Expected non-interactive mode with TERM=dumb")
	}

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

// TestWithMultipleCIEnvironmentVars tests various CI environment variables
func TestWithMultipleCIEnvironmentVars(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		value  string
	}{
		{name: "CircleCI", envVar: "CIRCLECI", value: "true"},
		{name: "Travis", envVar: "TRAVIS", value: "true"},
		{name: "Jenkins", envVar: "JENKINS_HOME", value: "/var/jenkins"},
		{name: "GitLab", envVar: "GITLAB_CI", value: "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env
			orig := os.Getenv(tt.envVar)
			defer os.Setenv(tt.envVar, orig)

			os.Setenv(tt.envVar, tt.value)

			app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
			if err != nil {
				t.Fatalf("NewApp() failed: %v", err)
			}

			flags := &bootstrap.Flags{}

			if err := app.Bootstrap(flags); err != nil {
				t.Fatalf("Bootstrap() failed: %v", err)
			}

			platform := app.GetPlatform()
			if platform == nil {
				t.Fatal("GetPlatform() returned nil")
			}

			// Should detect as non-interactive in CI
			runMode := app.GetRunMode()
			if runMode.IsInteractive() {
				t.Errorf("Expected non-interactive mode in %s", tt.name)
			}

			if err := app.Shutdown(); err != nil {
				t.Errorf("Shutdown() failed: %v", err)
			}
		})
	}
}
