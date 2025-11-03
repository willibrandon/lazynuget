package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// parseEnvVars scans all environment variables with the given prefix
// and returns a map of setting paths to values.
// Per FR-050: Environment variables use LAZYNUGET_ prefix
// Per FR-051: Nested settings use underscore notation (LAZYNUGET_COLOR_SCHEME_BORDER)
func parseEnvVars(prefix string) map[string]string {
	result := make(map[string]string)

	// Get all environment variables
	for _, env := range os.Environ() {
		// Split into key=value
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Check if it starts with our prefix (case-insensitive per FR-052)
		if !strings.HasPrefix(strings.ToUpper(key), strings.ToUpper(prefix)) {
			continue
		}

		// Remove prefix to get the setting path
		settingPath := strings.TrimPrefix(strings.ToUpper(key), strings.ToUpper(prefix))
		if settingPath == "" {
			continue
		}

		// Convert underscore notation to dot notation
		// LAZYNUGET_COLOR_SCHEME_BORDER -> colorScheme.border
		dotPath := convertEnvVarPathToDotNotation(settingPath)

		result[dotPath] = value
	}

	return result
}

// convertEnvVarPathToDotNotation converts underscore-separated env var path
// to dot-notation config path with proper camelCase.
// Per FR-051: LAZYNUGET_COLOR_SCHEME_BORDER -> colorScheme.border
// Special handling: LOG_LEVEL -> logLevel (multi-word field names)
func convertEnvVarPathToDotNotation(envPath string) string {
	// Split by underscores
	parts := splitEnvVarPath(envPath)

	// Convert to camelCase by joining parts into a single identifier
	// then checking if it matches a known nested structure
	fullPath := toCamelCaseMulti(parts)

	return fullPath
}

// splitEnvVarPath splits an environment variable path into components.
// Per FR-051: COLOR_SCHEME_BORDER -> ["COLOR", "SCHEME", "BORDER"]
// Handles special cases like LOG_ROTATION_MAX_SIZE -> ["LOG_ROTATION", "MAX_SIZE"]
func splitEnvVarPath(envPath string) []string {
	// For now, simple underscore split
	// In the future, we may need special handling for known multi-word settings
	return strings.Split(envPath, "_")
}

// toCamelCaseMulti converts a slice of uppercase words to camelCase path,
// detecting known nested structures.
// Examples:
//
//	["LOG", "LEVEL"] -> "logLevel"
//	["COLOR", "SCHEME", "BORDER"] -> "colorScheme.border"
//	["LOG", "ROTATION", "MAX", "SIZE"] -> "logRotation.maxSize"
func toCamelCaseMulti(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	// Known nested structures (parent.child format)
	knownNested := map[string][]string{
		"colorScheme": {"COLOR", "SCHEME"},
		"timeouts":    {"TIMEOUTS"},
		"logRotation": {"LOG", "ROTATION"},
		"keybindings": {"KEYBINDINGS"},
	}

	// Check if we have a known nested structure at the beginning
	for parentCamel, parentParts := range knownNested {
		if len(parts) > len(parentParts) && matchesParts(parts[:len(parentParts)], parentParts) {
			// Found a nested structure
			// Convert remaining parts to camelCase for the child field
			childParts := parts[len(parentParts):]
			childCamel := joinCamelCase(childParts)
			return parentCamel + "." + childCamel
		}
	}

	// No nested structure found - convert all parts to a single camelCase identifier
	return joinCamelCase(parts)
}

// matchesParts checks if actualParts matches expectedParts (case-insensitive)
func matchesParts(actualParts, expectedParts []string) bool {
	if len(actualParts) != len(expectedParts) {
		return false
	}
	for i := range actualParts {
		if !strings.EqualFold(actualParts[i], expectedParts[i]) {
			return false
		}
	}
	return true
}

