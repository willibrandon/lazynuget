package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// getPlatformConfigPath returns the platform-specific default configuration directory.
// See: FR-006
//
// Returns:
//   - macOS: ~/Library/Application Support/lazynuget/
//   - Linux: ~/.config/lazynuget/
//   - Windows: %APPDATA%\lazynuget\
//
// If the user's home directory cannot be determined, returns an empty string.
func getPlatformConfigPath() string {
	switch runtime.GOOS {
	case "darwin": // macOS
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, "Library", "Application Support", "lazynuget")

	case "linux":
		// Follow XDG Base Directory Specification
		// First check XDG_CONFIG_HOME, fall back to ~/.config
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			return filepath.Join(xdgConfigHome, "lazynuget")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, ".config", "lazynuget")

	case "windows":
		// Use %APPDATA% environment variable
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "lazynuget")
		}
		// Fallback: try user profile + AppData\Roaming
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, "AppData", "Roaming", "lazynuget")

	default:
		// Unknown platform - return empty string
		// Caller should handle this by using current directory or explicit --config flag
		return ""
	}
}
