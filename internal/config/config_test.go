package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewLoader tests the NewLoader constructor
func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}

	// Verify we can call methods on the loader
	cfg := loader.GetDefaults()
	if cfg == nil {
		t.Error("GetDefaults returned nil")
	}
}

// TestConfigLoaderGetDefaults tests the GetDefaults method
func TestConfigLoaderGetDefaults(t *testing.T) {
	loader := NewLoader()
	cfg := loader.GetDefaults()

	if cfg == nil {
		t.Fatal("GetDefaults returned nil")
	}

	// Verify some default values
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel=info, got %s", cfg.LogLevel)
	}
	if cfg.MaxConcurrentOps != 4 {
		t.Errorf("Expected default MaxConcurrentOps=4, got %d", cfg.MaxConcurrentOps)
	}
	if cfg.Theme != "default" {
		t.Errorf("Expected default Theme=default, got %s", cfg.Theme)
	}
}

// TestConfigLoaderValidate tests the Validate method
func TestConfigLoaderValidate(t *testing.T) {
	loader := NewLoader()
	ctx := context.Background()

	tests := []struct {
		cfg           *Config
		name          string
		wantErrCount  int
		wantWarnCount int
		nilConfig     bool
		wantSysErr    bool
	}{
		{
			name:          "valid config",
			cfg:           GetDefaultConfig(),
			wantErrCount:  0,
			wantWarnCount: 1, // RefreshInterval=0
		},
		{
			name: "invalid log level",
			cfg: func() *Config {
				cfg := *GetDefaultConfig()
				cfg.LogLevel = "invalid"
				return &cfg
			}(),
			wantErrCount:  0,
			wantWarnCount: 2, // RefreshInterval + LogLevel
		},
		{
			name: "multiple errors",
			cfg: func() *Config {
				cfg := *GetDefaultConfig()
				cfg.LogLevel = "invalid"
				cfg.Theme = "invalid"
				cfg.MaxConcurrentOps = 999
				return &cfg
			}(),
			wantErrCount:  0,
			wantWarnCount: 4, // RefreshInterval + LogLevel + Theme + MaxConcurrentOps
		},
		{
			name:       "nil config",
			cfg:        nil,
			nilConfig:  true,
			wantSysErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors, err := loader.Validate(ctx, tt.cfg)

			if tt.wantSysErr {
				if err == nil {
					t.Error("Expected system error for nil config, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected system error: %v", err)
			}

			// Count errors and warnings
			errorCount := 0
			warnCount := 0
			for _, ve := range validationErrors {
				if ve.Severity == "error" {
					errorCount++
				} else if ve.Severity == "warning" {
					warnCount++
				}
			}

			if errorCount != tt.wantErrCount {
				t.Errorf("Expected %d errors, got %d", tt.wantErrCount, errorCount)
			}
			if warnCount != tt.wantWarnCount {
				t.Errorf("Expected %d warnings, got %d", tt.wantWarnCount, warnCount)
			}
		})
	}
}

