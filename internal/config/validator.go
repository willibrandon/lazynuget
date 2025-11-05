package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// validator provides validation for Config struct fields.
// Validates all settings according to their constraints and applies fallback defaults.
// See: T029, T052-T056, FR-013, FR-014
type validator struct {
	schema *ConfigSchema
}

// newValidator creates a new validator with the provided schema.
func newValidator(schema *ConfigSchema) *validator {
	return &validator{
		schema: schema,
	}
}

// validate performs comprehensive validation on a Config struct and applies fallback defaults.
// Returns a slice of ValidationErrors (both blocking and non-blocking).
// Mutates cfg to apply fallback defaults for invalid values.
// See: T052-T056, FR-011, FR-012, FR-013
func (v *validator) validate(cfg *Config) []ValidationError {
	var errors []ValidationError
	defaults := GetDefaultConfig()

	// Validate theme (T052)
	if err := v.validateEnum(&cfg.Theme, []string{"default", "dark", "light", "solarized"}, "theme", defaults.Theme); err != nil {
		errors = append(errors, *err)
	}

	// Validate color scheme hex colors (T052, T053)
	v.validateAndFixHexColor(&cfg.ColorScheme.Border, "colorScheme.border", defaults.ColorScheme.Border, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.BorderFocus, "colorScheme.borderFocus", defaults.ColorScheme.BorderFocus, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.Text, "colorScheme.text", defaults.ColorScheme.Text, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.TextDim, "colorScheme.textDim", defaults.ColorScheme.TextDim, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.Background, "colorScheme.background", defaults.ColorScheme.Background, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.Highlight, "colorScheme.highlight", defaults.ColorScheme.Highlight, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.Error, "colorScheme.error", defaults.ColorScheme.Error, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.Warning, "colorScheme.warning", defaults.ColorScheme.Warning, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.Success, "colorScheme.success", defaults.ColorScheme.Success, &errors)
	v.validateAndFixHexColor(&cfg.ColorScheme.Info, "colorScheme.info", defaults.ColorScheme.Info, &errors)

	// Validate keybinding profile (T052)
	if err := v.validateEnum(&cfg.KeybindingProfile, []string{"default", "vim", "emacs"}, "keybindingProfile", defaults.KeybindingProfile); err != nil {
		errors = append(errors, *err)
	}

	// Validate keybinding conflicts (T057, FR-028)
	if keybindingErrors := v.validateKeybindingConflicts(cfg); len(keybindingErrors) > 0 {
		errors = append(errors, keybindingErrors...)
	}

	// Validate maxConcurrentOps range (T052, T053)
	if cfg.MaxConcurrentOps < 1 || cfg.MaxConcurrentOps > 16 {
		errors = append(errors, ValidationError{
			Key:          "maxConcurrentOps",
			Value:        cfg.MaxConcurrentOps,
			Constraint:   "must be between 1 and 16",
			SuggestedFix: "Set maxConcurrentOps to a value between 1 and 16",
			Severity:     "warning",
			DefaultUsed:  defaults.MaxConcurrentOps,
		})
		cfg.MaxConcurrentOps = defaults.MaxConcurrentOps // Apply fallback (T056)
	}

	// Validate cacheSize (T052)
	if cfg.CacheSize < 0 {
		errors = append(errors, ValidationError{
			Key:          "cacheSize",
			Value:        cfg.CacheSize,
			Constraint:   "must be non-negative",
			SuggestedFix: "Set cacheSize to 0 or higher",
			Severity:     "warning",
			DefaultUsed:  defaults.CacheSize,
		})
		cfg.CacheSize = defaults.CacheSize // Apply fallback (T056)
	}

	// Validate refreshInterval (T052, T053)
	if cfg.RefreshInterval < 5*time.Second {
		errors = append(errors, ValidationError{
			Key:          "refreshInterval",
			Value:        cfg.RefreshInterval,
			Constraint:   "must be at least 5 seconds",
			SuggestedFix: "Set refreshInterval to at least 5s",
			Severity:     "warning",
			DefaultUsed:  defaults.RefreshInterval,
		})
		cfg.RefreshInterval = defaults.RefreshInterval // Apply fallback (T056)
	}

	// Validate timeouts (T052, T053)
	if cfg.Timeouts.NetworkRequest < 1*time.Second {
		errors = append(errors, ValidationError{
			Key:          "timeouts.networkRequest",
			Value:        cfg.Timeouts.NetworkRequest,
			Constraint:   "must be at least 1 second",
			SuggestedFix: "Set timeouts.networkRequest to at least 1s",
			Severity:     "warning",
			DefaultUsed:  defaults.Timeouts.NetworkRequest,
		})
		cfg.Timeouts.NetworkRequest = defaults.Timeouts.NetworkRequest // Apply fallback (T056)
	}
	if cfg.Timeouts.DotnetCLI < 1*time.Second {
		errors = append(errors, ValidationError{
			Key:          "timeouts.dotnetCLI",
			Value:        cfg.Timeouts.DotnetCLI,
			Constraint:   "must be at least 1 second",
			SuggestedFix: "Set timeouts.dotnetCLI to at least 1s",
			Severity:     "warning",
			DefaultUsed:  defaults.Timeouts.DotnetCLI,
		})
		cfg.Timeouts.DotnetCLI = defaults.Timeouts.DotnetCLI // Apply fallback (T056)
	}
	if cfg.Timeouts.FileOperation < 100*time.Millisecond {
		errors = append(errors, ValidationError{
			Key:          "timeouts.fileOperation",
			Value:        cfg.Timeouts.FileOperation,
			Constraint:   "must be at least 100ms",
			SuggestedFix: "Set timeouts.fileOperation to at least 100ms",
			Severity:     "warning",
			DefaultUsed:  defaults.Timeouts.FileOperation,
		})
		cfg.Timeouts.FileOperation = defaults.Timeouts.FileOperation // Apply fallback (T056)
	}

	// Validate dotnet verbosity (T052)
	if err := v.validateEnum(&cfg.DotnetVerbosity, []string{"quiet", "minimal", "normal", "detailed", "diagnostic"}, "dotnetVerbosity", defaults.DotnetVerbosity); err != nil {
		errors = append(errors, *err)
	}

	// Validate log level (T052)
	if err := v.validateEnum(&cfg.LogLevel, []string{"debug", "info", "warn", "error"}, "logLevel", defaults.LogLevel); err != nil {
		errors = append(errors, *err)
	}

	// Validate log format (T052)
	if err := v.validateEnum(&cfg.LogFormat, []string{"text", "json"}, "logFormat", defaults.LogFormat); err != nil {
		errors = append(errors, *err)
	}

	// Validate date format (T052, T053)
	if err := v.validateDateFormat(cfg.DateFormat, "dateFormat"); err != nil {
		errors = append(errors, *err)
		cfg.DateFormat = defaults.DateFormat // Apply fallback (T056)
	}

	// Validate log rotation (T052)
	if cfg.LogRotation.MaxSize < 1 {
		errors = append(errors, ValidationError{
			Key:          "logRotation.maxSize",
			Value:        cfg.LogRotation.MaxSize,
			Constraint:   "must be at least 1 MB",
			SuggestedFix: "Set logRotation.maxSize to at least 1",
			Severity:     "warning",
			DefaultUsed:  defaults.LogRotation.MaxSize,
		})
		cfg.LogRotation.MaxSize = defaults.LogRotation.MaxSize // Apply fallback (T056)
	}
	if cfg.LogRotation.MaxAge < 1 {
		errors = append(errors, ValidationError{
			Key:          "logRotation.maxAge",
			Value:        cfg.LogRotation.MaxAge,
			Constraint:   "must be at least 1 day",
			SuggestedFix: "Set logRotation.maxAge to at least 1",
			Severity:     "warning",
			DefaultUsed:  defaults.LogRotation.MaxAge,
		})
		cfg.LogRotation.MaxAge = defaults.LogRotation.MaxAge // Apply fallback (T056)
	}
	if cfg.LogRotation.MaxBackups < 0 {
		errors = append(errors, ValidationError{
			Key:          "logRotation.maxBackups",
			Value:        cfg.LogRotation.MaxBackups,
			Constraint:   "must be non-negative",
			SuggestedFix: "Set logRotation.maxBackups to 0 or higher",
			Severity:     "warning",
			DefaultUsed:  defaults.LogRotation.MaxBackups,
		})
		cfg.LogRotation.MaxBackups = defaults.LogRotation.MaxBackups // Apply fallback (T056)
	}

	// Validate and normalize paths (T052, T053)
	if cfg.LogDir != "" {
		// Get platform-specific path resolver
		platformInfo, err := platform.New()
		if err == nil {
			pathResolver, err := platform.NewPathResolver(platformInfo)
			if err == nil {
				// Validate the path
				if validateErr := pathResolver.Validate(cfg.LogDir); validateErr != nil {
					errors = append(errors, ValidationError{
						Key:          "logDir",
						Value:        cfg.LogDir,
						Constraint:   "must be a valid path for the current platform",
						SuggestedFix: fmt.Sprintf("Path validation failed: %v", validateErr),
						Severity:     "warning",
						DefaultUsed:  defaults.LogDir,
					})
					cfg.LogDir = defaults.LogDir // Apply fallback (T056)
				} else {
					// Normalize the path (T053)
					cfg.LogDir = pathResolver.Normalize(cfg.LogDir)
				}
			}
		}
	}

	return errors
}

