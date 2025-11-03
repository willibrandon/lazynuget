// Package config provides configuration management for LazyNuGet.
package config

// AppConfig holds all application configuration.
type AppConfig struct {
	// LogLevel specifies the logging verbosity (debug, info, warn, error)
	LogLevel string

	// ConfigPath is the path to the configuration file (if specified)
	ConfigPath string

	// NonInteractive forces non-interactive mode (no TUI)
	NonInteractive bool
}

// Load loads configuration from all sources (CLI flags, env vars, files, defaults).
// For now, this is a stub that returns default configuration.
func Load() (*AppConfig, error) {
	return &AppConfig{
		LogLevel:       "info",
		ConfigPath:     "",
		NonInteractive: false,
	}, nil
}
