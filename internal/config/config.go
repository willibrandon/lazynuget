package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load creates configuration by merging sources with precedence: CLI > Env > File > Defaults
// This is the main entry point for configuration loading
func Load(flags *Flags) (*AppConfig, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Apply config path from flags FIRST (before loading file)
	// This allows --config flag to specify which file to load
	if flags != nil && flags.ConfigPath != "" {
		cfg.ConfigPath = flags.ConfigPath
	}

	// Try to load config file (gracefully handle missing file)
	if err := loadConfigFile(cfg); err != nil {
		// Only return error if file exists but can't be read/parsed
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
		// File doesn't exist - that's okay, use defaults
	}

	// Apply environment variables (overrides file)
	applyEnvironmentVariables(cfg)

	// Apply remaining CLI flags (highest precedence)
	if flags != nil {
		applyCLIFlags(cfg, flags)
	}

	// Validate the merged configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// applyCLIFlags applies command-line flags to configuration.
// CLI flags have the highest precedence in the configuration hierarchy.
func applyCLIFlags(cfg *AppConfig, flags *Flags) {
	if flags.ConfigPath != "" {
		cfg.ConfigPath = flags.ConfigPath
	}
	if flags.LogLevel != "" {
		cfg.LogLevel = flags.LogLevel
	}
	if flags.NonInteractive {
		cfg.NonInteractive = true
	}
}

// loadConfigFile attempts to load configuration from the default config file location
func loadConfigFile(cfg *AppConfig) error {
	// Determine config file path
	configPath := cfg.ConfigPath
	if configPath == "" {
		// Use default location
		configPath = filepath.Join(cfg.ConfigDir, "config.yml")
	}

	// Check if file exists
	// Validate config path (security: prevent path traversal, ensure valid extension)
	cleanPath := filepath.Clean(configPath)
	if !strings.HasSuffix(strings.ToLower(cleanPath), ".yml") && !strings.HasSuffix(strings.ToLower(cleanPath), ".yaml") {
		return fmt.Errorf("config file must have .yml or .yaml extension: %s", configPath)
	}

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		return err // Return NotExist error (will be handled gracefully)
	}

	// Read file
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML with strict mode to catch syntax errors
	var fileConfig struct {
		LogLevel         string        `yaml:"logLevel"`
		LogDir           string        `yaml:"logDir"`
		Theme            string        `yaml:"theme"`
		CompactMode      bool          `yaml:"compactMode"`
		ShowHints        bool          `yaml:"showHints"`
		StartupTimeout   time.Duration `yaml:"startupTimeout"`
		ShutdownTimeout  time.Duration `yaml:"shutdownTimeout"`
		MaxConcurrentOps int           `yaml:"maxConcurrentOps"`
	}

	// Use decoder with strict mode to catch unknown fields and syntax errors
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true) // Fail on unknown fields

	if err := decoder.Decode(&fileConfig); err != nil {
		// Provide helpful YAML parsing error with line/column info
		return fmt.Errorf("failed to parse YAML config file %s: %w\n\n"+
			"Please check the file for syntax errors:\n"+
			"  • Ensure proper indentation (use spaces, not tabs)\n"+
			"  • Check for missing colons or quotes\n"+
			"  • Validate YAML syntax at https://www.yamllint.com/", configPath, err)
	}

	// Apply file config (only if values are non-zero)
	if fileConfig.LogLevel != "" {
		cfg.LogLevel = fileConfig.LogLevel
	}
	if fileConfig.LogDir != "" {
		cfg.LogDir = fileConfig.LogDir
	}
	if fileConfig.Theme != "" {
		cfg.Theme = fileConfig.Theme
	}
	if fileConfig.CompactMode {
		cfg.CompactMode = fileConfig.CompactMode
	}
	if fileConfig.ShowHints {
		cfg.ShowHints = fileConfig.ShowHints
	}
	if fileConfig.StartupTimeout > 0 {
		cfg.StartupTimeout = fileConfig.StartupTimeout
	}
	if fileConfig.ShutdownTimeout > 0 {
		cfg.ShutdownTimeout = fileConfig.ShutdownTimeout
	}
	if fileConfig.MaxConcurrentOps > 0 {
		cfg.MaxConcurrentOps = fileConfig.MaxConcurrentOps
	}

	return nil
}

