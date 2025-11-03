package config

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// validator provides basic validation for Config struct fields.
// Full validation logic will be implemented in later phases.
// See: T029, FR-013, FR-014
type validator struct {
	schema *ConfigSchema
}

// newValidator creates a new validator with the provided schema.
func newValidator(schema *ConfigSchema) *validator {
	return &validator{
		schema: schema,
	}
}

// validate performs basic validation on a Config struct.
// Returns a slice of ValidationErrors (both blocking and non-blocking).
// See: FR-013, FR-014
func (v *validator) validate(cfg *Config) []ValidationError {
	var errors []ValidationError

	// Validate theme
	if err := v.validateEnum(cfg.Theme, []string{"default", "dark", "light", "solarized"}, "theme"); err != nil {
		errors = append(errors, *err)
	}

	// Validate color scheme hex colors
	colorChecks := map[string]string{
		"colorScheme.border":      cfg.ColorScheme.Border,
		"colorScheme.borderFocus": cfg.ColorScheme.BorderFocus,
		"colorScheme.text":        cfg.ColorScheme.Text,
		"colorScheme.textDim":     cfg.ColorScheme.TextDim,
		"colorScheme.background":  cfg.ColorScheme.Background,
		"colorScheme.highlight":   cfg.ColorScheme.Highlight,
		"colorScheme.error":       cfg.ColorScheme.Error,
		"colorScheme.warning":     cfg.ColorScheme.Warning,
		"colorScheme.success":     cfg.ColorScheme.Success,
		"colorScheme.info":        cfg.ColorScheme.Info,
	}
	for field, value := range colorChecks {
		if err := v.validateHexColor(value, field); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validate keybinding profile
	if err := v.validateEnum(cfg.KeybindingProfile, []string{"default", "vim", "emacs"}, "keybindingProfile"); err != nil {
		errors = append(errors, *err)
	}

	// Validate maxConcurrentOps range
	if cfg.MaxConcurrentOps < 1 || cfg.MaxConcurrentOps > 16 {
		errors = append(errors, ValidationError{
			Key:          "maxConcurrentOps",
			Value:        cfg.MaxConcurrentOps,
			Constraint:   "must be between 1 and 16",
			SuggestedFix: "Set maxConcurrentOps to a value between 1 and 16",
			Severity:     "warning",
			DefaultUsed:  4,
		})
	}

	// Validate cacheSize
	if cfg.CacheSize < 0 {
		errors = append(errors, ValidationError{
			Key:          "cacheSize",
			Value:        cfg.CacheSize,
			Constraint:   "must be non-negative",
			SuggestedFix: "Set cacheSize to 0 or higher",
			Severity:     "warning",
			DefaultUsed:  50,
		})
	}

	// Validate timeouts
	if cfg.Timeouts.NetworkRequest < 1*time.Second {
		errors = append(errors, ValidationError{
			Key:          "timeouts.networkRequest",
			Value:        cfg.Timeouts.NetworkRequest,
			Constraint:   "must be at least 1 second",
			SuggestedFix: "Set timeouts.networkRequest to at least 1s",
			Severity:     "warning",
			DefaultUsed:  30 * time.Second,
		})
	}
	if cfg.Timeouts.DotnetCLI < 1*time.Second {
		errors = append(errors, ValidationError{
			Key:          "timeouts.dotnetCLI",
			Value:        cfg.Timeouts.DotnetCLI,
			Constraint:   "must be at least 1 second",
			SuggestedFix: "Set timeouts.dotnetCLI to at least 1s",
			Severity:     "warning",
			DefaultUsed:  60 * time.Second,
		})
	}
	if cfg.Timeouts.FileOperation < 100*time.Millisecond {
		errors = append(errors, ValidationError{
			Key:          "timeouts.fileOperation",
			Value:        cfg.Timeouts.FileOperation,
			Constraint:   "must be at least 100ms",
			SuggestedFix: "Set timeouts.fileOperation to at least 100ms",
			Severity:     "warning",
			DefaultUsed:  5 * time.Second,
		})
	}

	// Validate dotnet verbosity
	if err := v.validateEnum(cfg.DotnetVerbosity, []string{"quiet", "minimal", "normal", "detailed", "diagnostic"}, "dotnetVerbosity"); err != nil {
		errors = append(errors, *err)
	}

	// Validate log level
	if err := v.validateEnum(cfg.LogLevel, []string{"debug", "info", "warn", "error"}, "logLevel"); err != nil {
		errors = append(errors, *err)
	}

	// Validate log format
	if err := v.validateEnum(cfg.LogFormat, []string{"text", "json"}, "logFormat"); err != nil {
		errors = append(errors, *err)
	}

	// Validate log rotation
	if cfg.LogRotation.MaxSize < 1 {
		errors = append(errors, ValidationError{
			Key:          "logRotation.maxSize",
			Value:        cfg.LogRotation.MaxSize,
			Constraint:   "must be at least 1 MB",
			SuggestedFix: "Set logRotation.maxSize to at least 1",
			Severity:     "warning",
			DefaultUsed:  10,
		})
	}
	if cfg.LogRotation.MaxAge < 1 {
		errors = append(errors, ValidationError{
			Key:          "logRotation.maxAge",
			Value:        cfg.LogRotation.MaxAge,
			Constraint:   "must be at least 1 day",
			SuggestedFix: "Set logRotation.maxAge to at least 1",
			Severity:     "warning",
			DefaultUsed:  30,
		})
	}
	if cfg.LogRotation.MaxBackups < 0 {
		errors = append(errors, ValidationError{
			Key:          "logRotation.maxBackups",
			Value:        cfg.LogRotation.MaxBackups,
			Constraint:   "must be non-negative",
			SuggestedFix: "Set logRotation.maxBackups to 0 or higher",
			Severity:     "warning",
			DefaultUsed:  3,
		})
	}

	return errors
}

// validateEnum checks if a value is in the allowed list.
func (v *validator) validateEnum(value string, allowed []string, field string) *ValidationError {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return &ValidationError{
		Key:          field,
		Value:        value,
		Constraint:   fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")),
		SuggestedFix: fmt.Sprintf("Set %s to one of the allowed values", field),
		Severity:     "warning",
		DefaultUsed:  allowed[0], // Use first value as default
	}
}

// validateHexColor checks if a string is a valid hex color (#RRGGBB or #RRGGBBAA).
func (v *validator) validateHexColor(value string, field string) *ValidationError {
	hexColorRegex := regexp.MustCompile(`^#([0-9A-Fa-f]{6}|[0-9A-Fa-f]{8})$`)
	if hexColorRegex.MatchString(value) {
		return nil
	}
	return &ValidationError{
		Key:          field,
		Value:        value,
		Constraint:   "must be valid hex color (#RRGGBB or #RRGGBBAA)",
		SuggestedFix: fmt.Sprintf("Set %s to a valid hex color", field),
		Severity:     "warning",
		DefaultUsed:  "#FFFFFF",
	}
}
