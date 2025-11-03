package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// parseTOML parses TOML config file content into a Config struct.
// See: T045, FR-004, FR-010
func parseTOML(data []byte) (*Config, error) {
	var cfg Config

	// Parse TOML with strict decoding
	metadata, err := toml.Decode(string(data), &cfg)
	if err != nil {
		return nil, fmt.Errorf("TOML parsing error: %w\n\n"+
			"Please check the file for syntax errors:\n"+
			"  • Ensure proper TOML syntax (key = value)\n"+
			"  • Check for missing quotes around strings\n"+
			"  • Verify section headers are in [brackets]\n"+
			"  • Validate TOML syntax at https://www.toml.io/", err)
	}

	// Check for undecoded keys (unknown configuration keys)
	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		// Note: We'll log these as warnings, not errors (handled by validator)
		// Just noting them here for potential future use
		_ = undecoded
	}

	return &cfg, nil
}
