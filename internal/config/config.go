package config

import (
	"context"
	"fmt"
)

// ConfigLoader is the primary interface for loading and managing application configuration.
// It handles loading from multiple sources, merging with precedence, and validation.
//
// Implementation: configLoader struct in this file (implemented in later tasks)
// See: specs/002-config-management/contracts/config_loader.md
type ConfigLoader interface {
	// Load reads configuration from all sources (defaults, file, env vars, CLI flags),
	// merges them according to precedence rules, validates the result, and returns
	// the final configuration.
	//
	// Sources are merged in order of increasing precedence:
	//   1. Hardcoded defaults (lowest precedence)
	//   2. User config file (YAML/TOML)
	//   3. Environment variables (LAZYNUGET_* prefix)
	//   4. CLI flags (highest precedence)
	//
	// Returns:
	//   - *Config: The merged and validated configuration
	//   - error: Blocking errors only (syntax errors, file too large, both formats present)
	//
	// Blocking errors (prevent application startup):
	//   - Config file syntax error (invalid YAML/TOML)
	//   - Config file exceeds 10 MB size limit
	//   - Both YAML and TOML config files exist
	//   - Explicitly specified config file (via --config) not found or unreadable
	//
	// Non-blocking errors (logged as warnings, fall back to defaults):
	//   - Setting value out of range (e.g., maxConcurrentOps: 999)
	//   - Invalid setting value (e.g., invalid hex color, malformed duration)
	//   - Unknown config keys (ignored with warning)
	//   - Encrypted value decryption failure
	//   - Keybinding conflicts (first encountered wins, others warned)
	//
	// Performance: Must complete in <500ms for typical config files (<100 KB)
	// See: FR-001, FR-002, FR-009, FR-010, FR-011, FR-012, FR-013, FR-014
	Load(ctx context.Context, opts LoadOptions) (*Config, error)

	// Validate checks a configuration against the schema without loading from sources.
	// Useful for the --validate-config CLI flag and testing.
	//
	// Returns:
	//   - []ValidationError: All validation errors found (both blocking and non-blocking)
	//   - error: System errors only (e.g., unable to access schema)
	//
	// See: FR-056
	Validate(ctx context.Context, cfg *Config) ([]ValidationError, error)

	// GetDefaults returns the hardcoded default configuration.
	// This is the base configuration used when no other sources are available.
	//
	// Returns: Complete Config struct with all default values populated
	// See: FR-001
	GetDefaults() *Config

	// PrintConfig returns a human-readable representation of the configuration
	// with provenance information (which source provided each setting).
	// Useful for the --print-config CLI flag for debugging.
	//
	// Returns: Multi-line string showing each setting, its value, and source
	// See: FR-055
	PrintConfig(cfg *Config) string
}

// LoadOptions configures the behavior of the config loading process.
// See: specs/002-config-management/contracts/config_loader.md
type LoadOptions struct {
	// ConfigFilePath explicitly specifies the config file path.
	// If empty, uses platform-specific default location.
	// Can be overridden by LAZYNUGET_CONFIG environment variable.
	// Maps to: --config CLI flag (FR-007) or LAZYNUGET_CONFIG env var (FR-008)
	ConfigFilePath string

	// CLIFlags provides command-line flag values to override other sources.
	// Maps to: various --flag arguments (FR-053, FR-054)
	CLIFlags CLIFlags

	// EnvVarPrefix specifies the prefix for environment variable overrides.
	// Default: "LAZYNUGET_"
	// Maps to: FR-050
	EnvVarPrefix string

	// StrictMode when true treats semantic validation errors as blocking.
	// Default: false (semantic errors log warnings and fall back to defaults)
	// This option is primarily for testing and CI environments.
	StrictMode bool

	// Logger for logging validation warnings and debug information.
	// If nil, logging is silently skipped (useful for testing).
	Logger Logger
}

// CLIFlags contains command-line flag values that override other config sources.
// See: specs/002-config-management/contracts/config_loader.md
type CLIFlags struct {
	// ConfigFile path (--config flag) - handled separately by LoadOptions.ConfigFilePath
	// Left here for reference

	// Common setting overrides
	LogLevel       string // --log-level flag (FR-054)
	NonInteractive bool   // --non-interactive flag (FR-054)
	NoColor        bool   // --no-color flag (FR-054)

	// Future: Add more flags as needed for specific settings
}

// Logger interface for configuration system logging.
// Allows the config package to log without depending on the application's logging implementation.
// See: specs/002-config-management/contracts/config_loader.md
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// configLoader is the concrete implementation of ConfigLoader interface.
// See: T026
type configLoader struct {
	schema    *ConfigSchema
	validator *validator
}

// NewConfigLoader creates a new ConfigLoader instance.
// See: T026
func NewConfigLoader() ConfigLoader {
	schema := GetConfigSchema()
	return &configLoader{
		schema:    schema,
		validator: newValidator(schema),
	}
}

// Load implements ConfigLoader.Load()
// Currently handles missing config file case by returning defaults.
// Full file loading will be implemented in later phases.
// See: T027, T031, FR-001
func (cl *configLoader) Load(ctx context.Context, opts LoadOptions) (*Config, error) {
	// Log that we're using default configuration (T031)
	if opts.Logger != nil {
		opts.Logger.Info("No config file found, using default configuration")
	}

	// For Phase 3 (US1), we only handle the missing config file case
	// Full file loading, env vars, and CLI flag merging will be implemented in Phases 4-6
	cfg := GetDefaultConfig()

	// Validate the config
	validationErrors := cl.validator.validate(cfg)

	// Log any validation warnings (semantic errors, non-blocking)
	if opts.Logger != nil && len(validationErrors) > 0 {
		for _, ve := range validationErrors {
			if ve.Severity == "warning" {
				opts.Logger.Warn("Config validation warning: %s", ve.Error())
			}
		}
	}

	return cfg, nil
}

// Validate implements ConfigLoader.Validate()
// See: T030, FR-056
func (cl *configLoader) Validate(ctx context.Context, cfg *Config) ([]ValidationError, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}

	// Run validation
	validationErrors := cl.validator.validate(cfg)

	return validationErrors, nil
}

// GetDefaults implements ConfigLoader.GetDefaults()
// See: T028, FR-001
func (cl *configLoader) GetDefaults() *Config {
	return GetDefaultConfig()
}

// PrintConfig implements ConfigLoader.PrintConfig()
// STUB: Will be fully implemented in later phase
// See: FR-055
func (cl *configLoader) PrintConfig(cfg *Config) string {
	// TODO: Implement in Phase 9 (polish)
	return "Config printing not yet implemented"
}
