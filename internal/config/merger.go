package config

import (
	"maps"
	"time"
)

// mergeConfigs merges configuration from multiple sources with precedence.
// Precedence order (highest to lowest): CLI flags > env vars > file > defaults
// See: T051, FR-002
func mergeConfigs(base, override *Config) *Config {
	// Start with base config (lower precedence)
	merged := *base

	// Override with non-zero values from override config (higher precedence)
	// This follows the "explicit override" pattern where only explicitly set values override

	// UI Settings
	if override.Theme != "" && override.Theme != base.Theme {
		merged.Theme = override.Theme
	}

	// ColorScheme - merge individual fields
	if override.ColorScheme.Border != "" && override.ColorScheme.Border != base.ColorScheme.Border {
		merged.ColorScheme.Border = override.ColorScheme.Border
	}
	if override.ColorScheme.BorderFocus != "" && override.ColorScheme.BorderFocus != base.ColorScheme.BorderFocus {
		merged.ColorScheme.BorderFocus = override.ColorScheme.BorderFocus
	}
	if override.ColorScheme.Text != "" && override.ColorScheme.Text != base.ColorScheme.Text {
		merged.ColorScheme.Text = override.ColorScheme.Text
	}
	if override.ColorScheme.TextDim != "" && override.ColorScheme.TextDim != base.ColorScheme.TextDim {
		merged.ColorScheme.TextDim = override.ColorScheme.TextDim
	}
	if override.ColorScheme.Background != "" && override.ColorScheme.Background != base.ColorScheme.Background {
		merged.ColorScheme.Background = override.ColorScheme.Background
	}
	if override.ColorScheme.Highlight != "" && override.ColorScheme.Highlight != base.ColorScheme.Highlight {
		merged.ColorScheme.Highlight = override.ColorScheme.Highlight
	}
	if override.ColorScheme.Error != "" && override.ColorScheme.Error != base.ColorScheme.Error {
		merged.ColorScheme.Error = override.ColorScheme.Error
	}
	if override.ColorScheme.Warning != "" && override.ColorScheme.Warning != base.ColorScheme.Warning {
		merged.ColorScheme.Warning = override.ColorScheme.Warning
	}
	if override.ColorScheme.Success != "" && override.ColorScheme.Success != base.ColorScheme.Success {
		merged.ColorScheme.Success = override.ColorScheme.Success
	}
	if override.ColorScheme.Info != "" && override.ColorScheme.Info != base.ColorScheme.Info {
		merged.ColorScheme.Info = override.ColorScheme.Info
	}

	// Bool fields - merge if different from base
	merged.CompactMode = override.CompactMode
	merged.ShowHints = override.ShowHints
	merged.ShowLineNumbers = override.ShowLineNumbers

	if override.DateFormat != "" && override.DateFormat != base.DateFormat {
		merged.DateFormat = override.DateFormat
	}

	// Keybindings
	if override.KeybindingProfile != "" && override.KeybindingProfile != base.KeybindingProfile {
		merged.KeybindingProfile = override.KeybindingProfile
	}
	if len(override.Keybindings) > 0 {
		// Merge keybindings map
		if merged.Keybindings == nil {
			merged.Keybindings = make(map[string]KeyBinding)
		}
		maps.Copy(merged.Keybindings, override.Keybindings)
	}

	// Performance
	if override.MaxConcurrentOps != 0 && override.MaxConcurrentOps != base.MaxConcurrentOps {
		merged.MaxConcurrentOps = override.MaxConcurrentOps
	}
	if override.CacheSize != 0 && override.CacheSize != base.CacheSize {
		merged.CacheSize = override.CacheSize
	}
	if override.RefreshInterval != 0 && override.RefreshInterval != base.RefreshInterval {
		merged.RefreshInterval = override.RefreshInterval
	}

	// Timeouts
	if override.Timeouts.NetworkRequest != 0 && override.Timeouts.NetworkRequest != base.Timeouts.NetworkRequest {
		merged.Timeouts.NetworkRequest = override.Timeouts.NetworkRequest
	}
	if override.Timeouts.DotnetCLI != 0 && override.Timeouts.DotnetCLI != base.Timeouts.DotnetCLI {
		merged.Timeouts.DotnetCLI = override.Timeouts.DotnetCLI
	}
	if override.Timeouts.FileOperation != 0 && override.Timeouts.FileOperation != base.Timeouts.FileOperation {
		merged.Timeouts.FileOperation = override.Timeouts.FileOperation
	}

	// Dotnet CLI
	if override.DotnetPath != "" && override.DotnetPath != base.DotnetPath {
		merged.DotnetPath = override.DotnetPath
	}
	if override.DotnetVerbosity != "" && override.DotnetVerbosity != base.DotnetVerbosity {
		merged.DotnetVerbosity = override.DotnetVerbosity
	}

	// Logging
	if override.LogLevel != "" && override.LogLevel != base.LogLevel {
		merged.LogLevel = override.LogLevel
	}
	if override.LogDir != "" && override.LogDir != base.LogDir {
		merged.LogDir = override.LogDir
	}
	if override.LogFormat != "" && override.LogFormat != base.LogFormat {
		merged.LogFormat = override.LogFormat
	}

	// Log Rotation
	if override.LogRotation.MaxSize != 0 && override.LogRotation.MaxSize != base.LogRotation.MaxSize {
		merged.LogRotation.MaxSize = override.LogRotation.MaxSize
	}
	if override.LogRotation.MaxAge != 0 && override.LogRotation.MaxAge != base.LogRotation.MaxAge {
		merged.LogRotation.MaxAge = override.LogRotation.MaxAge
	}
	if override.LogRotation.MaxBackups != base.LogRotation.MaxBackups {
		merged.LogRotation.MaxBackups = override.LogRotation.MaxBackups
	}
	merged.LogRotation.Compress = override.LogRotation.Compress

	// Hot-Reload
	merged.HotReload = override.HotReload

	// Update metadata to reflect merge
	merged.LoadedAt = time.Now()

	return &merged
}
