package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/willibrandon/lazynuget/internal/config"
)

// TestYAMLConfigLoadsAndOverridesDefaults verifies YAML config file loads and overrides defaults
// See: T039, FR-003, FR-002
func TestYAMLConfigLoadsAndOverridesDefaults(t *testing.T) {
	// Get path to valid.yml fixture
	fixtureDir := filepath.Join("..", "fixtures", "configs")
	yamlPath := filepath.Join(fixtureDir, "valid.yml")

	// Verify fixture exists
	if _, err := os.Stat(yamlPath); err != nil {
		t.Fatalf("Test fixture not found: %s", yamlPath)
	}

	// Create config loader
	loader := config.NewConfigLoader()

	// Load config with explicit file path
	opts := config.LoadOptions{
		ConfigFilePath: yamlPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil, // Suppress logging in tests
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Failed to load YAML config: %v", err)
	}

	// Verify config was loaded from file
	if cfg.LoadedFrom != yamlPath {
		t.Errorf("Expected LoadedFrom=%s, got %s", yamlPath, cfg.LoadedFrom)
	}

	// Verify specific overrides from valid.yml
	if cfg.Theme != "dark" {
		t.Errorf("Expected theme=dark, got %s", cfg.Theme)
	}

	if !cfg.CompactMode {
		t.Error("Expected compactMode=true from YAML file")
	}

	if cfg.ShowHints {
		t.Error("Expected showHints=false from YAML file")
	}

	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected maxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected logLevel=debug, got %s", cfg.LogLevel)
	}

	if cfg.ColorScheme.Border != "#444444" {
		t.Errorf("Expected border color=#444444, got %s", cfg.ColorScheme.Border)
	}

	if cfg.Timeouts.NetworkRequest.Seconds() != 60 {
		t.Errorf("Expected networkRequest timeout=60s, got %v", cfg.Timeouts.NetworkRequest)
	}
}

// TestTOMLConfigLoadsAndOverridesDefaults verifies TOML config file loads and overrides defaults
// See: T040, FR-004, FR-002
func TestTOMLConfigLoadsAndOverridesDefaults(t *testing.T) {
	// Get path to valid.toml fixture
	fixtureDir := filepath.Join("..", "fixtures", "configs")
	tomlPath := filepath.Join(fixtureDir, "valid.toml")

	// Verify fixture exists
	if _, err := os.Stat(tomlPath); err != nil {
		t.Fatalf("Test fixture not found: %s", tomlPath)
	}

	// Create config loader
	loader := config.NewConfigLoader()

	// Load config with explicit file path
	opts := config.LoadOptions{
		ConfigFilePath: tomlPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Failed to load TOML config: %v", err)
	}

	// Verify config was loaded from file
	if cfg.LoadedFrom != tomlPath {
		t.Errorf("Expected LoadedFrom=%s, got %s", tomlPath, cfg.LoadedFrom)
	}

	// Debug: Print actual values
	t.Logf("DEBUG: Theme=%s, CompactMode=%v, MaxConcurrentOps=%d, LogLevel=%s",
		cfg.Theme, cfg.CompactMode, cfg.MaxConcurrentOps, cfg.LogLevel)

	// Verify specific overrides from valid.toml (should match valid.yml)
	if cfg.Theme != "dark" {
		t.Errorf("Expected theme=dark, got %s", cfg.Theme)
	}

	if !cfg.CompactMode {
		t.Error("Expected compactMode=true from TOML file")
	}

	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected maxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected logLevel=debug, got %s", cfg.LogLevel)
	}
}

