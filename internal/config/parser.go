package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigFormat represents the format of a configuration file.
type ConfigFormat int

const (
	FormatUnknown ConfigFormat = iota
	FormatYAML
	FormatTOML
)

// MaxConfigFileSize is the maximum allowed size for config files (10 MB).
// See: FR-009
const MaxConfigFileSize = 10 * 1024 * 1024 // 10 MB in bytes

// detectFormat determines the config file format from its extension.
// See: T046, FR-003, FR-004
func detectFormat(filePath string) ConfigFormat {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".yml", ".yaml":
		return FormatYAML
	case ".toml":
		return FormatTOML
	default:
		return FormatUnknown
	}
}

// validateFileSize checks if a file is within the size limit.
// See: T047, FR-009
func validateFileSize(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	if info.Size() > MaxConfigFileSize {
		return fmt.Errorf("config file exceeds maximum size of 10 MB (actual: %d bytes)", info.Size())
	}

	return nil
}

// checkMultipleFormats checks if both YAML and TOML config files exist in a directory.
// See: T048, FR-005
func checkMultipleFormats(configDir string) error {
	// Check for common config file names
	yamlFiles := []string{
		filepath.Join(configDir, "config.yml"),
		filepath.Join(configDir, "config.yaml"),
	}
	tomlFiles := []string{
		filepath.Join(configDir, "config.toml"),
	}

	hasYAML := false
	for _, f := range yamlFiles {
		if _, err := os.Stat(f); err == nil {
			hasYAML = true
			break
		}
	}

	hasTOML := false
	for _, f := range tomlFiles {
		if _, err := os.Stat(f); err == nil {
			hasTOML = true
			break
		}
	}

	if hasYAML && hasTOML {
		return fmt.Errorf("both YAML and TOML config files found in %s - please use only one format", configDir)
	}

	return nil
}

// parseConfigFile loads and parses a config file, handling syntax errors.
// See: T049, FR-010
func parseConfigFile(filePath string) (*Config, error) {
	// Validate file size
	if err := validateFileSize(filePath); err != nil {
		return nil, err
	}

	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Detect format and parse
	format := detectFormat(filePath)
	switch format {
	case FormatYAML:
		return parseYAML(data)
	case FormatTOML:
		return parseTOML(data)
	default:
		return nil, fmt.Errorf("unsupported config file format (must be .yml, .yaml, or .toml): %s", filePath)
	}
}
