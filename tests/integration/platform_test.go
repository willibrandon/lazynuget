package integration

import (
	"testing"

	"github.com/willibrandon/lazynuget/internal/bootstrap"
)

// TestPlatformInfo tests that platform detection works in integration
func TestPlatformInfo(t *testing.T) {
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

	platform := app.GetPlatform()
	if platform == nil {
		t.Fatal("GetPlatform() returned nil")
	}

	// Test OS() method
	os := platform.OS()
	if os == "" {
		t.Error("OS() returned empty string")
	}

	// Test Arch() method
	arch := platform.Arch()
	if arch == "" {
		t.Error("Arch() returned empty string")
	}

	// Test RunMode String() method
	runMode := app.GetRunMode()
	runModeStr := runMode.String()
	if runModeStr != "interactive" && runModeStr != "non-interactive" {
		t.Errorf("RunMode.String() = %q, want 'interactive' or 'non-interactive'", runModeStr)
	}
}

// TestPlatformWithInteractiveMode tests platform detection in interactive mode
func TestPlatformWithInteractiveMode(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Don't set NonInteractive flag - let it detect naturally
	flags := &bootstrap.Flags{}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	platform := app.GetPlatform()
	if platform == nil {
		t.Fatal("GetPlatform() returned nil")
	}

	// Just verify the string method works for both modes
	runMode := app.GetRunMode()
	_ = runMode.String()
}