// applyEnvironmentVariables applies environment variable overrides
func applyEnvironmentVariables(cfg *AppConfig) {
	// Log level
	if logLevel := os.Getenv("LAZYNUGET_LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// Config path
	if configPath := os.Getenv("LAZYNUGET_CONFIG"); configPath != "" {
		cfg.ConfigPath = configPath
	}

	// Non-interactive mode detection from CI environment
	ciValue := os.Getenv("CI")
	noTTY := os.Getenv("NO_TTY")
	if ciValue == "true" || ciValue == "1" || noTTY == "1" {
		cfg.NonInteractive = true
	}
}

// validationError creates a structured validation error message
func validationError(field, constraint, currentValue string) error {
	return fmt.Errorf("configuration validation failed:\n"+
		"  Field: %s\n"+
		"  Constraint: %s\n"+
		"  Current value: %s", field, constraint, currentValue)
}

// Validate checks all configuration values against validation rules
func (cfg *AppConfig) Validate() error {
	// VR-006: Log level must be valid
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[strings.ToLower(cfg.LogLevel)] {
		return validationError(
			"logLevel",
			"must be one of: debug, info, warn, error",
			fmt.Sprintf("%q", cfg.LogLevel),
		)
	}

	// VR-007: Config path must exist if specified
	if cfg.ConfigPath != "" {
		if _, err := os.Stat(cfg.ConfigPath); os.IsNotExist(err) {
			return validationError(
				"configPath",
				"file must exist if specified",
				fmt.Sprintf("%q (file not found)", cfg.ConfigPath),
			)
		}
	}

	// VR-008: Startup timeout bounds
	if cfg.StartupTimeout < 1*time.Second || cfg.StartupTimeout > 30*time.Second {
		return validationError(
			"startupTimeout",
			"must be between 1s and 30s",
			fmt.Sprintf("%v", cfg.StartupTimeout),
		)
	}

	// VR-009: Shutdown timeout bounds
	if cfg.ShutdownTimeout < 1*time.Second || cfg.ShutdownTimeout > 10*time.Second {
		return validationError(
			"shutdownTimeout",
			"must be between 1s and 10s",
			fmt.Sprintf("%v", cfg.ShutdownTimeout),
		)
	}

	// VR-010: MaxConcurrentOps bounds
	if cfg.MaxConcurrentOps < 1 || cfg.MaxConcurrentOps > 16 {
		return validationError(
			"maxConcurrentOps",
			"must be between 1 and 16",
			fmt.Sprintf("%d", cfg.MaxConcurrentOps),
		)
	}

	// VR-011: Ensure all paths are absolute
	if !filepath.IsAbs(cfg.ConfigDir) {
		absPath, err := filepath.Abs(cfg.ConfigDir)
		if err != nil {
			return fmt.Errorf("failed to resolve config directory path: %w", err)
		}
		cfg.ConfigDir = absPath
	}

	if !filepath.IsAbs(cfg.LogDir) {
		absPath, err := filepath.Abs(cfg.LogDir)
		if err != nil {
			return fmt.Errorf("failed to resolve log directory path: %w", err)
		}
		cfg.LogDir = absPath
	}

	if !filepath.IsAbs(cfg.CacheDir) {
		absPath, err := filepath.Abs(cfg.CacheDir)
		if err != nil {
			return fmt.Errorf("failed to resolve cache directory path: %w", err)
		}
		cfg.CacheDir = absPath
	}

	return nil
}
