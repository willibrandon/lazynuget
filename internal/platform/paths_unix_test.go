//go:build !windows

package platform

import (
	"testing"
)

// TestNormalize_Unix tests path normalization on Unix systems
func TestNormalize_Unix(t *testing.T) {
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
			name:     "simple absolute path unchanged",
			input:    "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "redundant separators removed",
			input:    "/usr//local///bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "dot segments resolved",
			input:    "/usr/./local/./bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "dotdot segments resolved",
			input:    "/usr/local/share/../bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "multiple dotdot segments",
			input:    "/usr/local/share/../../bin",
			expected: "/usr/bin",
		},
		{
			name:     "trailing slash removed",
			input:    "/usr/local/bin/",
			expected: "/usr/local/bin",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path/file.txt",
			expected: "relative/path/file.txt",
		},
		{
			name:     "current directory path",
			input:    "./file.txt",
			expected: "file.txt",
		},
		{
			name:     "parent directory path",
			input:    "../file.txt",
			expected: "../file.txt",
		},
		{
			name:     "home directory tilde NOT expanded",
			input:    "~/config/file.txt",
			expected: "~/config/file.txt",
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

// TestIsAbsolute_Unix tests absolute path detection on Unix systems
func TestIsAbsolute_Unix(t *testing.T) {
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
			name:     "root path absolute",
			path:     "/",
			expected: true,
		},
		{
			name:     "standard absolute path",
			path:     "/usr/local/bin",
			expected: true,
		},
		{
			name:     "absolute path with trailing slash",
			path:     "/usr/local/bin/",
			expected: true,
		},
		{
			name:     "relative path",
			path:     "relative/path",
			expected: false,
		},
		{
			name:     "current directory relative",
			path:     "./file.txt",
			expected: false,
		},
		{
			name:     "parent directory relative",
			path:     "../file.txt",
			expected: false,
		},
		{
			name:     "home directory tilde relative",
			path:     "~/config/file.txt",
			expected: false,
		},
		{
			name:     "simple filename relative",
			path:     "file.txt",
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

// TestValidate_Unix tests path validation on Unix systems
func TestValidate_Unix(t *testing.T) {
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
			path:    "/usr/local/bin",
			wantErr: false,
		},
		{
			name:    "valid relative path",
			path:    "relative/path/file.txt",
			wantErr: false,
		},
		{
			name:    "valid path with spaces",
			path:    "/usr/local/My Documents/file.txt",
			wantErr: false,
		},
		{
			name:    "valid path with special chars",
			path:    "/usr/local/file-name_123.txt",
			wantErr: false,
		},
		{
			name:    "empty path invalid",
			path:    "",
			wantErr: true,
		},
		{
			name:    "path with null byte invalid",
			path:    "/usr/local/\x00file.txt",
			wantErr: true,
		},
		{
			name:    "valid hidden file",
			path:    "/home/user/.config",
			wantErr: false,
		},
		{
			name:    "valid path with dots in filename",
			path:    "/usr/local/file.tar.gz",
			wantErr: false,
		},
		{
			name:    "valid path with unicode",
			path:    "/usr/local/文件.txt",
			wantErr: false,
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

// TestResolve_Unix tests path resolution on Unix systems
func TestResolve_Unix(t *testing.T) {
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
			path:        "/usr/local/bin/lazynuget",
			shouldStart: "/usr/local/bin/lazynuget",
		},
		{
			name:        "root path unchanged",
			path:        "/",
			shouldStart: "/",
		},
		{
			name:        "relative path resolved to config dir",
			path:        "config.yml",
			shouldStart: configDir,
		},
		{
			name:        "relative path with subdirectory",
			path:        "subdir/config.yml",
			shouldStart: configDir,
		},
		{
			name:        "relative path with dot",
			path:        "./config.yml",
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

// TestSymlinksHandling tests that symlinks are handled gracefully
func TestSymlinksHandling(t *testing.T) {
	platformInfo, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	pathResolver, err := NewPathResolver(platformInfo)
	if err != nil {
		t.Fatalf("NewPathResolver() failed: %v", err)
	}

	// These are common system symlinks that should be valid paths
	// We don't test actual symlink resolution, just that they validate
	tests := []struct {
		name string
		path string
	}{
		{
			name: "tmp symlink",
			path: "/tmp",
		},
		{
			name: "var symlink",
			path: "/var",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should validate without error
			err := pathResolver.Validate(tt.path)
			if err != nil {
				t.Errorf("Validate(%q) error = %v, should accept symlink paths", tt.path, err)
			}

			// Should recognize as absolute
			if !pathResolver.IsAbsolute(tt.path) {
				t.Errorf("IsAbsolute(%q) = false, symlink paths should be recognized as absolute", tt.path)
			}
		})
	}
}
