package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
)

// T073: Test that --log-level flag overrides all other sources
// See: FR-054, FR-002 (precedence: CLI > Env > File > Default)
func TestCLIFlagOverridesAllSources(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yml"

	configContent := `
logLevel: info
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variable
	os.Setenv("LAZYNUGET_LOG_LEVEL", "warn")
	defer os.Unsetenv("LAZYNUGET_LOG_LEVEL")

	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
		CLIFlags: config.CLIFlags{
			LogLevel: "debug", // CLI flag should win
		},
		StrictMode: false,
		Logger:     nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify CLI flag overrode env var and file
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected logLevel=debug from CLI flag (overriding env and file), got %s", cfg.LogLevel)
	}
}

// T074: Test that --config flag specifies custom config file path
// See: FR-053
func TestConfigFlagSpecifiesPath(t *testing.T) {
	// Create temporary config file with custom settings
	tmpDir := t.TempDir()
	customConfigPath := tmpDir + "/custom-config.yml"

	configContent := `
logLevel: debug
theme: dark
maxConcurrentOps: 8
`
	if err := os.WriteFile(customConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: customConfigPath, // Custom path from --config flag
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify custom config file was loaded
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected logLevel=debug from custom config file, got %s", cfg.LogLevel)
	}
	if cfg.Theme != "dark" {
		t.Errorf("Expected theme=dark from custom config file, got %s", cfg.Theme)
	}
	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected maxConcurrentOps=8 from custom config file, got %d", cfg.MaxConcurrentOps)
	}
}

// T075: Test that --non-interactive flag works
// See: FR-054
func TestNonInteractiveFlagWorks(t *testing.T) {
	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		CLIFlags: config.CLIFlags{
			NonInteractive: true,
		},
		StrictMode: false,
		Logger:     nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Note: NonInteractive flag is consumed by bootstrap, not stored in Config
	// This test verifies the flag can be passed through LoadOptions
	// The actual non-interactive behavior is tested in noninteractive_test.go
	if cfg == nil {
		t.Error("Expected valid config, got nil")
	}
}

// T076: Test that --no-color flag works
// See: FR-054
func TestNoColorFlagWorks(t *testing.T) {
	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		CLIFlags: config.CLIFlags{
			NoColor: true,
		},
		StrictMode: false,
		Logger:     nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Note: NoColor flag is consumed by bootstrap/GUI, not stored in Config
	// This test verifies the flag can be passed through LoadOptions
	// The actual no-color behavior is handled by the terminal/GUI layer
	if cfg == nil {
		t.Error("Expected valid config, got nil")
	}
}

// T077: Test that --print-config outputs merged config
// See: FR-055
func TestPrintConfigOutputsMergedConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yml"

	configContent := `
logLevel: debug
theme: dark
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Test PrintConfig() method
	output := loader.PrintConfig(cfg)

	// Verify output contains key configuration sections
	if output == "" {
		t.Error("PrintConfig() returned empty string")
	}

	// Check for expected content
	expectedStrings := []string{
		"LazyNuGet Configuration",
		"Loaded from:",
		"theme:",
		"logLevel:",
		"maxConcurrentOps:",
	}

	for _, expected := range expectedStrings {
		if !contains(output, expected) {
			t.Errorf("PrintConfig() output missing expected string: %s", expected)
		}
	}
}

// T078: Test that --validate-config validates without starting app
// See: FR-056
func TestValidateConfigWithoutStarting(t *testing.T) {
	// Create temporary config file with valid settings
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yml"

	configContent := `
logLevel: debug
theme: dark
maxConcurrentOps: 8
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	// Load config
	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Validate config
	validationErrors, err := loader.Validate(ctx, cfg)
	if err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}

	// Verify no validation errors for valid config
	// Note: We may have warnings (severity="warning") but no errors (severity="error")
	hasErrors := false
	for _, ve := range validationErrors {
		if ve.Severity == "error" {
			hasErrors = true
			t.Logf("Validation error: %s", ve.Error())
		}
	}
	if hasErrors {
		t.Error("Expected no validation errors for valid config")
	}
}

// Test that --validate-config catches invalid config
func TestValidateConfigCatchesErrors(t *testing.T) {
	// Create temporary config file with invalid settings
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yml"

	configContent := `
logLevel: invalid_level
maxConcurrentOps: 999
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	// Load config (should succeed with warnings)
	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Validate config (should report errors)
	validationErrors, err := loader.Validate(ctx, cfg)
	if err != nil {
		t.Fatalf("Validate() system error: %v", err)
	}

	// Verify validation caught the errors
	if len(validationErrors) == 0 {
		t.Error("Expected validation errors for invalid config, got none")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