// validateEnum checks if a value is in the allowed list and applies fallback default if invalid.
// See: T053, T056, FR-012
func (v *validator) validateEnum(value *string, allowed []string, field, defaultValue string) *ValidationError {
	if slices.Contains(allowed, *value) {
		return nil
	}

	// Invalid value - apply fallback default (T056)
	originalValue := *value
	*value = defaultValue

	return &ValidationError{
		Key:          field,
		Value:        originalValue,
		Constraint:   fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
		SuggestedFix: fmt.Sprintf("Set %s to one of the allowed values", field),
		Severity:     "warning",
		DefaultUsed:  defaultValue,
	}
}

// validateAndFixHexColor validates a hex color and applies fallback default if invalid.
// See: T053, T056, FR-012
func (v *validator) validateAndFixHexColor(value *string, field, defaultValue string, errors *[]ValidationError) {
	hexColorRegex := regexp.MustCompile(`^#([0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$`)
	if hexColorRegex.MatchString(*value) {
		return
	}

	// Invalid hex color - apply fallback default (T056)
	originalValue := *value
	*value = defaultValue

	*errors = append(*errors, ValidationError{
		Key:          field,
		Value:        originalValue,
		Constraint:   "must be valid hex color (#RRGGBB or #RRGGBBAA)",
		SuggestedFix: fmt.Sprintf("Set %s to a valid hex color", field),
		Severity:     "warning",
		DefaultUsed:  defaultValue,
	})
}

