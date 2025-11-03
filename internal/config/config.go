package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load creates configuration by merging sources with precedence: CLI > Env > File > Defaults
// This is the main entry point for configuration loading
func Load() (*AppConfig, error) {
	// Start with defaults
	cfg := DefaultConfig()

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

	// Validate the merged configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
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
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return err // Return NotExist error (will be handled gracefully)
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
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

	if err := yaml.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
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
		return fmt.Errorf("invalid log level %q: must be debug, info, warn, or error", cfg.LogLevel)
	}

	// VR-007: Config path must exist if specified
	if cfg.ConfigPath != "" {
		if _, err := os.Stat(cfg.ConfigPath); os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", cfg.ConfigPath)
		}
	}

	// VR-008: Startup timeout bounds
	if cfg.StartupTimeout < 1*time.Second || cfg.StartupTimeout > 30*time.Second {
		return fmt.Errorf("startupTimeout must be between 1s and 30s, got %v", cfg.StartupTimeout)
	}

	// VR-009: Shutdown timeout bounds
	if cfg.ShutdownTimeout < 1*time.Second || cfg.ShutdownTimeout > 10*time.Second {
		return fmt.Errorf("shutdownTimeout must be between 1s and 10s, got %v", cfg.ShutdownTimeout)
	}

	// VR-010: MaxConcurrentOps bounds
	if cfg.MaxConcurrentOps < 1 || cfg.MaxConcurrentOps > 16 {
		return fmt.Errorf("maxConcurrentOps must be between 1 and 16, got %d", cfg.MaxConcurrentOps)
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
