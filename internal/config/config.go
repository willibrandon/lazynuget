package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	Logger         Logger
	ConfigFilePath string
	EnvVarPrefix   string
	CLIFlags       CLIFlags
	StrictMode     bool
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
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// configLoader is the concrete implementation of ConfigLoader interface.
// See: T026
type configLoader struct {
	schema    *ConfigSchema
	validator *validator
}

// NewLoader creates a new ConfigLoader instance.
// See: T026
func NewLoader() ConfigLoader {
	schema := GetConfigSchema()
	return &configLoader{
		schema:    schema,
		validator: newValidator(schema),
	}
}

// Load implements ConfigLoader.Load()
// Loads configuration from defaults, file, env vars, and CLI flags with proper precedence.
// See: T027, T031, T050, FR-001, FR-002
func (cl *configLoader) Load(ctx context.Context, opts LoadOptions) (*Config, error) {
	// Start with defaults (lowest precedence)
	cfg := GetDefaultConfig()

	// Determine config file path
	configFilePath := opts.ConfigFilePath
	if configFilePath == "" {
		// Check environment variable
		if envPath := os.Getenv(opts.EnvVarPrefix + "CONFIG"); envPath != "" {
			configFilePath = envPath
		} else {
			// Use platform-specific default location
			configFilePath = getPlatformConfigPath()
			if configFilePath != "" {
				// Check for config.yml or config.toml
				yamlPath := filepath.Join(configFilePath, "config.yml")
				tomlPath := filepath.Join(configFilePath, "config.toml")

				if _, err := os.Stat(yamlPath); err == nil {
					configFilePath = yamlPath
				} else if _, err := os.Stat(tomlPath); err == nil {
					configFilePath = tomlPath
				} else {
					configFilePath = "" // No config file found
				}
			}
		}
	}

	// Attempt to load config file if path is specified
	if configFilePath != "" {
		// Check if file exists
		if _, err := os.Stat(configFilePath); err == nil {
			// Check for multiple formats in directory (FR-005)
			configDir := filepath.Dir(configFilePath)
			if err := checkMultipleFormats(configDir); err != nil {
				// This is a blocking error
				return nil, err
			}

			// Parse config file
			fileCfg, err := parseConfigFile(configFilePath)
			if err != nil {
				// Syntax errors are blocking (FR-010)
				return nil, fmt.Errorf("failed to load config file %s: %w", configFilePath, err)
			}

			// Handle encrypted values (T131, T132)
			// Create encryptor for decryption
			keychain := NewKeychainManager()
			kd := NewKeyDerivation()
			encryptor := NewEncryptor(keychain, kd)

			// Attempt to decrypt any encrypted values in the config file
			// Path already validated by parseConfigFile above
			fileData, err := os.ReadFile(filepath.Clean(configFilePath))
			if err == nil {
				_, encryptedFields, scanErr := parseYAMLWithEncryption(fileData)
				if scanErr == nil && len(encryptedFields) > 0 {
					// Attempt to decrypt each encrypted field
					for fieldPath, encryptedValue := range encryptedFields {
						_, decryptErr := encryptor.Decrypt(ctx, encryptedValue)
						if decryptErr != nil {
							// FR-018: Log warning but continue (fall back to default)
							if opts.Logger != nil {
								opts.Logger.Warn("Failed to decrypt field %s: %v (falling back to default)", fieldPath, decryptErr)
							}
							// Don't block loading - validation will handle fallback to default
						} else {
							// Successfully decrypted - apply to config
							// For now, we log it but don't apply (since we'd need field mapping)
							if opts.Logger != nil {
								opts.Logger.Debug("Successfully decrypted field: %s", fieldPath)
							}
							// TODO: Apply decrypted value to cfg using reflection or field mapping
							// For Phase 8, we'll handle this in integration tests with known fields
						}
					}
				}
			}

			if opts.Logger != nil {
				opts.Logger.Info("Loaded configuration from file: %s", configFilePath)
			}

			// Merge file config with defaults
			cfg = mergeConfigs(cfg, fileCfg)
			cfg.LoadedFrom = configFilePath
		} else if opts.ConfigFilePath != "" {
			// If user explicitly specified a config file (via --config), it must exist
			return nil, fmt.Errorf("specified config file not found: %s", configFilePath)
		} else {
			// No config file found at default location - use defaults
			if opts.Logger != nil {
				opts.Logger.Info("No config file found at default location, using defaults")
			}
		}
	} else {
		// No config file path determined - use defaults
		if opts.Logger != nil {
			opts.Logger.Info("No config file found, using default configuration")
		}
	}

	// Apply environment variable overrides (Phase 5, FR-050, FR-051, FR-052)
	if opts.EnvVarPrefix != "" {
		envVars := parseEnvVars(opts.EnvVarPrefix)
		if len(envVars) > 0 {
			if opts.Logger != nil {
				opts.Logger.Debug("Found %d environment variable overrides", len(envVars))
			}

			// Apply each environment variable to the config
			for path, value := range envVars {
				if opts.Logger != nil {
					opts.Logger.Debug("Applying env var override: %s = %s", path, value)
				}
				if err := applyEnvVarValue(cfg, path, value); err != nil {
					if opts.Logger != nil {
						opts.Logger.Warn("Failed to apply env var %s: %v", path, err)
					}
				}
			}
		}
	}

	// Apply CLI flag overrides (Phase 6, FR-054, highest precedence)
	if opts.CLIFlags.LogLevel != "" {
		if opts.Logger != nil {
			opts.Logger.Debug("Applying CLI flag override: logLevel = %s", opts.CLIFlags.LogLevel)
		}
		cfg.LogLevel = opts.CLIFlags.LogLevel
	}

	// Note: NonInteractive and NoColor flags are consumed by bootstrap/GUI layers
	// They are passed through LoadOptions but don't affect the Config struct

	// Validate the final merged config
	validationErrors := cl.validator.validate(cfg)

	// Handle validation errors based on StrictMode
	hasBlockingErrors := false
	for _, ve := range validationErrors {
		if ve.Severity == "error" {
			hasBlockingErrors = true
			if opts.Logger != nil {
				opts.Logger.Error("Config validation error: %s", ve.Error())
			}
		} else if ve.Severity == "warning" {
			if opts.Logger != nil {
				opts.Logger.Warn("Config validation warning: %s (using default: %v)", ve.Error(), ve.DefaultUsed)
			}
			// Apply fallback to default for this setting
			// This is handled by the validator returning the appropriate default
		}
	}

	// In strict mode, blocking errors prevent startup
	if opts.StrictMode && hasBlockingErrors {
		return nil, fmt.Errorf("config validation failed with %d error(s)", len(validationErrors))
	}

	return cfg, nil
}