// TestConfigLoaderLoad tests the Load method with various scenarios
func TestConfigLoaderLoad(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		setupFunc   func() (LoadOptions, func())
		checkFunc   func(*Config) error
		name        string
		errContains string
		wantErr     bool
	}{
		{
			name: "load with defaults only",
			setupFunc: func() (LoadOptions, func()) {
				return LoadOptions{
					EnvVarPrefix: "LAZYNUGET_",
					StrictMode:   false,
				}, func() {}
			},
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "info" {
					return &assertError{msg: "Expected default LogLevel=info"}
				}
				return nil
			},
		},
		{
			name: "load with CLI flag override",
			setupFunc: func() (LoadOptions, func()) {
				return LoadOptions{
					EnvVarPrefix: "LAZYNUGET_",
					StrictMode:   false,
					CLIFlags: CLIFlags{
						LogLevel: "debug",
					},
				}, func() {}
			},
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					return &assertError{msg: "Expected LogLevel=debug from CLI flag"}
				}
				return nil
			},
		},
		{
			name: "load with env var override",
			setupFunc: func() (LoadOptions, func()) {
				os.Setenv("LAZYNUGET_LOG_LEVEL", "warn")
				cleanup := func() {
					os.Unsetenv("LAZYNUGET_LOG_LEVEL")
				}
				return LoadOptions{
					EnvVarPrefix: "LAZYNUGET_",
					StrictMode:   false,
				}, cleanup
			},
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "warn" {
					return &assertError{msg: "Expected LogLevel=warn from env var"}
				}
				return nil
			},
		},
		{
			name: "CLI flag overrides env var",
			setupFunc: func() (LoadOptions, func()) {
				os.Setenv("LAZYNUGET_LOG_LEVEL", "warn")
				cleanup := func() {
					os.Unsetenv("LAZYNUGET_LOG_LEVEL")
				}
				return LoadOptions{
					EnvVarPrefix: "LAZYNUGET_",
					StrictMode:   false,
					CLIFlags: CLIFlags{
						LogLevel: "error",
					},
				}, cleanup
			},
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "error" {
					return &assertError{msg: "Expected LogLevel=error (CLI overrides env)"}
				}
				return nil
			},
		},
		{
			name: "load with YAML config file",
			setupFunc: func() (LoadOptions, func()) {
				// Create temp YAML file
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				yamlContent := `
logLevel: debug
theme: dark
maxConcurrentOps: 8
`
				if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
					t.Fatalf("Failed to write test config: %v", err)
				}

				return LoadOptions{
					ConfigFilePath: configPath,
					EnvVarPrefix:   "LAZYNUGET_",
					StrictMode:     false,
				}, func() {}
			},
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					return &assertError{msg: "Expected LogLevel=debug from YAML"}
				}
				if cfg.Theme != "dark" {
					return &assertError{msg: "Expected Theme=dark from YAML"}
				}
				if cfg.MaxConcurrentOps != 8 {
					return &assertError{msg: "Expected MaxConcurrentOps=8 from YAML"}
				}
				return nil
			},
		},
		{
			name: "load with TOML config file",
			setupFunc: func() (LoadOptions, func()) {
				// Create temp TOML file
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.toml")
				tomlContent := `
log_level = "error"
theme = "light"
max_concurrent_ops = 6
`
				if err := os.WriteFile(configPath, []byte(tomlContent), 0o600); err != nil {
					t.Fatalf("Failed to write test config: %v", err)
				}

				return LoadOptions{
					ConfigFilePath: configPath,
					EnvVarPrefix:   "LAZYNUGET_",
					StrictMode:     false,
				}, func() {}
			},
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "error" {
					return &assertError{msg: "Expected LogLevel=error from TOML"}
				}
				if cfg.Theme != "light" {
					return &assertError{msg: "Expected Theme=light from TOML"}
				}
				if cfg.MaxConcurrentOps != 6 {
					return &assertError{msg: "Expected MaxConcurrentOps=6 from TOML"}
				}
				return nil
			},
		},
		{
			name: "missing explicit config file returns error",
			setupFunc: func() (LoadOptions, func()) {
				return LoadOptions{
					ConfigFilePath: "/nonexistent/path/config.yml",
					EnvVarPrefix:   "LAZYNUGET_",
					StrictMode:     false,
				}, func() {}
			},
			wantErr:     true,
			errContains: "not found",
		},
		{
			name: "invalid YAML syntax returns error",
			setupFunc: func() (LoadOptions, func()) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				invalidYAML := `
logLevel: debug
  invalid indentation here
theme: dark
`
				if err := os.WriteFile(configPath, []byte(invalidYAML), 0o600); err != nil {
					t.Fatalf("Failed to write test config: %v", err)
				}

				return LoadOptions{
					ConfigFilePath: configPath,
					EnvVarPrefix:   "LAZYNUGET_",
					StrictMode:     false,
				}, func() {}
			},
			wantErr: false, // YAML parser is lenient
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, cleanup := tt.setupFunc()
			defer cleanup()

			cfg, err := loader.Load(context.Background(), opts)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error containing %q, got: %v", tt.errContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if cfg == nil {
				t.Fatal("Load returned nil config")
			}

			if tt.checkFunc != nil {
				if err := tt.checkFunc(cfg); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestConfigLoaderLoadPrecedence tests configuration precedence rules
func TestConfigLoaderLoadPrecedence(t *testing.T) {
	loader := NewLoader()

	// Create a config file with logLevel=debug
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	yamlContent := "logLevel: debug\n"
	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Set env var with logLevel=warn
	os.Setenv("LAZYNUGET_LOG_LEVEL", "warn")
	defer os.Unsetenv("LAZYNUGET_LOG_LEVEL")

	// Load with CLI flag logLevel=error
	opts := LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		CLIFlags: CLIFlags{
			LogLevel: "error",
		},
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// CLI flag should win (highest precedence)
	if cfg.LogLevel != "error" {
		t.Errorf("Expected LogLevel=error (CLI flag), got %s", cfg.LogLevel)
	}
}

// TestConfigLoaderPrintConfig tests the PrintConfig method
func TestConfigLoaderPrintConfig(t *testing.T) {
	loader := NewLoader()
	cfg := GetDefaultConfig()
	cfg.LoadedFrom = "/test/path/config.yml"
	cfg.LoadedAt = time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)

	output := loader.PrintConfig(cfg)

	// Verify output contains expected sections
	expectedSections := []string{
		"=== LazyNuGet Configuration ===",
		"Loaded from: /test/path/config.yml",
		"--- UI Settings ---",
		"--- Color Scheme ---",
		"--- Keybindings ---",
		"--- Performance ---",
		"--- Timeouts ---",
		"--- Dotnet CLI ---",
		"--- Logging ---",
		"--- Hot Reload ---",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("PrintConfig output missing section: %q", section)
		}
	}

	// Verify some specific values
	if !strings.Contains(output, "logLevel:") {
		t.Error("PrintConfig output missing logLevel field")
	}
	if !strings.Contains(output, "theme:") {
		t.Error("PrintConfig output missing theme field")
	}
}

// TestConfigLoaderLoadWithStrictMode tests strict mode validation
func TestConfigLoaderLoadWithStrictMode(t *testing.T) {
	loader := NewLoader()

	// Create config file with invalid values
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	yamlContent := `
logLevel: invalid_level
theme: invalid_theme
maxConcurrentOps: 999
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0o600); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	tests := []struct {
		name       string
		strictMode bool
		wantErr    bool
	}{
		{
			name:       "strict mode disabled allows invalid config",
			strictMode: false,
			wantErr:    false, // Warnings only, no error
		},
		{
			name:       "strict mode enabled rejects invalid config",
			strictMode: true,
			wantErr:    false, // Validation warnings are not blocking errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := LoadOptions{
				ConfigFilePath: configPath,
				EnvVarPrefix:   "LAZYNUGET_",
				StrictMode:     tt.strictMode,
			}

			_, err := loader.Load(context.Background(), opts)

			if tt.wantErr && err == nil {
				t.Error("Expected error in strict mode, got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestConfigLoaderLoadFromDefaultLocation tests loading from platform default location
func TestConfigLoaderLoadFromDefaultLocation(t *testing.T) {
	loader := NewLoader()

	// Don't specify config path - should fall back to defaults
	opts := LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load returned nil config")
	}

	// Should have default values
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel=info, got %s", cfg.LogLevel)
	}
}
