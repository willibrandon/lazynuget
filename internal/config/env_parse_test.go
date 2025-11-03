package config

import (
	"os"
	"testing"
	"time"
)

// TestParseEnvVars tests the main environment variable parsing function
func TestParseEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		prefix   string
		wantVars map[string]string
	}{
		{
			name: "parse simple env vars",
			envVars: map[string]string{
				"LAZYNUGET_LOG_LEVEL": "debug",
				"LAZYNUGET_THEME":     "dark",
			},
			prefix: "LAZYNUGET_",
			wantVars: map[string]string{
				"logLevel": "debug",
				"theme":    "dark",
			},
		},
		{
			name: "parse nested env vars",
			envVars: map[string]string{
				"LAZYNUGET_COLOR_SCHEME_BORDER": "#CUSTOM",
				"LAZYNUGET_TIMEOUTS_NETWORK_REQUEST": "60s",
			},
			prefix: "LAZYNUGET_",
			wantVars: map[string]string{
				"colorScheme.border":        "#CUSTOM",
				"timeouts.networkRequest":   "60s",
			},
		},
		{
			name: "parse multi-word fields",
			envVars: map[string]string{
				"LAZYNUGET_MAX_CONCURRENT_OPS": "8",
				"LAZYNUGET_SHOW_LINE_NUMBERS":  "true",
			},
			prefix: "LAZYNUGET_",
			wantVars: map[string]string{
				"maxConcurrentOps": "8",
				"showLineNumbers":  "true",
			},
		},
		{
			name: "ignore env vars without prefix",
			envVars: map[string]string{
				"PATH":                  "/usr/bin",
				"HOME":                  "/home/user",
				"LAZYNUGET_LOG_LEVEL": "debug",
			},
			prefix: "LAZYNUGET_",
			wantVars: map[string]string{
				"logLevel": "debug",
			},
		},
		{
			name: "LAZYNUGET_CONFIG is parsed like any other",
			envVars: map[string]string{
				"LAZYNUGET_CONFIG":     "/path/to/config",
				"LAZYNUGET_LOG_LEVEL": "debug",
			},
			prefix: "LAZYNUGET_",
			wantVars: map[string]string{
				"config":   "/path/to/config",
				"logLevel": "debug",
			},
		},
		{
			name:     "no matching env vars",
			envVars:  map[string]string{
				"PATH": "/usr/bin",
				"HOME": "/home/user",
			},
			prefix:   "LAZYNUGET_",
			wantVars: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Parse env vars
			result := parseEnvVars(tt.prefix)

			// Check result
			if len(result) != len(tt.wantVars) {
				t.Errorf("Expected %d parsed vars, got %d", len(tt.wantVars), len(result))
			}

			for key, wantValue := range tt.wantVars {
				gotValue, ok := result[key]
				if !ok {
					t.Errorf("Expected key %q in result, not found", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("For key %q: expected value %q, got %q", key, wantValue, gotValue)
				}
			}

			// Check for unexpected keys
			for key := range result {
				if _, ok := tt.wantVars[key]; !ok {
					t.Errorf("Unexpected key %q in result with value %q", key, result[key])
				}
			}
		})
	}
}

// TestParseEnvVarsIntegration tests full integration with Config struct
func TestParseEnvVarsIntegration(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		prefix    string
		checkFunc func(*Config) error
	}{
		{
			name: "apply simple fields",
			envVars: map[string]string{
				"LAZYNUGET_LOG_LEVEL": "debug",
				"LAZYNUGET_THEME":     "dark",
			},
			prefix: "LAZYNUGET_",
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					return &assertError{msg: "Expected LogLevel=debug"}
				}
				if cfg.Theme != "dark" {
					return &assertError{msg: "Expected Theme=dark"}
				}
				return nil
			},
		},
		{
			name: "apply boolean fields",
			envVars: map[string]string{
				"LAZYNUGET_COMPACT_MODE": "true",
				"LAZYNUGET_SHOW_HINTS":   "false",
			},
			prefix: "LAZYNUGET_",
			checkFunc: func(cfg *Config) error {
				if !cfg.CompactMode {
					return &assertError{msg: "Expected CompactMode=true"}
				}
				if cfg.ShowHints {
					return &assertError{msg: "Expected ShowHints=false"}
				}
				return nil
			},
		},
		{
			name: "apply integer fields",
			envVars: map[string]string{
				"LAZYNUGET_MAX_CONCURRENT_OPS": "8",
				"LAZYNUGET_CACHE_SIZE":         "256",
			},
			prefix: "LAZYNUGET_",
			checkFunc: func(cfg *Config) error {
				if cfg.MaxConcurrentOps != 8 {
					return &assertError{msg: "Expected MaxConcurrentOps=8"}
				}
				if cfg.CacheSize != 256 {
					return &assertError{msg: "Expected CacheSize=256"}
				}
				return nil
			},
		},
		{
			name: "apply nested fields",
			envVars: map[string]string{
				"LAZYNUGET_COLOR_SCHEME_BORDER": "#CUSTOM",
				"LAZYNUGET_COLOR_SCHEME_ERROR":  "#FF0000",
			},
			prefix: "LAZYNUGET_",
			checkFunc: func(cfg *Config) error {
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
			name: "apply duration fields",
			envVars: map[string]string{
				"LAZYNUGET_REFRESH_INTERVAL": "10m",
			},
			prefix: "LAZYNUGET_",
			checkFunc: func(cfg *Config) error {
				// Duration parsing converts "10m" to 10 * time.Minute
				expected := 10 * time.Minute
				if cfg.RefreshInterval != expected {
					return &assertError{msg: "Expected RefreshInterval=10m0s"}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			// Start with default config
			cfg := GetDefaultConfig()

			// Parse and apply env vars
			envVars := parseEnvVars(tt.prefix)
			for path, value := range envVars {
				if err := applyEnvVarValue(cfg, path, value); err != nil {
					t.Errorf("Failed to apply env var %s=%s: %v", path, value, err)
				}
			}

			// Check result
			if tt.checkFunc != nil {
				if err := tt.checkFunc(cfg); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestParseEnvVarsDoubleNested tests deeply nested environment variables
func TestParseEnvVarsDoubleNested(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		value    string
		expected string
	}{
		{
			name:     "log rotation max size",
			envVar:   "LAZYNUGET_LOG_ROTATION_MAX_SIZE",
			value:    "50",
			expected: "logRotation.maxSize",
		},
		{
			name:     "log rotation compress",
			envVar:   "LAZYNUGET_LOG_ROTATION_COMPRESS",
			value:    "true",
			expected: "logRotation.compress",
		},
		{
			name:     "timeouts dotnet CLI",
			envVar:   "LAZYNUGET_TIMEOUTS_DOTNET_CLI",
			value:    "5m",
			expected: "timeouts.dotnetCli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.envVar, tt.value)
			defer os.Unsetenv(tt.envVar)

			result := parseEnvVars("LAZYNUGET_")

			if len(result) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(result))
			}

			if _, ok := result[tt.expected]; !ok {
				t.Errorf("Expected key %q, got keys: %v", tt.expected, result)
			}

			if result[tt.expected] != tt.value {
				t.Errorf("Expected value %q, got %q", tt.value, result[tt.expected])
			}
		})
	}
}
