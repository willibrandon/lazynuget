//go:build windows

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// TestWindowsConfigDirectory verifies that on Windows, the config directory
// resolves to %APPDATA%\lazynuget
func TestWindowsConfigDirectory(t *testing.T) {
	// Get platform instance
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	// Verify we're on Windows
	if !platformInfo.IsWindows() {
		t.Skip("Test only runs on Windows")
	}

	// Create PathResolver (will fail until implemented)
	// This test is expected to FAIL initially (TDD approach)
	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get config directory
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	// Verify it's under %APPDATA%
	appData := os.Getenv("APPDATA")
	if appData == "" {
		t.Fatal("APPDATA environment variable not set")
	}

	expectedDir := filepath.Join(appData, "lazynuget")
	if configDir != expectedDir {
		t.Errorf("ConfigDir() = %q, want %q", configDir, expectedDir)
	}

	// Verify path uses Windows separators (backslashes)
	if !filepath.IsAbs(configDir) {
		t.Errorf("ConfigDir() returned relative path: %q", configDir)
	}
}

// TestWindowsCacheDirectory verifies that on Windows, the cache directory
// resolves to %LOCALAPPDATA%\lazynuget
func TestWindowsCacheDirectory(t *testing.T) {
	// Get platform instance
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	// Verify we're on Windows
	if !platformInfo.IsWindows() {
		t.Skip("Test only runs on Windows")
	}

	// Create PathResolver
	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get cache directory
	cacheDir, err := pathResolver.CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() failed: %v", err)
	}

	// Verify it's under %LOCALAPPDATA%
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		t.Fatal("LOCALAPPDATA environment variable not set")
	}

	expectedDir := filepath.Join(localAppData, "lazynuget")
	if cacheDir != expectedDir {
		t.Errorf("CacheDir() = %q, want %q", cacheDir, expectedDir)
	}

	// Verify path is absolute
	if !filepath.IsAbs(cacheDir) {
		t.Errorf("CacheDir() returned relative path: %q", cacheDir)
	}
}
