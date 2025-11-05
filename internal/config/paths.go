package config

import (
	"github.com/willibrandon/lazynuget/internal/platform"
)

// getPlatformConfigPath returns the platform-specific default configuration directory
// using the platform abstraction layer.
// See: FR-006, specs/003-platform-abstraction
//
// Returns:
//   - macOS: ~/Library/Application Support/lazynuget/
//   - Linux: $XDG_CONFIG_HOME/lazynuget/ or ~/.config/lazynuget/
//   - Windows: %APPDATA%\lazynuget\
//
// If platform detection or path resolution fails, returns an empty string.
// Caller should handle this by using explicit --config flag or failing gracefully.
func getPlatformConfigPath() string {
	// Get platform info singleton
	platformInfo, err := platform.New()
	if err != nil {
		// Platform detection failed - return empty string
		// Caller will handle by checking for explicit --config flag
		return ""
	}

	// Create path resolver
	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		// Path resolver creation failed - return empty string
		return ""
	}

	// Get config directory
	configDir, err := pathResolver.ConfigDir()
	if err != nil {
		// Config directory resolution failed - return empty string
		return ""
	}

	return configDir
}
