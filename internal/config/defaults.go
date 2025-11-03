package config

import (
	"time"
)

// GetDefaultConfig returns a Config with all default values populated.
// This is the base configuration used when no other sources are available.
// See: specs/002-config-management/plan.md, FR-001
func GetDefaultConfig() *Config {
	return &Config{
		// Meta
		Version:    "1.0",
		LoadedFrom: "defaults",
		LoadedAt:   time.Now(),

		// UI Settings (FR-020 through FR-025)
		Theme: "default",
		ColorScheme: ColorScheme{
			Border:      "#FFFFFF",
			BorderFocus: "#00FF00",
			Text:        "#FFFFFF",
			TextDim:     "#808080",
			Background:  "#000000",
			Highlight:   "#FFFF00",
			Error:       "#FF0000",
			Warning:     "#FFA500",
			Success:     "#00FF00",
			Info:        "#00FFFF",
		},
		CompactMode:     false,
		ShowHints:       true,
		ShowLineNumbers: false,
		DateFormat:      "2006-01-02",

		// Keybindings (FR-026 through FR-030)
		Keybindings:       make(map[string]KeyBinding),
		KeybindingProfile: "default",

		// Performance (FR-031 through FR-034)
		MaxConcurrentOps: 4,
		CacheSize:        50, // MB
		RefreshInterval:  0,  // Disabled
		Timeouts: Timeouts{
			NetworkRequest: 30 * time.Second,
			DotnetCLI:      60 * time.Second,
			FileOperation:  5 * time.Second,
		},

		// Dotnet CLI Integration (FR-035 through FR-038)
		DotnetPath:      "", // Empty = auto-detect from PATH
		DotnetVerbosity: "minimal",

		// Logging (FR-039 through FR-042)
		LogLevel:  "info",
		LogDir:    "", // Empty = platform default
		LogFormat: "text",
		LogRotation: LogRotation{
			MaxSize:    10,   // MB
			MaxAge:     30,   // Days
			MaxBackups: 3,    // Keep 3 old logs
			Compress:   true, // Compress rotated logs
		},

		// Hot-Reload (FR-043)
		HotReload: false, // Disabled by default for safety
	}
}
