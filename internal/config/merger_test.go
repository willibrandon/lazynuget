package config

import (
	"testing"
	"time"
)

// TestMergeConfigs tests the mergeConfigs function with table-driven tests
func TestMergeConfigs(t *testing.T) {
	tests := []struct {
		base     *Config
		override *Config
		want     map[string]any
		name     string
	}{
		{
			name: "override simple fields",
			base: GetDefaultConfig(),
			override: &Config{
				LogLevel:         "debug",
				MaxConcurrentOps: 8,
				Theme:            "dark",
			},
			want: map[string]any{
				"LogLevel":         "debug",
				"MaxConcurrentOps": 8,
				"Theme":            "dark",
			},
		},
		{
			name:     "empty override keeps defaults",
			base:     GetDefaultConfig(),
			override: &Config{},
			want: map[string]any{
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
			want: map[string]any{
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
			want: map[string]any{
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
			want: map[string]any{
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
			want: map[string]any{
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
				var actual any
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

// TestMergeConfigsAllColorSchemeFields tests all ColorScheme fields are merged
func TestMergeConfigsAllColorSchemeFields(t *testing.T) {
	base := &Config{
		ColorScheme: ColorScheme{
			Border:      "#FFFFFF",
			BorderFocus: "#00FF00",
			Text:        "#000000",
			TextDim:     "#808080",
			Background:  "#FFFFFF",
			Highlight:   "#FFFF00",
			Error:       "#FF0000",
			Warning:     "#FFA500",
			Success:     "#00FF00",
			Info:        "#0000FF",
		},
	}

	override := &Config{
		ColorScheme: ColorScheme{
			Border:      "#FF0000",
			BorderFocus: "#0000FF",
			Text:        "#FFFFFF",
			TextDim:     "#404040",
			Background:  "#000000",
			Highlight:   "#00FFFF",
			Error:       "#FF00FF",
			Warning:     "#FFFF00",
			Success:     "#00FFFF",
			Info:        "#FF00FF",
		},
	}

	merged := mergeConfigs(base, override)

	if merged.ColorScheme.Border != "#FF0000" {
		t.Errorf("Border = %v, want #FF0000", merged.ColorScheme.Border)
	}
	if merged.ColorScheme.BorderFocus != "#0000FF" {
		t.Errorf("BorderFocus = %v, want #0000FF", merged.ColorScheme.BorderFocus)
	}
	if merged.ColorScheme.Text != "#FFFFFF" {
		t.Errorf("Text = %v, want #FFFFFF", merged.ColorScheme.Text)
	}
	if merged.ColorScheme.TextDim != "#404040" {
		t.Errorf("TextDim = %v, want #404040", merged.ColorScheme.TextDim)
	}
	if merged.ColorScheme.Background != "#000000" {
		t.Errorf("Background = %v, want #000000", merged.ColorScheme.Background)
	}
	if merged.ColorScheme.Highlight != "#00FFFF" {
		t.Errorf("Highlight = %v, want #00FFFF", merged.ColorScheme.Highlight)
	}
	if merged.ColorScheme.Error != "#FF00FF" {
		t.Errorf("Error = %v, want #FF00FF", merged.ColorScheme.Error)
	}
	if merged.ColorScheme.Warning != "#FFFF00" {
		t.Errorf("Warning = %v, want #FFFF00", merged.ColorScheme.Warning)
	}
	if merged.ColorScheme.Success != "#00FFFF" {
		t.Errorf("Success = %v, want #00FFFF", merged.ColorScheme.Success)
	}
	if merged.ColorScheme.Info != "#FF00FF" {
		t.Errorf("Info = %v, want #FF00FF", merged.ColorScheme.Info)
	}
}

// TestMergeConfigsLogRotation tests LogRotation merging
func TestMergeConfigsLogRotation(t *testing.T) {
	base := &Config{
		LogRotation: LogRotation{
			MaxSize:    100,
			MaxAge:     30,
			MaxBackups: 5,
			Compress:   false,
		},
	}

	override := &Config{
		LogRotation: LogRotation{
			MaxSize:    200,
			MaxAge:     60,
			MaxBackups: 10,
			Compress:   true,
		},
	}

	merged := mergeConfigs(base, override)

	if merged.LogRotation.MaxSize != 200 {
		t.Errorf("MaxSize = %v, want 200", merged.LogRotation.MaxSize)
	}
	if merged.LogRotation.MaxAge != 60 {
		t.Errorf("MaxAge = %v, want 60", merged.LogRotation.MaxAge)
	}
	if merged.LogRotation.MaxBackups != 10 {
		t.Errorf("MaxBackups = %v, want 10", merged.LogRotation.MaxBackups)
	}
	if !merged.LogRotation.Compress {
		t.Error("Compress should be true")
	}
}

// TestMergeConfigsAllTimeouts tests all Timeout fields
func TestMergeConfigsAllTimeouts(t *testing.T) {
	base := &Config{
		Timeouts: Timeouts{
			NetworkRequest: 30 * time.Second,
			DotnetCLI:      60 * time.Second,
			FileOperation:  10 * time.Second,
		},
	}

	override := &Config{
		Timeouts: Timeouts{
			NetworkRequest: 45 * time.Second,
			DotnetCLI:      90 * time.Second,
			FileOperation:  15 * time.Second,
		},
	}

	merged := mergeConfigs(base, override)

	if merged.Timeouts.NetworkRequest != 45*time.Second {
		t.Errorf("NetworkRequest = %v, want 45s", merged.Timeouts.NetworkRequest)
	}
	if merged.Timeouts.DotnetCLI != 90*time.Second {
		t.Errorf("DotnetCLI = %v, want 90s", merged.Timeouts.DotnetCLI)
	}
	if merged.Timeouts.FileOperation != 15*time.Second {
		t.Errorf("FileOperation = %v, want 15s", merged.Timeouts.FileOperation)
	}
}

// TestMergeConfigsKeybindings tests keybindings merging with nil map
func TestMergeConfigsKeybindingsNilMap(t *testing.T) {
	base := &Config{
		Keybindings: nil,
	}

	override := &Config{
		Keybindings: map[string]KeyBinding{
			"quit": {Key: "q", Action: "quit", Context: "global"},
		},
	}

	merged := mergeConfigs(base, override)

	if merged.Keybindings == nil {
		t.Fatal("Keybindings should not be nil")
	}
	if len(merged.Keybindings) != 1 {
		t.Errorf("Expected 1 keybinding, got %d", len(merged.Keybindings))
	}
	if merged.Keybindings["quit"].Key != "q" {
		t.Error("quit keybinding not found")
	}
}

// TestMergeConfigsAllStringFields tests all string field merging
func TestMergeConfigsAllStringFields(t *testing.T) {
	base := &Config{
		DateFormat:        "2006-01-02",
		DotnetPath:        "dotnet",
		DotnetVerbosity:   "minimal",
		LogDir:            "/var/log",
		LogFormat:         "text",
		KeybindingProfile: "default",
	}

	override := &Config{
		DateFormat:        "01/02/2006",
		DotnetPath:        "/custom/dotnet",
		DotnetVerbosity:   "detailed",
		LogDir:            "/custom/log",
		LogFormat:         "json",
		KeybindingProfile: "vim",
	}

	merged := mergeConfigs(base, override)

	if merged.DateFormat != "01/02/2006" {
		t.Errorf("DateFormat = %v, want 01/02/2006", merged.DateFormat)
	}
	if merged.DotnetPath != "/custom/dotnet" {
		t.Errorf("DotnetPath = %v, want /custom/dotnet", merged.DotnetPath)
	}
	if merged.DotnetVerbosity != "detailed" {
		t.Errorf("DotnetVerbosity = %v, want detailed", merged.DotnetVerbosity)
	}
	if merged.LogDir != "/custom/log" {
		t.Errorf("LogDir = %v, want /custom/log", merged.LogDir)
	}
	if merged.LogFormat != "json" {
		t.Errorf("LogFormat = %v, want json", merged.LogFormat)
	}
	if merged.KeybindingProfile != "vim" {
		t.Errorf("KeybindingProfile = %v, want vim", merged.KeybindingProfile)
	}
}