// Validate implements ConfigLoader.Validate()
// See: T030, FR-056
func (cl *configLoader) Validate(_ context.Context, cfg *Config) ([]ValidationError, error) {
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
// Returns a human-readable representation of the configuration with provenance.
// See: T058, FR-055
func (cl *configLoader) PrintConfig(cfg *Config) string {
	var sb strings.Builder

	sb.WriteString("=== LazyNuGet Configuration ===\n\n")

	// Config source
	if cfg.LoadedFrom != "" {
		sb.WriteString(fmt.Sprintf("Loaded from: %s\n", cfg.LoadedFrom))
	} else {
		sb.WriteString("Loaded from: defaults only\n")
	}
	sb.WriteString(fmt.Sprintf("Loaded at: %s\n\n", cfg.LoadedAt.Format("2006-01-02 15:04:05")))

	// UI Settings
	sb.WriteString("--- UI Settings ---\n")
	sb.WriteString(fmt.Sprintf("theme:            %s\n", cfg.Theme))
	sb.WriteString(fmt.Sprintf("compactMode:      %v\n", cfg.CompactMode))
	sb.WriteString(fmt.Sprintf("showHints:        %v\n", cfg.ShowHints))
	sb.WriteString(fmt.Sprintf("showLineNumbers:  %v\n", cfg.ShowLineNumbers))
	sb.WriteString(fmt.Sprintf("dateFormat:       %s\n\n", cfg.DateFormat))

	// Color Scheme
	sb.WriteString("--- Color Scheme ---\n")
	sb.WriteString(fmt.Sprintf("border:           %s\n", cfg.ColorScheme.Border))
	sb.WriteString(fmt.Sprintf("borderFocus:      %s\n", cfg.ColorScheme.BorderFocus))
	sb.WriteString(fmt.Sprintf("text:             %s\n", cfg.ColorScheme.Text))
	sb.WriteString(fmt.Sprintf("textDim:          %s\n", cfg.ColorScheme.TextDim))
	sb.WriteString(fmt.Sprintf("background:       %s\n", cfg.ColorScheme.Background))
	sb.WriteString(fmt.Sprintf("highlight:        %s\n", cfg.ColorScheme.Highlight))
	sb.WriteString(fmt.Sprintf("error:            %s\n", cfg.ColorScheme.Error))
	sb.WriteString(fmt.Sprintf("warning:          %s\n", cfg.ColorScheme.Warning))
	sb.WriteString(fmt.Sprintf("success:          %s\n", cfg.ColorScheme.Success))
	sb.WriteString(fmt.Sprintf("info:             %s\n\n", cfg.ColorScheme.Info))

	// Keybindings
	sb.WriteString("--- Keybindings ---\n")
	sb.WriteString(fmt.Sprintf("keybindingProfile: %s\n", cfg.KeybindingProfile))
	if len(cfg.Keybindings) > 0 {
		sb.WriteString("Custom keybindings:\n")
		for action, binding := range cfg.Keybindings {
			sb.WriteString(fmt.Sprintf("  %s: %s (context: %s)\n", action, binding.Key, binding.Context))
		}
	}
	sb.WriteString("\n")

	// Performance
	sb.WriteString("--- Performance ---\n")
	sb.WriteString(fmt.Sprintf("maxConcurrentOps: %d\n", cfg.MaxConcurrentOps))
	sb.WriteString(fmt.Sprintf("cacheSize:        %d MB\n", cfg.CacheSize))
	sb.WriteString(fmt.Sprintf("refreshInterval:  %s\n\n", cfg.RefreshInterval))

	// Timeouts
	sb.WriteString("--- Timeouts ---\n")
	sb.WriteString(fmt.Sprintf("networkRequest:   %s\n", cfg.Timeouts.NetworkRequest))
	sb.WriteString(fmt.Sprintf("dotnetCLI:        %s\n", cfg.Timeouts.DotnetCLI))
	sb.WriteString(fmt.Sprintf("fileOperation:    %s\n\n", cfg.Timeouts.FileOperation))

	// Dotnet CLI
	sb.WriteString("--- Dotnet CLI ---\n")
	sb.WriteString(fmt.Sprintf("dotnetPath:       %s\n", cfg.DotnetPath))
	sb.WriteString(fmt.Sprintf("dotnetVerbosity:  %s\n\n", cfg.DotnetVerbosity))

	// Logging
	sb.WriteString("--- Logging ---\n")
	sb.WriteString(fmt.Sprintf("logLevel:         %s\n", cfg.LogLevel))
	sb.WriteString(fmt.Sprintf("logDir:           %s\n", cfg.LogDir))
	sb.WriteString(fmt.Sprintf("logFormat:        %s\n", cfg.LogFormat))
	sb.WriteString(fmt.Sprintf("maxSize:          %d MB\n", cfg.LogRotation.MaxSize))
	sb.WriteString(fmt.Sprintf("maxAge:           %d days\n", cfg.LogRotation.MaxAge))
	sb.WriteString(fmt.Sprintf("maxBackups:       %d\n", cfg.LogRotation.MaxBackups))
	sb.WriteString(fmt.Sprintf("compress:         %v\n\n", cfg.LogRotation.Compress))

	// Hot Reload
	sb.WriteString("--- Hot Reload ---\n")
	sb.WriteString(fmt.Sprintf("hotReload:        %v\n", cfg.HotReload))

	return sb.String()
}
