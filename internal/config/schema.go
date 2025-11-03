package config

import (
	"reflect"
	"time"
)

// GetConfigSchema returns the complete schema for all configuration settings.
// This is the single source of truth for validation, defaults, and hot-reload support.
// See: specs/002-config-management/data-model.md entity #10
func GetConfigSchema() *ConfigSchema {
	return &ConfigSchema{
		Settings: map[string]SettingSchema{
			// Meta fields (not user-configurable, but included for completeness)
			"version": {
				Path:          "version",
				Type:          reflect.TypeOf(""),
				Constraints:   []Constraint{},
				Default:       "1.0",
				HotReloadable: false,
				Description:   "Configuration version number",
			},

			// UI Settings (FR-020 through FR-025)
			"theme": {
				Path: "theme",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{
						Type:    "enum",
						Params:  []string{"default", "dark", "light", "solarized"},
						Message: "must be one of: default, dark, light, solarized",
					},
				},
				Default:       "default",
				HotReloadable: true,
				Description:   "UI theme (default, dark, light, solarized)",
			},

			// ColorScheme nested fields
			"colorScheme.border": {
				Path: "colorScheme.border",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#FFFFFF",
				HotReloadable: true,
				Description:   "Border color for UI elements",
			},
			"colorScheme.borderFocus": {
				Path: "colorScheme.borderFocus",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#00FF00",
				HotReloadable: true,
				Description:   "Border color for focused UI elements",
			},
			"colorScheme.text": {
				Path: "colorScheme.text",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#FFFFFF",
				HotReloadable: true,
				Description:   "Text color",
			},
			"colorScheme.textDim": {
				Path: "colorScheme.textDim",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#808080",
				HotReloadable: true,
				Description:   "Dimmed text color",
			},
			"colorScheme.background": {
				Path: "colorScheme.background",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#000000",
				HotReloadable: true,
				Description:   "Background color",
			},
			"colorScheme.highlight": {
				Path: "colorScheme.highlight",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#FFFF00",
				HotReloadable: true,
				Description:   "Highlight color for selected items",
			},
			"colorScheme.error": {
				Path: "colorScheme.error",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#FF0000",
				HotReloadable: true,
				Description:   "Error message color",
			},
			"colorScheme.warning": {
				Path: "colorScheme.warning",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#FFA500",
				HotReloadable: true,
				Description:   "Warning message color",
			},
			"colorScheme.success": {
				Path: "colorScheme.success",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#00FF00",
				HotReloadable: true,
				Description:   "Success message color",
			},
			"colorScheme.info": {
				Path: "colorScheme.info",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "hexcolor", Params: nil, Message: "must be valid hex color (#RRGGBB)"},
				},
				Default:       "#00FFFF",
				HotReloadable: true,
				Description:   "Info message color",
			},

			"compactMode": {
				Path:          "compactMode",
				Type:          reflect.TypeOf(false),
				Constraints:   []Constraint{},
				Default:       false,
				HotReloadable: true,
				Description:   "Enable compact UI mode with reduced padding",
			},
			"showHints": {
				Path:          "showHints",
				Type:          reflect.TypeOf(false),
				Constraints:   []Constraint{},
				Default:       true,
				HotReloadable: true,
				Description:   "Show keyboard hints at bottom of screen",
			},
			"showLineNumbers": {
				Path:          "showLineNumbers",
				Type:          reflect.TypeOf(false),
				Constraints:   []Constraint{},
				Default:       false,
				HotReloadable: true,
				Description:   "Show line numbers in list views",
			},
			"dateFormat": {
				Path: "dateFormat",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{Type: "dateformat", Params: nil, Message: "must be valid Go time format string"},
				},
				Default:       "2006-01-02",
				HotReloadable: true,
				Description:   "Date format string (Go time layout)",
			},

			// Keybindings (FR-026 through FR-030)
			"keybindingProfile": {
				Path: "keybindingProfile",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{
						Type:    "enum",
						Params:  []string{"default", "vim", "emacs"},
						Message: "must be one of: default, vim, emacs",
					},
				},
				Default:       "default",
				HotReloadable: false,
				Description:   "Keybinding profile (default, vim, emacs) - requires restart",
			},

			// Performance (FR-031 through FR-034)
			"maxConcurrentOps": {
				Path: "maxConcurrentOps",
				Type: reflect.TypeOf(0),
				Constraints: []Constraint{
					{
						Type:    "range",
						Params:  map[string]int{"min": 1, "max": 16},
						Message: "must be between 1 and 16",
					},
				},
				Default:       4,
				HotReloadable: true,
				Description:   "Maximum number of concurrent operations (1-16)",
			},
			"cacheSize": {
				Path: "cacheSize",
				Type: reflect.TypeOf(0),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  0,
						Message: "must be non-negative",
					},
				},
				Default:       50,
				HotReloadable: true,
				Description:   "Cache size in megabytes (0 = disabled)",
			},
			"refreshInterval": {
				Path: "refreshInterval",
				Type: reflect.TypeOf(time.Duration(0)),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  time.Duration(0),
						Message: "must be non-negative duration",
					},
				},
				Default:       time.Duration(0),
				HotReloadable: true,
				Description:   "Auto-refresh interval (0 = disabled)",
			},

			// Timeouts nested fields
			"timeouts.networkRequest": {
				Path: "timeouts.networkRequest",
				Type: reflect.TypeOf(time.Duration(0)),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  1 * time.Second,
						Message: "must be at least 1 second",
					},
				},
				Default:       30 * time.Second,
				HotReloadable: true,
				Description:   "Network request timeout (minimum 1s)",
			},
			"timeouts.dotnetCLI": {
				Path: "timeouts.dotnetCLI",
				Type: reflect.TypeOf(time.Duration(0)),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  1 * time.Second,
						Message: "must be at least 1 second",
					},
				},
				Default:       60 * time.Second,
				HotReloadable: true,
				Description:   "Dotnet CLI command timeout (minimum 1s)",
			},
			"timeouts.fileOperation": {
				Path: "timeouts.fileOperation",
				Type: reflect.TypeOf(time.Duration(0)),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  100 * time.Millisecond,
						Message: "must be at least 100ms",
					},
				},
				Default:       5 * time.Second,
				HotReloadable: true,
				Description:   "File operation timeout (minimum 100ms)",
			},

			// Dotnet CLI Integration (FR-035 through FR-038)
			"dotnetPath": {
				Path:          "dotnetPath",
				Type:          reflect.TypeOf(""),
				Constraints:   []Constraint{},
				Default:       "",
				HotReloadable: false,
				Description:   "Path to dotnet CLI executable (empty = auto-detect from PATH) - requires restart",
			},
			"dotnetVerbosity": {
				Path: "dotnetVerbosity",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{
						Type:    "enum",
						Params:  []string{"quiet", "minimal", "normal", "detailed", "diagnostic"},
						Message: "must be one of: quiet, minimal, normal, detailed, diagnostic",
					},
				},
				Default:       "minimal",
				HotReloadable: true,
				Description:   "Dotnet CLI verbosity level",
			},

			// Logging (FR-039 through FR-042)
			"logLevel": {
				Path: "logLevel",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{
						Type:    "enum",
						Params:  []string{"debug", "info", "warn", "error"},
						Message: "must be one of: debug, info, warn, error",
					},
				},
				Default:       "info",
				HotReloadable: true,
				Description:   "Logging level",
			},
			"logDir": {
				Path:          "logDir",
				Type:          reflect.TypeOf(""),
				Constraints:   []Constraint{},
				Default:       "",
				HotReloadable: false,
				Description:   "Log directory path (empty = platform default) - requires restart",
			},
			"logFormat": {
				Path: "logFormat",
				Type: reflect.TypeOf(""),
				Constraints: []Constraint{
					{
						Type:    "enum",
						Params:  []string{"text", "json"},
						Message: "must be one of: text, json",
					},
				},
				Default:       "text",
				HotReloadable: false,
				Description:   "Log output format - requires restart",
			},

			// LogRotation nested fields
			"logRotation.maxSize": {
				Path: "logRotation.maxSize",
				Type: reflect.TypeOf(0),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  1,
						Message: "must be at least 1 MB",
					},
				},
				Default:       10,
				HotReloadable: true,
				Description:   "Maximum log file size in megabytes before rotation",
			},
			"logRotation.maxAge": {
				Path: "logRotation.maxAge",
				Type: reflect.TypeOf(0),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  1,
						Message: "must be at least 1 day",
					},
				},
				Default:       30,
				HotReloadable: true,
				Description:   "Maximum age of log files in days",
			},
			"logRotation.maxBackups": {
				Path: "logRotation.maxBackups",
				Type: reflect.TypeOf(0),
				Constraints: []Constraint{
					{
						Type:    "min",
						Params:  0,
						Message: "must be non-negative",
					},
				},
				Default:       3,
				HotReloadable: true,
				Description:   "Maximum number of old log files to retain",
			},
			"logRotation.compress": {
				Path:          "logRotation.compress",
				Type:          reflect.TypeOf(false),
				Constraints:   []Constraint{},
				Default:       true,
				HotReloadable: true,
				Description:   "Compress rotated log files with gzip",
			},

			// Hot-Reload (FR-043 through FR-049)
			"hotReload": {
				Path:          "hotReload",
				Type:          reflect.TypeOf(false),
				Constraints:   []Constraint{},
				Default:       false,
				HotReloadable: false,
				Description:   "Enable hot-reload of configuration file changes - requires restart to enable",
			},
		},
	}
}
