package config

import (
	"testing"
	"time"
)

// TestConvertEnvVarPathToDotNotation tests environment variable path conversion
func TestConvertEnvVarPathToDotNotation(t *testing.T) {
	tests := []struct {
		name    string
		envPath string
		want    string
	}{
		{
			name:    "simple top-level field",
			envPath: "LOG_LEVEL",
			want:    "logLevel",
		},
		{
			name:    "nested color scheme",
			envPath: "COLOR_SCHEME_BORDER",
			want:    "colorScheme.border",
		},
		{
			name:    "nested timeouts",
			envPath: "TIMEOUTS_NETWORK_REQUEST",
			want:    "timeouts.networkRequest",
		},
		{
			name:    "nested log rotation",
			envPath: "LOG_ROTATION_MAX_SIZE",
			want:    "logRotation.maxSize",
		},
		{
			name:    "simple field lowercase",
			envPath: "THEME",
			want:    "theme",
		},
		{
			name:    "multi-word field",
			envPath: "MAX_CONCURRENT_OPS",
			want:    "maxConcurrentOps",
		},
		{
			name:    "boolean field",
			envPath: "COMPACT_MODE",
			want:    "compactMode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertEnvVarPathToDotNotation(tt.envPath)
			if got != tt.want {
				t.Errorf("convertEnvVarPathToDotNotation(%q) = %q, want %q", tt.envPath, got, tt.want)
			}
		})
	}
}

// TestJoinCamelCase tests camelCase conversion
func TestJoinCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{
			name:  "single word",
			parts: []string{"LOG"},
			want:  "log",
		},
		{
			name:  "two words",
			parts: []string{"LOG", "LEVEL"},
			want:  "logLevel",
		},
		{
			name:  "three words",
			parts: []string{"MAX", "CONCURRENT", "OPS"},
			want:  "maxConcurrentOps",
		},
		{
			name:  "empty slice",
			parts: []string{},
			want:  "",
		},
		{
			name:  "mixed case input",
			parts: []string{"Log", "Level"},
			want:  "logLevel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinCamelCase(tt.parts)
			if got != tt.want {
				t.Errorf("joinCamelCase(%v) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}

// TestApplyEnvVarValue tests applying environment variable values to Config
func TestApplyEnvVarValue(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		value     string
		checkFunc func(*Config) error
	}{
		{
			name:  "set logLevel",
			path:  "logLevel",
			value: "debug",
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					return assert{}.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				return nil
			},
		},
		{
			name:  "set maxConcurrentOps",
			path:  "maxConcurrentOps",
			value: "8",
			checkFunc: func(cfg *Config) error {
				if cfg.MaxConcurrentOps != 8 {
					return assert{}.Errorf("Expected MaxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
				}
				return nil
			},
		},
		{
			name:  "set boolean true",
			path:  "compactMode",
			value: "true",
			checkFunc: func(cfg *Config) error {
				if !cfg.CompactMode {
					return assert{}.Errorf("Expected CompactMode=true")
				}
				return nil
			},
		},
		{
			name:  "set boolean false",
			path:  "showHints",
			value: "false",
			checkFunc: func(cfg *Config) error {
				if cfg.ShowHints {
					return assert{}.Errorf("Expected ShowHints=false")
				}
				return nil
			},
		},
		{
			name:  "set duration",
			path:  "refreshInterval",
			value: "10m",
			checkFunc: func(cfg *Config) error {
				if cfg.RefreshInterval != 10*time.Minute {
					return assert{}.Errorf("Expected RefreshInterval=10m, got %v", cfg.RefreshInterval)
				}
				return nil
			},
		},
		{
			name:  "set nested colorScheme.border",
			path:  "colorScheme.border",
			value: "#CUSTOM",
			checkFunc: func(cfg *Config) error {
				if cfg.ColorScheme.Border != "#CUSTOM" {
					return assert{}.Errorf("Expected Border=#CUSTOM, got %s", cfg.ColorScheme.Border)
				}
				return nil
			},
		},
		{
			name:  "set nested timeouts.networkRequest",
			path:  "timeouts.networkRequest",
			value: "60s",
			checkFunc: func(cfg *Config) error {
				if cfg.Timeouts.NetworkRequest != 60*time.Second {
					return assert{}.Errorf("Expected NetworkRequest=60s, got %v", cfg.Timeouts.NetworkRequest)
				}
				return nil
			},
		},
		{
			name:  "set nested logRotation.maxSize",
			path:  "logRotation.maxSize",
			value: "50",
			checkFunc: func(cfg *Config) error {
				if cfg.LogRotation.MaxSize != 50 {
					return assert{}.Errorf("Expected MaxSize=50, got %d", cfg.LogRotation.MaxSize)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := GetDefaultConfig()
			err := applyEnvVarValue(cfg, tt.path, tt.value)
			if err != nil {
				t.Fatalf("applyEnvVarValue failed: %v", err)
			}

			if tt.checkFunc != nil {
				if err := tt.checkFunc(cfg); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

// TestParseBool tests boolean parsing with various formats
func TestParseBool(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		want      bool
		wantError bool
	}{
		{"true lowercase", "true", true, false},
		{"true uppercase", "TRUE", true, false},
		{"true mixed case", "True", true, false},
		{"false lowercase", "false", false, false},
		{"false uppercase", "FALSE", false, false},
		{"1 for true", "1", true, false},
		{"0 for false", "0", false, false},
		{"yes for true", "yes", true, false},
		{"no for false", "no", false, false},
		{"on for true", "on", true, false},
		{"off for false", "off", false, false},
		{"invalid value", "invalid", false, true},
		{"empty string", "", false, true},
		{"yes uppercase", "YES", true, false},
		{"no uppercase", "NO", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBool(tt.value)
			hasError := err != nil

			if hasError != tt.wantError {
				t.Errorf("parseBool(%q) error = %v, wantError %v", tt.value, err, tt.wantError)
				return
			}

			if !tt.wantError && got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

// assert is a simple assertion helper
type assert struct{}

func (assert) Errorf(format string, args ...interface{}) error {
	return &assertError{msg: format, args: args}
}

type assertError struct {
	msg  string
	args []interface{}
}

func (e *assertError) Error() string {
	if len(e.args) == 0 {
		return e.msg
	}
	// Simple format string replacement (not using fmt to keep it simple)
	return e.msg
}