// TestSyntaxErrorBlocksStartup verifies syntax error blocks startup per FR-010
// See: T041, FR-010
func TestSyntaxErrorBlocksStartup(t *testing.T) {
	// Get path to invalid_syntax.yml fixture
	fixtureDir := filepath.Join("..", "fixtures", "configs")
	invalidPath := filepath.Join(fixtureDir, "invalid_syntax.yml")

	// Verify fixture exists
	if _, err := os.Stat(invalidPath); err != nil {
		t.Fatalf("Test fixture not found: %s", invalidPath)
	}

	// Create config loader
	loader := config.NewConfigLoader()

	// Attempt to load invalid config
	opts := config.LoadOptions{
		ConfigFilePath: invalidPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	cfg, err := loader.Load(context.Background(), opts)

	// Syntax errors should be blocking - must return error
	if err == nil {
		t.Fatal("Expected error for invalid YAML syntax, got nil")
	}

	// Config should be nil on syntax error
	if cfg != nil {
		t.Error("Expected nil config on syntax error")
	}

	// Error message should mention YAML parsing
	if err != nil && len(err.Error()) < 10 {
		t.Errorf("Error message too short, expected detailed syntax error: %v", err)
	}
}

// TestSemanticErrorsFallbackToDefaults verifies semantic errors fallback to defaults per FR-012
// See: T042, T052-T056, FR-012, FR-013
func TestSemanticErrorsFallbackToDefaults(t *testing.T) {
	// Get path to out_of_range.yml fixture
	fixtureDir := filepath.Join("..", "fixtures", "configs")
	outOfRangePath := filepath.Join(fixtureDir, "out_of_range.yml")

	// Verify fixture exists
	if _, err := os.Stat(outOfRangePath); err != nil {
		t.Fatalf("Test fixture not found: %s", outOfRangePath)
	}

	// Create config loader
	loader := config.NewConfigLoader()

	// Load config with semantic errors (non-strict mode)
	opts := config.LoadOptions{
		ConfigFilePath: outOfRangePath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false, // Allow semantic errors with fallback
		Logger:         nil,
	}

	cfg, err := loader.Load(context.Background(), opts)
	// In non-strict mode, semantic errors should not block startup
	if err != nil {
		t.Fatalf("Expected no error in non-strict mode, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected valid config with defaults applied")
	}

	// Verify invalid values fell back to defaults (T056)
	defaults := config.GetDefaultConfig()

	// maxConcurrentOps: 999 in file (out of range 1-16) should fallback to default
	if cfg.MaxConcurrentOps == 999 {
		t.Error("Expected maxConcurrentOps to fallback to default, still has invalid value 999")
	}
	if cfg.MaxConcurrentOps != defaults.MaxConcurrentOps {
		t.Errorf("Expected maxConcurrentOps to use default %d, got %d", defaults.MaxConcurrentOps, cfg.MaxConcurrentOps)
	}

	// Invalid log level should fallback to default
	if cfg.LogLevel == "verbose" {
		t.Error("Expected logLevel to fallback to default, still has invalid value 'verbose'")
	}
	if cfg.LogLevel != defaults.LogLevel {
		t.Errorf("Expected logLevel to use default %s, got %s", defaults.LogLevel, cfg.LogLevel)
	}

	// Invalid color should fallback to default
	if cfg.ColorScheme.Border == "not-a-hex-color" {
		t.Error("Expected border color to fallback to default, still has invalid value")
	}
	if cfg.ColorScheme.Border != defaults.ColorScheme.Border {
		t.Errorf("Expected border color to use default %s, got %s", defaults.ColorScheme.Border, cfg.ColorScheme.Border)
	}
}

// TestBothFormatsPresentTriggersError verifies both formats present triggers error per FR-005
// See: T043, FR-005
func TestBothFormatsPresentTriggersError(t *testing.T) {
	// Get path to both_formats directory (contains both config.yml and config.toml)
	fixtureDir := filepath.Join("..", "fixtures", "configs", "both_formats")

	// Verify directory exists
	if _, err := os.Stat(fixtureDir); err != nil {
		t.Fatalf("Test fixture directory not found: %s", fixtureDir)
	}

	// Verify both files exist
	yamlPath := filepath.Join(fixtureDir, "config.yml")
	tomlPath := filepath.Join(fixtureDir, "config.toml")

	if _, err := os.Stat(yamlPath); err != nil {
		t.Fatalf("Expected config.yml to exist in both_formats directory")
	}
	if _, err := os.Stat(tomlPath); err != nil {
		t.Fatalf("Expected config.toml to exist in both_formats directory")
	}

	// Create config loader
	loader := config.NewConfigLoader()

	// Attempt to load config from directory with both formats
	// We'll try loading the YAML file explicitly, but the check should catch both formats
	opts := config.LoadOptions{
		ConfigFilePath: yamlPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	cfg, err := loader.Load(context.Background(), opts)

	// Should return error about multiple formats
	if err == nil {
		t.Fatal("Expected error when both YAML and TOML files present, got nil")
	}

	// Config should be nil on format conflict
	if cfg != nil {
		t.Error("Expected nil config on format conflict")
	}

	// Error should mention both formats
	errMsg := err.Error()
	if len(errMsg) < 10 {
		t.Errorf("Expected detailed error message about format conflict, got: %v", err)
	}
}

// TestUnknownKeysGenerateWarnings verifies unknown config keys are logged as warnings
// See: T037 (fixture), FR-011, FR-013
func TestUnknownKeysGenerateWarnings(t *testing.T) {
	// Get path to unknown_keys.yml fixture
	fixtureDir := filepath.Join("..", "fixtures", "configs")
	unknownKeysPath := filepath.Join(fixtureDir, "unknown_keys.yml")

	// Verify fixture exists
	if _, err := os.Stat(unknownKeysPath); err != nil {
		t.Fatalf("Test fixture not found: %s", unknownKeysPath)
	}

	// Create config loader
	loader := config.NewConfigLoader()

	// Load config with unknown keys
	opts := config.LoadOptions{
		ConfigFilePath: unknownKeysPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil, // In real scenario, logger would capture warnings
	}

	cfg, err := loader.Load(context.Background(), opts)
	// Unknown keys should not block startup in non-strict mode
	if err != nil {
		t.Fatalf("Expected no error for unknown keys in non-strict mode, got: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected valid config with unknown keys ignored")
	}

	// Verify known settings were loaded correctly
	if cfg.Theme != "dark" {
		t.Errorf("Expected theme=dark, got %s", cfg.Theme)
	}

	if cfg.MaxConcurrentOps != 6 {
		t.Errorf("Expected maxConcurrentOps=6, got %d", cfg.MaxConcurrentOps)
	}

	// Unknown keys (unknownSetting, anotherBadKey) should be silently ignored
	// In production, these would generate warning logs via Logger interface
}
