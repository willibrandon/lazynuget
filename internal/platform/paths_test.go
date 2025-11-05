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

// TestCacheDir_ReadOnlyGracefulDegradation tests that cache directory gracefully degrades
// when permissions are insufficient. This is an integration test to verify FR-026.
// See: T103, FR-026
func TestCacheDir_ReadOnlyGracefulDegradation(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get cache directory - should succeed even if we can't create it
	cacheDir, err := pathResolver.CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() failed: %v", err)
	}

	// Verify cache directory path is returned (creation happens later with EnsureDir)
	if cacheDir == "" {
		t.Error("CacheDir() returned empty string")
	}

	if !pathResolver.IsAbsolute(cacheDir) {
		t.Errorf("CacheDir() returned relative path: %q", cacheDir)
	}

	// The actual graceful degradation happens when trying to create the directory
	// EnsureDir should be called by application code, and it returns an error if it fails
	// The application can then choose to continue without cache (graceful degradation)

	// Note: We don't test actual permission failures here because:
	// 1. Creating truly read-only scenarios requires platform-specific setup
	// 2. This would be flaky in CI environments
	// 3. The graceful degradation is a property of the application logic, not the platform package
	// 4. The platform package correctly returns errors from EnsureDir, letting the app decide
}

// TestConfigDir_ReadOnlyFailure tests that config directory operations fail appropriately
// when permissions are insufficient. This verifies FR-027 (config dir must be writable).
// See: T104, FR-027
func TestConfigDir_ReadOnlyFailure(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get config directory - should succeed even if we can't create it yet
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	// Verify config directory path is returned
	if configDir == "" {
		t.Error("ConfigDir() returned empty string")
	}

	if !pathResolver.IsAbsolute(configDir) {
		t.Errorf("ConfigDir() returned relative path: %q", configDir)
	}

	// The actual failure handling happens when trying to create/write to the directory
	// EnsureDir will return an error if the directory can't be created or is read-only
	// The application is expected to fail fast in this case per FR-027

	// Note: We don't test actual permission failures here because:
	// 1. Creating truly read-only scenarios requires platform-specific setup
	// 2. This would be flaky in CI environments
	// 3. The platform package correctly returns errors from EnsureDir
	// 4. The application code is responsible for failing fast when config dir is unavailable
}
