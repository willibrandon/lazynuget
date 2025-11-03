package config

import (
	"testing"
	"time"
)

// TestValidatorValidateConfig tests the validation of complete Config structs
func TestValidatorValidateConfig(t *testing.T) {
	schema := GetConfigSchema()
	v := newValidator(schema)

	defaults := GetDefaultConfig()

	// Helper to create a config starting from defaults
	copyWithOverride := func(overrides func(*Config)) *Config {
		cfg := *defaults // Copy defaults
		overrides(&cfg)
		return &cfg
	}

	tests := []struct {
		name          string
		cfg           *Config
		wantErrCount  int
		wantWarnCount int
		checkErrors   []string // Expected error keys
	}{
		{
			name:          "valid default config",
			cfg:           GetDefaultConfig(),
			wantErrCount:  0,
			wantWarnCount: 1, // RefreshInterval=0 fails validation (should be >= 5s)
		},
		{
			name: "invalid maxConcurrentOps too low",
			cfg: copyWithOverride(func(c *Config) {
				c.MaxConcurrentOps = 0
			}),
			wantErrCount:  0, // Falls back to default
			wantWarnCount: 2, // maxConcurrentOps + refreshInterval
			checkErrors:   []string{"maxConcurrentOps"},
		},
		{
			name: "invalid maxConcurrentOps too high",
			cfg: copyWithOverride(func(c *Config) {
				c.MaxConcurrentOps = 999
			}),
			wantErrCount:  0,
			wantWarnCount: 2, // maxConcurrentOps + refreshInterval
			checkErrors:   []string{"maxConcurrentOps"},
		},
		{
			name: "invalid cacheSize too low",
			cfg: copyWithOverride(func(c *Config) {
				c.CacheSize = -1
			}),
			wantErrCount:  0,
			wantWarnCount: 2, // cacheSize + refreshInterval
			checkErrors:   []string{"cacheSize"},
		},
		{
			name: "invalid log level",
			cfg: copyWithOverride(func(c *Config) {
				c.LogLevel = "invalid"
			}),
			wantErrCount:  0,
			wantWarnCount: 2, // logLevel + refreshInterval
			checkErrors:   []string{"logLevel"},
		},
		{
			name: "invalid theme",
			cfg: copyWithOverride(func(c *Config) {
				c.Theme = "nonexistent"
			}),
			wantErrCount:  0,
			wantWarnCount: 2, // theme + refreshInterval
			checkErrors:   []string{"theme"},
		},
		{
			name: "invalid color scheme",
			cfg: copyWithOverride(func(c *Config) {
				c.ColorScheme.Border = "invalid"
				c.ColorScheme.Error = "GGGGGG"
			}),
			wantErrCount:  0,
			wantWarnCount: 3, // 2 colors + refreshInterval
			checkErrors:   []string{"colorScheme.border", "colorScheme.error"},
		},
		{
			name: "invalid timeouts",
			cfg: copyWithOverride(func(c *Config) {
				c.Timeouts.NetworkRequest = 0
				c.Timeouts.DotnetCLI = 100 * time.Millisecond
				c.Timeouts.FileOperation = 50 * time.Millisecond
			}),
			wantErrCount:  0,
			wantWarnCount: 4, // 3 timeouts + refreshInterval
			checkErrors:   []string{"timeouts.networkRequest", "timeouts.dotnetCLI", "timeouts.fileOperation"},
		},
		{
			name: "multiple validation errors",
			cfg: copyWithOverride(func(c *Config) {
				c.LogLevel = "invalid"
				c.MaxConcurrentOps = 999
				c.CacheSize = -1
			}),
			wantErrCount:  0,
			wantWarnCount: 4, // 3 invalid fields + refreshInterval
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.validate(tt.cfg)

			// Count errors vs warnings
			errCount := 0
			warnCount := 0
			for _, e := range errors {
				if e.Severity == "error" {
					errCount++
				} else {
					warnCount++
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("Expected %d errors, got %d", tt.wantErrCount, errCount)
				for _, e := range errors {
					if e.Severity == "error" {
						t.Logf("  Error: %v", e)
					}
				}
			}

			if warnCount != tt.wantWarnCount {
				t.Errorf("Expected %d warnings, got %d", tt.wantWarnCount, warnCount)
				for _, e := range errors {
					if e.Severity == "warning" {
						t.Logf("  Warning: %v", e)
					}
				}
			}

			// Check specific error keys if provided
			if len(tt.checkErrors) > 0 {
				errorKeys := make(map[string]bool)
				for _, e := range errors {
					errorKeys[e.Key] = true
				}

				for _, expectedKey := range tt.checkErrors {
					if !errorKeys[expectedKey] {
						t.Errorf("Expected error for key %q but didn't find it", expectedKey)
					}
				}
			}
		})
	}
}

// TestValidatorFallbackDefaults tests that validator applies fallback defaults correctly
func TestValidatorFallbackDefaults(t *testing.T) {
	schema := GetConfigSchema()
	v := newValidator(schema)
	defaults := GetDefaultConfig()

	tests := []struct {
		name      string
		cfg       *Config
		checkFunc func(*Config) error
	}{
		{
			name: "invalid maxConcurrentOps falls back",
			cfg: &Config{
				MaxConcurrentOps: 999,
			},
			checkFunc: func(cfg *Config) error {
				if cfg.MaxConcurrentOps != defaults.MaxConcurrentOps {
					t.Errorf("Expected fallback to %d, got %d", defaults.MaxConcurrentOps, cfg.MaxConcurrentOps)
				}
				return nil
			},
		},
		{
			name: "invalid theme falls back",
			cfg: &Config{
				Theme: "nonexistent",
			},
			checkFunc: func(cfg *Config) error {
				if cfg.Theme != defaults.Theme {
					t.Errorf("Expected fallback to %s, got %s", defaults.Theme, cfg.Theme)
				}
				return nil
			},
		},
		{
			name: "invalid color falls back",
			cfg: &Config{
				ColorScheme: ColorScheme{
					Border: "invalid",
				},
			},
			checkFunc: func(cfg *Config) error {
				if cfg.ColorScheme.Border != defaults.ColorScheme.Border {
					t.Errorf("Expected fallback to %s, got %s", defaults.ColorScheme.Border, cfg.ColorScheme.Border)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = v.validate(tt.cfg)
			if tt.checkFunc != nil {
				_ = tt.checkFunc(tt.cfg)
			}
		})
	}
}
