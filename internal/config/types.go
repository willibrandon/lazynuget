package config

import "time"

// AppConfig represents the merged configuration from all sources (CLI, env, file, defaults)
// After loading, configuration is immutable
type AppConfig struct {
	Theme            string
	DotnetPath       string
	LogLevel         string
	LogDir           string
	ConfigPath       string
	ConfigDir        string
	CacheDir         string
	StartupTimeout   time.Duration
	ShutdownTimeout  time.Duration
	MaxConcurrentOps int
	IsInteractive    bool
	NonInteractive   bool
	CompactMode      bool
	ShowHints        bool
	ShowVersion      bool
	ShowHelp         bool
}

// Flags holds the parsed command-line flags
// Separate from AppConfig to allow early --version/--help handling
type Flags struct {
	ConfigPath     string
	LogLevel       string
	ShowVersion    bool
	ShowHelp       bool
	NonInteractive bool
}
