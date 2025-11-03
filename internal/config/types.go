package config

import "time"

// AppConfig represents the merged configuration from all sources (CLI, env, file, defaults)
// After loading, configuration is immutable
type AppConfig struct {
	// CLI Flags (early exit)
	ShowVersion bool
	ShowHelp    bool

	// Logging
	LogLevel string // "debug", "info", "warn", "error"
	LogDir   string // Directory for log files

	// Paths
	ConfigPath string // Explicit config file path (if provided)
	ConfigDir  string // Configuration directory
	CacheDir   string // Cache directory

	// Mode
	NonInteractive bool // Force non-interactive mode
	IsInteractive  bool // TTY detected (auto-set)

	// UI Preferences (loaded from file)
	Theme       string // Color theme name
	CompactMode bool   // Use compact UI layout
	ShowHints   bool   // Display keyboard hints

	// Performance
	StartupTimeout   time.Duration // Max startup time
	ShutdownTimeout  time.Duration // Max shutdown time
	MaxConcurrentOps int           // Parallel operation limit

	// Environment Detection
	DotnetPath string // Path to dotnet CLI
}

// Flags holds the parsed command-line flags
// Separate from AppConfig to allow early --version/--help handling
type Flags struct {
	ShowVersion    bool
	ShowHelp       bool
	ConfigPath     string
	LogLevel       string
	NonInteractive bool
}