// joinCamelCase converts a slice of uppercase words to a single camelCase identifier.
// Examples:
//
//	["LOG", "LEVEL"] -> "logLevel"
//	["MAX", "SIZE"] -> "maxSize"
//	["BORDER"] -> "border"
func joinCamelCase(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	result := strings.ToLower(parts[0])
	for i := 1; i < len(parts); i++ {
		// Capitalize first letter of each subsequent word
		word := parts[i]
		if word != "" {
			result += strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}

	return result
}

// applyEnvVarValue attempts to set a field in the config struct based on the
// dot-notation path and string value from an environment variable.
// Per FR-052: Supports type conversion for bool/int/duration/string
// Per FR-012: Invalid values fall back to defaults (handled by caller)
func applyEnvVarValue(cfg *Config, path, value string) error {
	// Split path into components
	parts := strings.Split(path, ".")

	// Handle top-level and nested settings
	switch len(parts) {
	case 1:
		// Top-level setting
		return applyTopLevelSetting(cfg, parts[0], value)
	case 2:
		// Nested setting (e.g., colorScheme.border)
		return applyNestedSetting(cfg, parts[0], parts[1], value)
	case 3:
		// Double-nested setting (e.g., logRotation.maxSize)
		return applyDoubleNestedSetting(cfg, parts[0], parts[1], parts[2], value)
	default:
		// Unsupported nesting depth
		return nil
	}
}

// applyTopLevelSetting sets a top-level config field from an env var string value
func applyTopLevelSetting(cfg *Config, field, value string) error {
	switch field {
	case "version":
		cfg.Version = value
	case "loadedFrom":
		cfg.LoadedFrom = value
	case "theme":
		cfg.Theme = value
	case "compactMode":
		if b, err := parseBool(value); err == nil {
			cfg.CompactMode = b
		}
	case "showHints":
		if b, err := parseBool(value); err == nil {
			cfg.ShowHints = b
		}
	case "showLineNumbers":
		if b, err := parseBool(value); err == nil {
			cfg.ShowLineNumbers = b
		}
	case "dateFormat":
		cfg.DateFormat = value
	case "keybindingProfile":
		cfg.KeybindingProfile = value
	case "maxConcurrentOps":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.MaxConcurrentOps = i
		}
	case "cacheSize":
		if i, err := strconv.Atoi(value); err == nil {
			cfg.CacheSize = i
		}
	case "refreshInterval":
		if d, err := time.ParseDuration(value); err == nil {
			cfg.RefreshInterval = d
		}
	case "dotnetPath":
		cfg.DotnetPath = value
	case "dotnetVerbosity":
		cfg.DotnetVerbosity = value
	case "logLevel":
		cfg.LogLevel = value
	case "logDir":
		cfg.LogDir = value
	case "logFormat":
		cfg.LogFormat = value
	case "hotReload":
		if b, err := parseBool(value); err == nil {
			cfg.HotReload = b
		}
	}

	return nil
}

// applyNestedSetting sets a nested config field (e.g., colorScheme.border)
func applyNestedSetting(cfg *Config, parent, field, value string) error {
	switch parent {
	case "colorScheme":
		switch field {
		case "border":
			cfg.ColorScheme.Border = value
		case "error":
			cfg.ColorScheme.Error = value
		case "warning":
			cfg.ColorScheme.Warning = value
		case "success":
			cfg.ColorScheme.Success = value
		case "info":
			cfg.ColorScheme.Info = value
		case "highlight":
			cfg.ColorScheme.Highlight = value
		case "background":
			cfg.ColorScheme.Background = value
		case "text":
			cfg.ColorScheme.Text = value
		case "textDim":
			cfg.ColorScheme.TextDim = value
		case "borderFocus":
			cfg.ColorScheme.BorderFocus = value
		}
	case "timeouts":
		switch field {
		case "networkRequest":
			if d, err := time.ParseDuration(value); err == nil {
				cfg.Timeouts.NetworkRequest = d
			}
		case "dotnetCli":
			if d, err := time.ParseDuration(value); err == nil {
				cfg.Timeouts.DotnetCLI = d
			}
		case "fileOperation":
			if d, err := time.ParseDuration(value); err == nil {
				cfg.Timeouts.FileOperation = d
			}
		}
	case "logRotation":
		switch field {
		case "maxSize":
			if i, err := strconv.Atoi(value); err == nil {
				cfg.LogRotation.MaxSize = i
			}
		case "maxAge":
			if i, err := strconv.Atoi(value); err == nil {
				cfg.LogRotation.MaxAge = i
			}
		case "maxBackups":
			if i, err := strconv.Atoi(value); err == nil {
				cfg.LogRotation.MaxBackups = i
			}
		case "compress":
			if b, err := parseBool(value); err == nil {
				cfg.LogRotation.Compress = b
			}
		}
	}

	return nil
}

// applyDoubleNestedSetting sets a double-nested config field (future expansion)
func applyDoubleNestedSetting(_ *Config, _, _, _, _ string) error {
	// Currently no triple-nested settings in our config
	// This is here for future extensibility
	return nil
}

// parseBool converts a string to a boolean value.
// Per FR-052: Supports "true", "false", "1", "0", "yes", "no" (case-insensitive)
func parseBool(value string) (bool, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return strconv.ParseBool(value)
	}
}
