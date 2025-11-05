//go:build darwin

package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// TestDarwinConfigDirectory verifies that on macOS, the config directory
// resolves to ~/Library/Application Support/lazynuget
func TestDarwinConfigDirectory(t *testing.T) {
	// Get platform instance
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	// Verify we're on macOS
	if !platformInfo.IsDarwin() {
		t.Skip("Test only runs on macOS")
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

	// Verify it's under ~/Library/Application Support
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}

	expectedDir := filepath.Join(homeDir, "Library", "Application Support", "lazynuget")
	if configDir != expectedDir {
		t.Errorf("ConfigDir() = %q, want %q", configDir, expectedDir)
	}

	// Verify path uses Unix separators (forward slashes)
	if !filepath.IsAbs(configDir) {
		t.Errorf("ConfigDir() returned relative path: %q", configDir)
	}
}

// TestDarwinCacheDirectory verifies that on macOS, the cache directory
// resolves to ~/Library/Caches/lazynuget
func TestDarwinCacheDirectory(t *testing.T) {
	// Get platform instance
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	// Verify we're on macOS
	if !platformInfo.IsDarwin() {
		t.Skip("Test only runs on macOS")
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

	// Verify it's under ~/Library/Caches
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() failed: %v", err)
	}

	expectedDir := filepath.Join(homeDir, "Library", "Caches", "lazynuget")
	if cacheDir != expectedDir {
		t.Errorf("CacheDir() = %q, want %q", cacheDir, expectedDir)
	}

	// Verify path is absolute
	if !filepath.IsAbs(cacheDir) {
		t.Errorf("CacheDir() returned relative path: %q", cacheDir)
	}
}
