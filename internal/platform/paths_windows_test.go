//go:build windows

package platform

import (
	"os"
	"strings"
	"testing"
)

// TestNormalize_WindowsDriveLetter tests that drive letters are normalized to uppercase
func TestNormalize_WindowsDriveLetter(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase drive letter",
			input:    "c:\\users\\test\\file.txt",
			expected: "C:\\users\\test\\file.txt",
		},
		{
			name:     "uppercase drive letter unchanged",
			input:    "D:\\Projects\\LazyNuGet",
			expected: "D:\\Projects\\LazyNuGet",
		},
		{
			name:     "mixed case drive letter",
			input:    "e:\\temp\\config.yml",
			expected: "E:\\temp\\config.yml",
		},
		{
			name:     "forward slashes converted to backslashes",
			input:    "c:/users/test/file.txt",
			expected: "C:\\users\\test\\file.txt",
		},
		{
			name:     "mixed separators",
			input:    "c:/users\\test/file.txt",
			expected: "C:\\users\\test\\file.txt",
		},
		{
			name:     "redundant separators removed",
			input:    "C:\\\\users\\\\test\\\\file.txt",
			expected: "C:\\users\\test\\file.txt",
		},
		{
			name:     "dot segments resolved",
			input:    "C:\\users\\test\\.\\file.txt",
			expected: "C:\\users\\test\\file.txt",
		},
		{
			name:     "dotdot segments resolved",
			input:    "C:\\users\\test\\..\\other\\file.txt",
			expected: "C:\\users\\other\\file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathResolver.Normalize(tt.input)
			if got != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestNormalize_WindowsUNCPaths tests UNC path handling
func TestNormalize_WindowsUNCPaths(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic UNC path",
			input:    "\\\\server\\share\\file.txt",
			expected: "\\\\server\\share\\file.txt",
		},
		{
			name:     "UNC path with forward slashes",
			input:    "//server/share/file.txt",
			expected: "\\\\server\\share\\file.txt",
		},
		{
			name:     "UNC path with mixed separators",
			input:    "\\\\server/share\\file.txt",
			expected: "\\\\server\\share\\file.txt",
		},
		{
			name:     "UNC path with redundant separators",
			input:    "\\\\\\\\server\\\\share\\\\file.txt",
			expected: "\\\\server\\share\\file.txt",
		},
		{
			name:     "UNC path with dot segments",
			input:    "\\\\server\\share\\.\\file.txt",
			expected: "\\\\server\\share\\file.txt",
		},
		{
			name:     "UNC path with dotdot segments",
			input:    "\\\\server\\share\\dir\\..\\file.txt",
			expected: "\\\\server\\share\\file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathResolver.Normalize(tt.input)
			if got != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// TestIsAbsolute_Windows tests absolute path detection on Windows
func TestIsAbsolute_Windows(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "drive letter absolute",
			path:     "C:\\Users\\test",
			expected: true,
		},
		{
			name:     "lowercase drive letter absolute",
			path:     "c:\\users\\test",
			expected: true,
		},
		{
			name:     "UNC path absolute",
			path:     "\\\\server\\share\\file.txt",
			expected: true,
		},
		{
			name:     "relative path",
			path:     "relative\\path\\file.txt",
			expected: false,
		},
		{
			name:     "current directory relative",
			path:     ".\\file.txt",
			expected: false,
		},
		{
			name:     "parent directory relative",
			path:     "..\\file.txt",
			expected: false,
		},
		{
			name:     "drive-relative path (weird Windows thing)",
			path:     "C:file.txt",
			expected: false,
		},
		{
			name:     "rooted but not absolute (no drive)",
			path:     "\\file.txt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathResolver.IsAbsolute(tt.path)
			if got != tt.expected {
				t.Errorf("IsAbsolute(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

// TestValidate_Windows tests path validation on Windows
func TestValidate_Windows(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid absolute path",
			path:    "C:\\Users\\test\\file.txt",
			wantErr: false,
		},
		{
			name:    "valid UNC path",
			path:    "\\\\server\\share\\file.txt",
			wantErr: false,
		},
		{
			name:    "valid relative path",
			path:    "relative\\path\\file.txt",
			wantErr: false,
		},
		{
			name:    "empty path invalid",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path with invalid characters",
			path:    "C:\\users\\test\\<invalid>.txt",
			wantErr: true,
		},
		{
			name:    "path with question mark",
			path:    "C:\\users\\test\\file?.txt",
			wantErr: true,
		},
		{
			name:    "path with asterisk",
			path:    "C:\\users\\test\\*.txt",
			wantErr: true,
		},
		{
			name:    "path with pipe",
			path:    "C:\\users\\test\\file|.txt",
			wantErr: true,
		},
		{
			name:    "path with quote",
			path:    "C:\\users\\test\\\"file\".txt",
			wantErr: true,
		},
		{
			name:    "reserved device name CON",
			path:    "C:\\users\\CON",
			wantErr: true,
		},
		{
			name:    "reserved device name PRN",
			path:    "C:\\users\\PRN",
			wantErr: true,
		},
		{
			name:    "reserved device name AUX",
			path:    "C:\\users\\AUX",
			wantErr: true,
		},
		{
			name:    "reserved device name NUL",
			path:    "C:\\users\\NUL",
			wantErr: true,
		},
		{
			name:    "reserved device name COM1",
			path:    "C:\\users\\COM1",
			wantErr: true,
		},
		{
			name:    "reserved device name LPT1",
			path:    "C:\\users\\LPT1",
			wantErr: true,
		},
		{
			name:    "path ending with space",
			path:    "C:\\users\\test\\file.txt ",
			wantErr: true,
		},
		{
			name:    "path ending with period",
			path:    "C:\\users\\test\\file.txt.",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pathResolver.Validate(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

// TestResolve_Windows tests path resolution on Windows
func TestResolve_Windows(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Get config directory for relative path resolution
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		shouldStart string // prefix the result should have
	}{
		{
			name:        "absolute path unchanged",
			path:        "C:\\Users\\test\\file.txt",
			shouldStart: "C:\\Users\\test\\file.txt",
		},
		{
			name:        "UNC path unchanged",
			path:        "\\\\server\\share\\file.txt",
			shouldStart: "\\\\server\\share\\file.txt",
		},
		{
			name:        "relative path resolved to config dir",
			path:        "config.yml",
			shouldStart: configDir,
		},
		{
			name:        "relative path with subdirectory",
			path:        "subdir\\config.yml",
			shouldStart: configDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pathResolver.Resolve(tt.path)
			if err != nil {
				t.Errorf("Resolve(%q) error = %v", tt.path, err)
				return
			}
			if len(got) < len(tt.shouldStart) || got[:len(tt.shouldStart)] != tt.shouldStart {
				t.Errorf("Resolve(%q) = %q, should start with %q", tt.path, got, tt.shouldStart)
			}
		})
	}
}

// TestEnvVarPrecedence_Windows tests that APPDATA takes precedence over XDG variables
// This is important for WSL scenarios where XDG variables might be set
// See: T102, FR-026
func TestEnvVarPrecedence_Windows(t *testing.T) {
	// Save original environment variables
	origAPPDATA := os.Getenv("APPDATA")
	origLOCALAPPDATA := os.Getenv("LOCALAPPDATA")
	origXDGCONFIG := os.Getenv("XDG_CONFIG_HOME")
	origXDGCACHE := os.Getenv("XDG_CACHE_HOME")

	// Restore after test
	defer func() {
		os.Setenv("APPDATA", origAPPDATA)
		os.Setenv("LOCALAPPDATA", origLOCALAPPDATA)
		os.Setenv("XDG_CONFIG_HOME", origXDGCONFIG)
		os.Setenv("XDG_CACHE_HOME", origXDGCACHE)
	}()

	// Set up test environment: Windows vars + XDG vars (simulating WSL)
	os.Setenv("APPDATA", "C:\\Users\\Test\\AppData\\Roaming")
	os.Setenv("LOCALAPPDATA", "C:\\Users\\Test\\AppData\\Local")
	os.Setenv("XDG_CONFIG_HOME", "/home/user/.config")
	os.Setenv("XDG_CACHE_HOME", "/home/user/.cache")

	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// Test ConfigDir: should use APPDATA, NOT XDG_CONFIG_HOME
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		t.Fatalf("ConfigDir() failed: %v", err)
	}

	expectedConfigPrefix := "C:\\Users\\Test\\AppData\\Roaming"
	if !strings.HasPrefix(configDir, expectedConfigPrefix) {
		t.Errorf("ConfigDir() = %q, should start with %q (APPDATA), not XDG_CONFIG_HOME", configDir, expectedConfigPrefix)
	}

	if strings.Contains(configDir, "/home/user/.config") {
		t.Errorf("ConfigDir() = %q incorrectly used XDG_CONFIG_HOME instead of APPDATA", configDir)
	}

	// Test CacheDir: should use LOCALAPPDATA, NOT XDG_CACHE_HOME
	cacheDir, err := pathResolver.CacheDir()
	if err != nil {
		t.Fatalf("CacheDir() failed: %v", err)
	}

	expectedCachePrefix := "C:\\Users\\Test\\AppData\\Local"
	if !strings.HasPrefix(cacheDir, expectedCachePrefix) {
		t.Errorf("CacheDir() = %q, should start with %q (LOCALAPPDATA), not XDG_CACHE_HOME", cacheDir, expectedCachePrefix)
	}

	if strings.Contains(cacheDir, "/home/user/.cache") {
		t.Errorf("CacheDir() = %q incorrectly used XDG_CACHE_HOME instead of LOCALAPPDATA", cacheDir)
	}
}
