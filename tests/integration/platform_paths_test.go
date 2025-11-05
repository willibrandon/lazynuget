package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// TestPathNormalization_Integration tests path normalization works correctly across platforms
func TestPathNormalization_Integration(t *testing.T) {
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("platform.NewPathResolver() failed: %v", err)
	}

	// Test that Normalize produces platform-appropriate separators
	t.Run("normalize produces platform separators", func(t *testing.T) {
		// Use a relative path with mixed separators
		mixedPath := "some/path\\to\\file.txt"
		normalized := pathResolver.Normalize(mixedPath)

		// Verify the result uses the platform's separator
		expectedSep := string(filepath.Separator)

		// On Windows, should have backslashes; on Unix, forward slashes
		if runtime.GOOS == "windows" {
			if !containsChar(normalized, '\\') {
				t.Errorf("Normalize(%q) = %q, expected backslashes on Windows", mixedPath, normalized)
			}
			if containsChar(normalized, '/') {
				t.Errorf("Normalize(%q) = %q, should not contain forward slashes on Windows", mixedPath, normalized)
			}
		} else {
			if !containsChar(normalized, '/') {
				t.Errorf("Normalize(%q) = %q, expected forward slashes on Unix", mixedPath, normalized)
			}
			// Backslashes are valid filename characters on Unix, so we can't check for their absence
		}

		// Verify redundant separators are removed
		redundantPath := "path" + expectedSep + expectedSep + "to" + expectedSep + expectedSep + expectedSep + "file.txt"
		normalized = pathResolver.Normalize(redundantPath)
		expected := "path" + expectedSep + "to" + expectedSep + "file.txt"
		if normalized != expected {
			t.Errorf("Normalize(%q) = %q, want %q (redundant separators should be removed)", redundantPath, normalized, expected)
		}
	})

	// Test that ConfigDir returns a valid absolute path
	t.Run("configdir is absolute", func(t *testing.T) {
		configDir, err := pathResolver.ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() failed: %v", err)
		}

		if !pathResolver.IsAbsolute(configDir) {
			t.Errorf("ConfigDir() = %q is not absolute", configDir)
		}

		// Verify it's a valid path format for the platform
		if err := pathResolver.Validate(configDir); err != nil {
			t.Errorf("ConfigDir() returned invalid path: %v", err)
		}
	})

	// Test that CacheDir returns a valid absolute path
	t.Run("cachedir is absolute", func(t *testing.T) {
		cacheDir, err := pathResolver.CacheDir()
		if err != nil {
			t.Fatalf("CacheDir() failed: %v", err)
		}

		if !pathResolver.IsAbsolute(cacheDir) {
			t.Errorf("CacheDir() = %q is not absolute", cacheDir)
		}

		// Verify it's a valid path format for the platform
		if err := pathResolver.Validate(cacheDir); err != nil {
			t.Errorf("CacheDir() returned invalid path: %v", err)
		}
	})

	// Test Resolve with relative and absolute paths
	t.Run("resolve handles absolute and relative", func(t *testing.T) {
		configDir, err := pathResolver.ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() failed: %v", err)
		}

		// Absolute path should be unchanged
		resolved, err := pathResolver.Resolve(configDir)
		if err != nil {
			t.Fatalf("Resolve(%q) failed: %v", configDir, err)
		}
		if resolved != configDir {
			t.Errorf("Resolve(%q) = %q, absolute path should be unchanged", configDir, resolved)
		}

		// Relative path should be resolved to config directory
		relativePath := "config.yml"
		resolved, err = pathResolver.Resolve(relativePath)
		if err != nil {
			t.Fatalf("Resolve(%q) failed: %v", relativePath, err)
		}

		expectedPrefix := configDir
		if len(resolved) < len(expectedPrefix) || resolved[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("Resolve(%q) = %q, should start with config dir %q", relativePath, resolved, expectedPrefix)
		}

		// Resolved path should be absolute
		if !pathResolver.IsAbsolute(resolved) {
			t.Errorf("Resolve(%q) = %q is not absolute", relativePath, resolved)
		}
	})

	// Test EnsureDir creates and validates directories
	t.Run("ensuredir creates directory", func(t *testing.T) {
		// Create a test directory in temp
		tempDir := filepath.Join(os.TempDir(), "lazynuget-test-ensuredir")
		defer os.RemoveAll(tempDir) // Clean up

		// Remove if it exists from a previous test
		os.RemoveAll(tempDir)

		// EnsureDir should create it
		err := pathResolver.EnsureDir(tempDir)
		if err != nil {
			t.Fatalf("EnsureDir(%q) failed: %v", tempDir, err)
		}

		// Verify it exists and is a directory
		info, err := os.Stat(tempDir)
		if err != nil {
			t.Fatalf("Stat(%q) failed after EnsureDir: %v", tempDir, err)
		}
		if !info.IsDir() {
			t.Errorf("EnsureDir(%q) created non-directory", tempDir)
		}

		// Calling EnsureDir again should not error
		err = pathResolver.EnsureDir(tempDir)
		if err != nil {
			t.Errorf("EnsureDir(%q) failed on existing directory: %v", tempDir, err)
		}
	})

	// Test EnsureDir fails on file path
	t.Run("ensuredir fails on file", func(t *testing.T) {
		// Create a test file in temp
		tempFile := filepath.Join(os.TempDir(), "lazynuget-test-file.txt")
		defer os.Remove(tempFile) // Clean up

		if err := os.WriteFile(tempFile, []byte("test"), 0o600); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}

		// EnsureDir should fail because path is a file, not directory
		err := pathResolver.EnsureDir(tempFile)
		if err == nil {
			t.Error("EnsureDir() should fail when path is a file")
		}
	})
}

