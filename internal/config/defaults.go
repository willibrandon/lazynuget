package config

import (
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// DefaultConfig returns a configuration with all default values
// Platform-specific paths are resolved based on the operating system
func DefaultConfig() *AppConfig {
	return &AppConfig{
		// Flags
		ShowVersion: false,
		ShowHelp:    false,

		// Logging
		LogLevel: "info",
		LogDir:   getDefaultLogDir(),

		// Paths
		ConfigPath: "",
		ConfigDir:  getDefaultConfigDir(),
		CacheDir:   getDefaultCacheDir(),

		// Mode
		NonInteractive: false,
		IsInteractive:  true, // Will be auto-detected

		// UI Preferences
		Theme:       "default",
		CompactMode: false,
		ShowHints:   true,

		// Performance
		StartupTimeout:   5 * time.Second,
		ShutdownTimeout:  3 * time.Second,
		MaxConcurrentOps: 4,

		// Environment
		DotnetPath: findDotnetPath(),
	}
}

// getDefaultConfigDir returns the platform-specific configuration directory
func getDefaultConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, "lazynuget")
	case "darwin":
		home := os.Getenv("HOME")
		return filepath.Join(home, ".config", "lazynuget")
	default: // linux, bsd, etc.
		// Follow XDG Base Directory specification
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			home := os.Getenv("HOME")
			xdgConfig = filepath.Join(home, ".config")
		}
		return filepath.Join(xdgConfig, "lazynuget")
	}
}

// getDefaultLogDir returns the default directory for log files
func getDefaultLogDir() string {
	return filepath.Join(getDefaultConfigDir(), "logs")
}

// getDefaultCacheDir returns the default directory for cache files
func getDefaultCacheDir() string {
	return filepath.Join(getDefaultConfigDir(), "cache")
}

// findDotnetPath attempts to locate the dotnet CLI in PATH
// Returns empty string if not found
func findDotnetPath() string {
	// Will be properly implemented in platform detection phase
	// For now, return empty - caller should check PATH
	return ""
}
