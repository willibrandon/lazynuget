package config

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// parseYAML parses YAML config file content into a Config struct.
// Unknown fields are silently ignored per FR-011 and FR-013.
// See: T044, FR-003, FR-010, FR-011
func parseYAML(data []byte) (*Config, error) {
	var cfg Config

	// Use decoder WITHOUT strict mode - unknown fields should be ignored
	// Per FR-011: Unknown config keys are ignored with warning
	// Per FR-013: Unknown keys are non-blocking semantic errors
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(false) // Allow unknown fields

	if err := decoder.Decode(&cfg); err != nil {
		// Provide helpful error message with line/column info if available
		return nil, fmt.Errorf("YAML parsing error: %w\n\n"+
			"Please check the file for syntax errors:\n"+
			"  • Ensure proper indentation (use spaces, not tabs)\n"+
			"  • Check for missing colons or quotes\n"+
			"  • Validate YAML syntax at https://www.yamllint.com/", err)
	}

	return &cfg, nil
}
