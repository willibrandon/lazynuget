//go:build linux

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// TestLinuxConfigDirectory verifies that on Linux, the config directory
// respects XDG_CONFIG_HOME or defaults to ~/.config/lazynuget
func TestLinuxConfigDirectory(t *testing.T) {
	// Get platform instance
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	// Verify we're on Linux
	if !platformInfo.IsLinux() {
		t.Skip("Test only runs on Linux")
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

	// Determine expected directory
	var expectedDir string
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		expectedDir = filepath.Join(xdgConfig, "lazynuget")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() failed: %v", err)
		}
		expectedDir = filepath.Join(homeDir, ".config", "lazynuget")
	}

	if configDir != expectedDir {
		t.Errorf("ConfigDir() = %q, want %q", configDir, expectedDir)
	}

	// Verify path uses Unix separators (forward slashes)
	if !filepath.IsAbs(configDir) {
		t.Errorf("ConfigDir() returned relative path: %q", configDir)
	}
}

// TestLinuxCacheDirectory verifies that on Linux, the cache directory
// respects XDG_CACHE_HOME or defaults to ~/.cache/lazynuget
func TestLinuxCacheDirectory(t *testing.T) {
	// Get platform instance
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	// Verify we're on Linux
	if !platformInfo.IsLinux() {
		t.Skip("Test only runs on Linux")
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

	// Determine expected directory
	var expectedDir string
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		expectedDir = filepath.Join(xdgCache, "lazynuget")
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("os.UserHomeDir() failed: %v", err)
		}
		expectedDir = filepath.Join(homeDir, ".cache", "lazynuget")
	}

	if cacheDir != expectedDir {
		t.Errorf("CacheDir() = %q, want %q", cacheDir, expectedDir)
	}

	// Verify path is absolute
	if !filepath.IsAbs(cacheDir) {
		t.Errorf("CacheDir() returned relative path: %q", cacheDir)
	}
}

// TestLinuxXDGConfigHomeOverride verifies that XDG_CONFIG_HOME is respected
func TestLinuxXDGConfigHomeOverride(t *testing.T) {
	// Get platform instance
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	// Verify we're on Linux
	if !platformInfo.IsLinux() {
		t.Skip("Test only runs on Linux")
	}

	// Save original XDG_CONFIG_HOME
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if origXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	// Set custom XDG_CONFIG_HOME
	customConfigHome := "/tmp/custom-config"
	os.Setenv("XDG_CONFIG_HOME", customConfigHome)

	// Create PathResolver
	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get config directory
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	expectedDir := filepath.Join(customConfigHome, "lazynuget")
	if configDir != expectedDir {
		t.Errorf("ConfigDir() with XDG_CONFIG_HOME = %q, want %q", configDir, expectedDir)
	}
}