// validateDateFormat validates a Go time format string.
// See: T053, FR-012
func (v *validator) validateDateFormat(format, field string) *ValidationError {
	// Test the format with a known time (time.Format doesn't panic, so no recovery needed)
	testTime := time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	result := testTime.Format(format)

	// If we can format and get a non-empty result, format is valid
	if result != "" {
		return nil
	}

	return &ValidationError{
		Key:          field,
		Value:        format,
		Constraint:   "must be valid Go time format string",
		SuggestedFix: "Use a valid Go time format (e.g., \"2006-01-02 15:04:05\")",
		Severity:     "warning",
		DefaultUsed:  "2006-01-02 15:04:05",
	}
}

// validateKeybindingConflicts detects duplicate key assignments in keybindings.
// See: T057, FR-028
func (v *validator) validateKeybindingConflicts(cfg *Config) []ValidationError {
	var errors []ValidationError

	// Track keys used in each context
	keysByContext := make(map[string]map[string]string) // context -> key -> action

	for action, binding := range cfg.Keybindings {
		context := binding.Context
		key := binding.Key

		// Initialize context map if needed
		if keysByContext[context] == nil {
			keysByContext[context] = make(map[string]string)
		}

		// Check for conflict
		if existingAction, exists := keysByContext[context][key]; exists {
			// Conflict detected - first binding wins, warn about duplicate
			errors = append(errors, ValidationError{
				Key:          fmt.Sprintf("keybindings.%s", action),
				Value:        fmt.Sprintf("key '%s' in context '%s'", key, context),
				Constraint:   fmt.Sprintf("key already assigned to action '%s'", existingAction),
				SuggestedFix: fmt.Sprintf("Assign a different key to '%s' or use different context", action),
				Severity:     "warning",
				DefaultUsed:  "conflicting binding ignored",
			})
			// Note: We don't modify the config here - first binding wins by default
		} else {
			// Record this key assignment
			keysByContext[context][key] = action
		}
	}

	return errors
}
