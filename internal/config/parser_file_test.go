package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectFormat tests config file format detection
func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name string
		path string
		want ConfigFormat
	}{
		{
			name: "detect YAML from .yml extension",
			path: "/path/to/config.yml",
			want: FormatYAML,
		},
		{
			name: "detect YAML from .yaml extension",
			path: "/path/to/config.yaml",
			want: FormatYAML,
		},
		{
			name: "detect TOML from .toml extension",
			path: "/path/to/config.toml",
			want: FormatTOML,
		},
		{
			name: "unsupported extension returns FormatUnknown",
			path: "/path/to/config.json",
			want: FormatUnknown,
		},
		{
			name: "no extension returns FormatUnknown",
			path: "/path/to/config",
			want: FormatUnknown,
		},
		{
			name: "uppercase extension detected",
			path: "/path/to/config.YML",
			want: FormatYAML,
		},
		{
			name: "mixed case extension detected",
			path: "/path/to/config.Toml",
			want: FormatTOML,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectFormat(tt.path)

			if got != tt.want {
				t.Errorf("detectFormat(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

// TestValidateFileSize tests file size validation
func TestValidateFileSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int64
		wantErr bool
	}{
		{
			name:    "empty file",
			size:    0,
			wantErr: false,
		},
		{
			name:    "small file 1KB",
			size:    1024,
			wantErr: false,
		},
		{
			name:    "medium file 1MB",
			size:    1024 * 1024,
			wantErr: false,
		},
		{
			name:    "large file 9MB",
			size:    9 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "max size 10MB",
			size:    10 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "exceeds max size 11MB",
			size:    11 * 1024 * 1024,
			wantErr: true,
		},
		{
			name:    "far exceeds max size 100MB",
			size:    100 * 1024 * 1024,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file of specified size
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.yml")

			// Create file with specified size
			f, err := os.Create(tmpFile)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			if tt.size > 0 {
				if err := f.Truncate(tt.size); err != nil {
					f.Close()
					t.Fatalf("Failed to set file size: %v", err)
				}
			}
			f.Close()

			// Test validation
			err = validateFileSize(tmpFile)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestCheckMultipleFormats tests detection of multiple config file formats
func TestCheckMultipleFormats(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		wantErr   bool
		errContains string
	}{
		{
			name:    "no config files",
			files:   []string{},
			wantErr: false,
		},
		{
			name:    "only YAML file",
			files:   []string{"config.yml"},
			wantErr: false,
		},
		{
			name:    "only TOML file",
			files:   []string{"config.toml"},
			wantErr: false,
		},
		{
			name:    "non-config files",
			files:   []string{"readme.txt", "data.json"},
			wantErr: false,
		},
		{
			name:        "both YAML and TOML",
			files:       []string{"config.yml", "config.toml"},
			wantErr:     true,
			errContains: "both YAML and TOML",
		},
		{
			name:    "multiple YAML variants allowed",
			files:   []string{"config.yml", "config.yaml"},
			wantErr: false, // Both treated as YAML, not an error
		},
		{
			name:    "YAML with other files",
			files:   []string{"config.yml", "readme.md", "data.json"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test files
			for _, filename := range tt.files {
				path := filepath.Join(tmpDir, filename)
				if err := os.WriteFile(path, []byte("test content"), 0600); err != nil {
					t.Fatalf("Failed to create test file %s: %v", filename, err)
				}
			}

			err := checkMultipleFormats(tmpDir)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestParseConfigFile tests the main config file parsing function
func TestParseConfigFile(t *testing.T) {
	tests := []struct {
		name      string
		extension string
		content   string
		checkFunc func(*Config) error
		wantErr   bool
	}{
		{
			name:      "parse valid YAML file",
			extension: ".yml",
			content: `
logLevel: debug
theme: dark
maxConcurrentOps: 8
`,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					return &assertError{msg: "Expected LogLevel=debug"}
				}
				if cfg.Theme != "dark" {
					return &assertError{msg: "Expected Theme=dark"}
				}
				if cfg.MaxConcurrentOps != 8 {
					return &assertError{msg: "Expected MaxConcurrentOps=8"}
				}
				return nil
			},
		},
		{
			name:      "parse valid TOML file",
			extension: ".toml",
			content: `
log_level = "error"
theme = "light"
max_concurrent_ops = 6
`,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "error" {
					return &assertError{msg: "Expected LogLevel=error"}
				}
				if cfg.Theme != "light" {
					return &assertError{msg: "Expected Theme=light"}
				}
				if cfg.MaxConcurrentOps != 6 {
					return &assertError{msg: "Expected MaxConcurrentOps=6"}
				}
				return nil
			},
		},
		{
			name:      "parse YAML with nested structures",
			extension: ".yml",
			content: `
logLevel: info
colorScheme:
  border: "#CUSTOM"
  error: "#FF0000"
timeouts:
  networkRequest: 60s
`,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "info" {
					return &assertError{msg: "Expected LogLevel=info"}
				}
				if cfg.ColorScheme.Border != "#CUSTOM" {
					return &assertError{msg: "Expected Border=#CUSTOM"}
				}
				if cfg.ColorScheme.Error != "#FF0000" {
					return &assertError{msg: "Expected Error=#FF0000"}
				}
				return nil
			},
		},
		{
			name:      "parse TOML with nested structures",
			extension: ".toml",
			content: `
log_level = "warn"

[color_scheme]
border = "#123456"
error = "#ABCDEF"
`,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "warn" {
					return &assertError{msg: "Expected LogLevel=warn"}
				}
				if cfg.ColorScheme.Border != "#123456" {
					return &assertError{msg: "Expected Border=#123456"}
				}
				return nil
			},
		},
		{
			name:      "invalid YAML syntax",
			extension: ".yml",
			content: `
logLevel: "unclosed quote
theme: dark
`,
			wantErr: true,
		},
		{
			name:      "invalid TOML syntax",
			extension: ".toml",
			content: `
log_level = missing quotes
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config"+tt.extension)

			if err := os.WriteFile(tmpFile, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			cfg, err := parseConfigFile(tmpFile)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if cfg == nil {
				t.Fatal("parseConfigFile returned nil config")
			}

			if tt.checkFunc != nil {
				if err := tt.checkFunc(cfg); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestParseConfigFileErrors tests error cases for parseConfigFile
func TestParseConfigFileErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() string
		errContains string
	}{
		{
			name: "file does not exist",
			setupFunc: func() string {
				return "/nonexistent/path/config.yml"
			},
			errContains: "",
		},
		{
			name: "unsupported file format",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "config.json")
				os.WriteFile(tmpFile, []byte(`{"logLevel": "debug"}`), 0600)
				return tmpFile
			},
			errContains: "unsupported",
		},
		{
			name: "file too large",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "config.yml")
				f, _ := os.Create(tmpFile)
				// Create 11MB file (exceeds 10MB limit)
				f.Truncate(11 * 1024 * 1024)
				f.Close()
				return tmpFile
			},
			errContains: "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFunc()
			_, err := parseConfigFile(path)

			if err == nil {
				t.Error("Expected error but got none")
				return
			}

			if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
			}
		})
	}
}

// TestValidateConfigFilePath tests config file path validation
func TestValidateConfigFilePath(t *testing.T) {
	// Create a temp file for valid path tests
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(validFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid file path",
			path:    validFile,
			wantErr: false,
		},
		{
			name:    "file does not exist",
			path:    "/nonexistent/path/config.yml",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfigFilePath(tt.path)

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
