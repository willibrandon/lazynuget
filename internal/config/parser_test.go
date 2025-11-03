package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTOMLParserBasic(t *testing.T) {
	tomlData := `
log_level = "debug"
max_concurrent_ops = 8
theme = "dark"
`

	cfg, err := parseTOML([]byte(tomlData))
	if err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
	}

	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected MaxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
	}

	if cfg.Theme != "dark" {
		t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
	}
}

func TestYAMLParserBasic(t *testing.T) {
	yamlData := `
logLevel: debug
maxConcurrentOps: 8
theme: dark
`

	cfg, err := parseYAML([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
	}

	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected MaxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
	}

	if cfg.Theme != "dark" {
		t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
	}
}

func TestParseActualTOMLFixture(t *testing.T) {
	// Read the actual valid.toml fixture used in integration tests
	fixturePath := filepath.Join("..", "..", "tests", "fixtures", "configs", "valid.toml")

	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixture: %v", err)
	}

	t.Logf("TOML file size: %d bytes", len(data))
	t.Logf("First 500 chars:\n%s", string(data[:min(500, len(data))]))

	cfg, err := parseTOML(data)
	if err != nil {
		t.Fatalf("Failed to parse TOML: %v", err)
	}

	t.Logf("Parsed config: LogLevel=%q, MaxConcurrentOps=%d, Theme=%q, CompactMode=%v",
		cfg.LogLevel, cfg.MaxConcurrentOps, cfg.Theme, cfg.CompactMode)

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel=debug, got %q", cfg.LogLevel)
	}

	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected MaxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
	}

	if cfg.Theme != "dark" {
		t.Errorf("Expected Theme=dark, got %q", cfg.Theme)
	}

	if !cfg.CompactMode {
		t.Error("Expected CompactMode=true")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