// TestPathValidation_Integration tests validation across platforms
func TestPathValidation_Integration(t *testing.T) {
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("platform.NewPathResolver() failed: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		onlyOnOS   string // empty means all platforms
		shouldFail bool
	}{
		{
			name:       "empty path always invalid",
			path:       "",
			shouldFail: true,
		},
		{
			name:       "simple filename valid",
			path:       "file.txt",
			shouldFail: false,
		},
		{
			name:       "relative path valid",
			path:       "some/path/to/file.txt",
			shouldFail: false,
		},
		{
			name:       "windows drive letter valid on windows",
			path:       "C:\\Users\\test",
			shouldFail: runtime.GOOS != "windows",
			onlyOnOS:   "windows",
		},
		{
			name:       "unix absolute path valid on unix",
			path:       "/usr/local/bin",
			shouldFail: runtime.GOOS == "windows",
			onlyOnOS:   "",
		},
		{
			name:       "null byte always invalid",
			path:       "file\x00name.txt",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		// Skip tests that are OS-specific
		if tt.onlyOnOS != "" && tt.onlyOnOS != runtime.GOOS {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pathResolver.Validate(tt.path)
			if tt.shouldFail && err == nil {
				t.Errorf("Validate(%q) should fail but didn't", tt.path)
			}
			if !tt.shouldFail && err != nil {
				t.Errorf("Validate(%q) should succeed but failed: %v", tt.path, err)
			}
		})
	}
}

// TestPlatformSpecificPaths_Integration tests platform-specific path behavior
func TestPlatformSpecificPaths_Integration(t *testing.T) {
	platformInfo, err := platform.New()
	if err != nil {
		t.Fatalf("platform.New() failed: %v", err)
	}

	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("platform.NewPathResolver() failed: %v", err)
	}

	switch runtime.GOOS {
	case "windows":
		t.Run("windows drive letters", func(t *testing.T) {
			// Test various drive letter formats
			paths := []string{
				"C:\\",
				"c:\\",
				"D:\\Users",
				"E:/Windows", // Mixed separators
			}

			for _, p := range paths {
				if !pathResolver.IsAbsolute(p) {
					t.Errorf("IsAbsolute(%q) = false, should be true on Windows", p)
				}
			}
		})

		t.Run("windows UNC paths", func(t *testing.T) {
			uncPath := "\\\\server\\share"
			if !pathResolver.IsAbsolute(uncPath) {
				t.Errorf("IsAbsolute(%q) = false, UNC paths should be absolute", uncPath)
			}
		})

	case "darwin", "linux":
		t.Run("unix root path", func(t *testing.T) {
			rootPath := "/"
			if !pathResolver.IsAbsolute(rootPath) {
				t.Errorf("IsAbsolute(%q) = false, root should be absolute", rootPath)
			}
		})

		t.Run("unix home relative", func(t *testing.T) {
			homePath := "~/config"
			if pathResolver.IsAbsolute(homePath) {
				t.Errorf("IsAbsolute(%q) = true, tilde paths are relative", homePath)
			}
		})
	}
}

// containsChar checks if string contains a specific character
func containsChar(s string, ch rune) bool {
	for _, c := range s {
		if c == ch {
			return true
		}
	}
	return false
}
