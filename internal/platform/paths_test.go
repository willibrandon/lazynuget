package platform

import (
	"testing"
)

// TestConfigDir tests ConfigDir() method
// This test is platform-agnostic and will work on any supported platform
func TestConfigDir(t *testing.T) {
	// Get platform instance
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create PathResolver (will fail until implemented)
	// This test is expected to FAIL initially (TDD approach)
	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get config directory
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	// Verify result is not empty
	if configDir == "" {
		t.Error("ConfigDir() returned empty string")
	}

	// Verify result is absolute path
	if !pathResolver.IsAbsolute(configDir) {
		t.Errorf("ConfigDir() returned relative path: %q", configDir)
	}

	// Verify path contains "lazynuget"
	normalized := pathResolver.Normalize(configDir)
	if !contains(normalized, "lazynuget") {
		t.Errorf("ConfigDir() = %q does not contain 'lazynuget'", configDir)
	}
}

// TestCacheDir tests CacheDir() method
// This test is platform-agnostic and will work on any supported platform
func TestCacheDir(t *testing.T) {
	// Get platform instance
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create PathResolver
	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get cache directory
	cacheDir, err := pathResolver.CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() failed: %v", err)
	}

	// Verify result is not empty
	if cacheDir == "" {
		t.Error("CacheDir() returned empty string")
	}

	// Verify result is absolute path
	if !pathResolver.IsAbsolute(cacheDir) {
		t.Errorf("CacheDir() returned relative path: %q", cacheDir)
	}

	// Verify path contains "lazynuget"
	normalized := pathResolver.Normalize(cacheDir)
	if !contains(normalized, "lazynuget") {
		t.Errorf("CacheDir() = %q does not contain 'lazynuget'", cacheDir)
	}
}

// TestConfigDirAndCacheDirAreDifferent verifies that config and cache directories are different
func TestConfigDirAndCacheDirAreDifferent(t *testing.T) {
	// Get platform instance
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Create PathResolver
	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get both directories
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	cacheDir, err := pathResolver.CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() failed: %v", err)
	}

	// Verify they're different
	if configDir == cacheDir {
		t.Errorf("ConfigDir() and CacheDir() returned same path: %q", configDir)
	}
}

// contains is a simple substring check helper
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr))))
}
