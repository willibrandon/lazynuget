package config

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// parseYAML parses YAML config file content into a Config struct.
// Uses strict mode to catch unknown fields and syntax errors.
// See: T044, FR-003, FR-010
func parseYAML(data []byte) (*Config, error) {
	var cfg Config

	// Use decoder with strict mode to catch unknown fields
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&cfg); err != nil {
		// Provide helpful error message with line/column info if available
		return nil, fmt.Errorf("YAML parsing error: %w\n\n"+
			"Please check the file for syntax errors:\n"+
			"  • Ensure proper indentation (use spaces, not tabs)\n"+
			"  • Check for missing colons or quotes\n"+
			"  • Verify all keys are valid configuration settings\n"+
			"  • Validate YAML syntax at https://www.yamllint.com/", err)
	}

	return &cfg, nil
}
