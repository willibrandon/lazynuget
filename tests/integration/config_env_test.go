package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
)

// T060: Test that simple environment variable overrides default
// See: FR-050, FR-002
func TestEnvVarOverridesDefault(t *testing.T) {
	// Set environment variable
	os.Setenv("LAZYNUGET_LOG_LEVEL", "debug")
	defer os.Unsetenv("LAZYNUGET_LOG_LEVEL")

	loader := config.NewLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		// No config file - just defaults + env vars
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
		Logger:       nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify environment variable overrode default
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected logLevel=debug from env var, got %s", cfg.LogLevel)
	}
}

// T061: Test that nested environment variable works (colorScheme.border)
// See: FR-051
func TestNestedEnvVarWorks(t *testing.T) {
	// Set nested environment variable using underscore notation
	os.Setenv("LAZYNUGET_COLOR_SCHEME_BORDER", "#FF0000")
	defer os.Unsetenv("LAZYNUGET_COLOR_SCHEME_BORDER")

	loader := config.NewLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
		Logger:       nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify nested environment variable applied
	if cfg.ColorScheme.Border != "#FF0000" {
		t.Errorf("Expected colorScheme.border=#FF0000 from env var, got %s", cfg.ColorScheme.Border)
	}
}

// T062: Test that environment variable overrides config file value
// See: FR-002 (precedence: CLI > Env > File > Default)
func TestEnvVarOverridesConfigFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := tmpDir + "/config.yml"

	configContent := `
logLevel: info
maxConcurrentOps: 6
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Set environment variable that should override file
	os.Setenv("LAZYNUGET_LOG_LEVEL", "debug")
	defer os.Unsetenv("LAZYNUGET_LOG_LEVEL")

	loader := config.NewLoader()

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

	// Verify env var overrode file value
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected logLevel=debug from env var (overriding file), got %s", cfg.LogLevel)
	}

	// Verify file value used when no env var present
	if cfg.MaxConcurrentOps != 6 {
		t.Errorf("Expected maxConcurrentOps=6 from file, got %d", cfg.MaxConcurrentOps)
	}
}

// T063: Test that invalid environment variable value triggers fallback to default
// See: FR-012
func TestInvalidEnvVarFallbackToDefault(t *testing.T) {
	// Set invalid environment variable value
	os.Setenv("LAZYNUGET_LOG_LEVEL", "invalid_level")
	defer os.Unsetenv("LAZYNUGET_LOG_LEVEL")

	loader := config.NewLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
		Logger:       nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify invalid value fell back to default
	defaults := config.GetDefaultConfig()
	if cfg.LogLevel != defaults.LogLevel {
		t.Errorf("Expected logLevel to fallback to default %s, got %s", defaults.LogLevel, cfg.LogLevel)
	}
}

// T064: Test type conversion for bool/int/duration environment variables
// See: FR-052
func TestEnvVarTypeConversion(t *testing.T) {
	// Set various typed environment variables
	os.Setenv("LAZYNUGET_COMPACT_MODE", "true")
	os.Setenv("LAZYNUGET_MAX_CONCURRENT_OPS", "8")
	os.Setenv("LAZYNUGET_REFRESH_INTERVAL", "10s")
	defer os.Unsetenv("LAZYNUGET_COMPACT_MODE")
	defer os.Unsetenv("LAZYNUGET_MAX_CONCURRENT_OPS")
	defer os.Unsetenv("LAZYNUGET_REFRESH_INTERVAL")

	loader := config.NewLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
		Logger:       nil,
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify bool conversion
	if !cfg.CompactMode {
		t.Error("Expected compactMode=true from env var (bool conversion)")
	}

	// Verify int conversion
	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected maxConcurrentOps=8 from env var (int conversion), got %d", cfg.MaxConcurrentOps)
	}

	// Verify duration conversion
	if cfg.RefreshInterval != 10*time.Second {
		t.Errorf("Expected refreshInterval=10s from env var (duration conversion), got %v", cfg.RefreshInterval)
	}
}
