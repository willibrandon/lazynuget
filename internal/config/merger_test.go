package config

import (
	"testing"
	"time"
)

// TestMergeConfigs tests the mergeConfigs function with table-driven tests
func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		name     string
		base     *Config
		override *Config
		want     map[string]interface{} // field name -> expected value
	}{
		{
			name: "override simple fields",
			base: GetDefaultConfig(),
			override: &Config{
				LogLevel:         "debug",
				MaxConcurrentOps: 8,
				Theme:            "dark",
			},
			want: map[string]interface{}{
				"LogLevel":         "debug",
				"MaxConcurrentOps": 8,
				"Theme":            "dark",
			},
		},
		{
			name: "empty override keeps defaults",
			base: GetDefaultConfig(),
			override: &Config{},
			want: map[string]interface{}{
				"LogLevel":         GetDefaultConfig().LogLevel,
				"MaxConcurrentOps": GetDefaultConfig().MaxConcurrentOps,
				"Theme":            GetDefaultConfig().Theme,
			},
		},
		{
			name: "partial override",
			base: GetDefaultConfig(),
			override: &Config{
				LogLevel: "error",
			},
			want: map[string]interface{}{
				"LogLevel":         "error",
				"MaxConcurrentOps": GetDefaultConfig().MaxConcurrentOps,
				"CacheSize":        GetDefaultConfig().CacheSize,
			},
		},
		{
			name: "override nested colorScheme",
			base: GetDefaultConfig(),
			override: &Config{
				ColorScheme: ColorScheme{
					Border: "#CUSTOM",
					Error:  "#FF0000",
				},
			},
			want: map[string]interface{}{
				"ColorScheme.Border": "#CUSTOM",
				"ColorScheme.Error":  "#FF0000",
			},
		},
		{
			name: "override timeouts",
			base: GetDefaultConfig(),
			override: &Config{
				Timeouts: Timeouts{
					NetworkRequest: 60 * time.Second,
					DotnetCLI:      5 * time.Minute,
				},
			},
			want: map[string]interface{}{
				"Timeouts.NetworkRequest": 60 * time.Second,
				"Timeouts.DotnetCLI":      5 * time.Minute,
			},
		},
		{
			name: "override boolean flags",
			base: GetDefaultConfig(),
			override: &Config{
				CompactMode:     true,
				ShowHints:       false,
				ShowLineNumbers: true,
				HotReload:       true,
			},
			want: map[string]interface{}{
				"CompactMode":     true,
				"ShowHints":       false,
				"ShowLineNumbers": true,
				"HotReload":       true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			merged := mergeConfigs(tt.base, tt.override)

			// Check expected values
			for field, expected := range tt.want {
				var actual interface{}
				switch field {
				case "LogLevel":
					actual = merged.LogLevel
				case "MaxConcurrentOps":
					actual = merged.MaxConcurrentOps
				case "Theme":
					actual = merged.Theme
				case "CacheSize":
					actual = merged.CacheSize
				case "CompactMode":
					actual = merged.CompactMode
				case "ShowHints":
					actual = merged.ShowHints
				case "ShowLineNumbers":
					actual = merged.ShowLineNumbers
				case "HotReload":
					actual = merged.HotReload
				case "ColorScheme.Border":
					actual = merged.ColorScheme.Border
				case "ColorScheme.Error":
					actual = merged.ColorScheme.Error
				case "Timeouts.NetworkRequest":
					actual = merged.Timeouts.NetworkRequest
				case "Timeouts.DotnetCLI":
					actual = merged.Timeouts.DotnetCLI
				default:
					t.Fatalf("Unknown field in test: %s", field)
				}

				if actual != expected {
					t.Errorf("Field %s: expected %v, got %v", field, expected, actual)
				}
			}
		})
	}
}

// TestMergeConfigsPreservesDefaults ensures that fields not overridden retain default values
func TestMergeConfigsPreservesDefaults(t *testing.T) {
	base := GetDefaultConfig()
	override := &Config{
		LogLevel: "debug", // Only override one field
	}

	merged := mergeConfigs(base, override)

	// Verify the override was applied
	if merged.LogLevel != "debug" {
		t.Errorf("Expected LogLevel=debug, got %s", merged.LogLevel)
	}

	// Verify other fields kept defaults
	if merged.MaxConcurrentOps != base.MaxConcurrentOps {
		t.Errorf("Expected MaxConcurrentOps=%d (default), got %d", base.MaxConcurrentOps, merged.MaxConcurrentOps)
	}
	if merged.CacheSize != base.CacheSize {
		t.Errorf("Expected CacheSize=%d (default), got %d", base.CacheSize, merged.CacheSize)
	}
	if merged.Theme != base.Theme {
		t.Errorf("Expected Theme=%s (default), got %s", base.Theme, merged.Theme)
	}
}
