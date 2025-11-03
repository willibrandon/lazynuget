package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
)

// T023: Test that GetDefaultConfig returns a valid Config with all expected default values
func TestGetDefaultConfig(t *testing.T) {
	cfg := config.GetDefaultConfig()

	// Verify config is not nil
	if cfg == nil {
		t.Fatal("GetDefaultConfig() returned nil")
	}

	// Verify meta fields
	if cfg.Version != "1.0" {
		t.Errorf("Expected Version '1.0', got '%s'", cfg.Version)
	}
	if cfg.LoadedFrom != "defaults" {
		t.Errorf("Expected LoadedFrom 'defaults', got '%s'", cfg.LoadedFrom)
	}

	// Verify UI settings
	if cfg.Theme != "default" {
		t.Errorf("Expected Theme 'default', got '%s'", cfg.Theme)
	}
	if cfg.CompactMode != false {
		t.Errorf("Expected CompactMode false, got %v", cfg.CompactMode)
	}
	if cfg.ShowHints != true {
		t.Errorf("Expected ShowHints true, got %v", cfg.ShowHints)
	}
	if cfg.ShowLineNumbers != false {
		t.Errorf("Expected ShowLineNumbers false, got %v", cfg.ShowLineNumbers)
	}
	if cfg.DateFormat != "2006-01-02" {
		t.Errorf("Expected DateFormat '2006-01-02', got '%s'", cfg.DateFormat)
	}

	// Verify color scheme defaults
	if cfg.ColorScheme.Border != "#FFFFFF" {
		t.Errorf("Expected Border '#FFFFFF', got '%s'", cfg.ColorScheme.Border)
	}
	if cfg.ColorScheme.Error != "#FF0000" {
		t.Errorf("Expected Error '#FF0000', got '%s'", cfg.ColorScheme.Error)
	}

	// Verify keybindings
	if cfg.KeybindingProfile != "default" {
		t.Errorf("Expected KeybindingProfile 'default', got '%s'", cfg.KeybindingProfile)
	}
	if cfg.Keybindings == nil {
		t.Error("Expected Keybindings map to be initialized, got nil")
	}

	// Verify performance settings
	if cfg.MaxConcurrentOps != 4 {
		t.Errorf("Expected MaxConcurrentOps 4, got %d", cfg.MaxConcurrentOps)
	}
	if cfg.CacheSize != 50 {
		t.Errorf("Expected CacheSize 50, got %d", cfg.CacheSize)
	}
	if cfg.RefreshInterval != 0 {
		t.Errorf("Expected RefreshInterval 0, got %v", cfg.RefreshInterval)
	}

	// Verify timeouts
	if cfg.Timeouts.NetworkRequest != 30*time.Second {
		t.Errorf("Expected NetworkRequest timeout 30s, got %v", cfg.Timeouts.NetworkRequest)
	}
	if cfg.Timeouts.DotnetCLI != 60*time.Second {
		t.Errorf("Expected DotnetCLI timeout 60s, got %v", cfg.Timeouts.DotnetCLI)
	}
	if cfg.Timeouts.FileOperation != 5*time.Second {
		t.Errorf("Expected FileOperation timeout 5s, got %v", cfg.Timeouts.FileOperation)
	}

	// Verify dotnet settings
	if cfg.DotnetPath != "" {
		t.Errorf("Expected DotnetPath empty (auto-detect), got '%s'", cfg.DotnetPath)
	}
	if cfg.DotnetVerbosity != "minimal" {
		t.Errorf("Expected DotnetVerbosity 'minimal', got '%s'", cfg.DotnetVerbosity)
	}

	// Verify logging settings
	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel 'info', got '%s'", cfg.LogLevel)
	}
	if cfg.LogDir != "" {
		t.Errorf("Expected LogDir empty (platform default), got '%s'", cfg.LogDir)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("Expected LogFormat 'text', got '%s'", cfg.LogFormat)
	}

	// Verify log rotation
	if cfg.LogRotation.MaxSize != 10 {
		t.Errorf("Expected LogRotation.MaxSize 10, got %d", cfg.LogRotation.MaxSize)
	}
	if cfg.LogRotation.MaxAge != 30 {
		t.Errorf("Expected LogRotation.MaxAge 30, got %d", cfg.LogRotation.MaxAge)
	}
	if cfg.LogRotation.MaxBackups != 3 {
		t.Errorf("Expected LogRotation.MaxBackups 3, got %d", cfg.LogRotation.MaxBackups)
	}
	if cfg.LogRotation.Compress != true {
		t.Errorf("Expected LogRotation.Compress true, got %v", cfg.LogRotation.Compress)
	}

	// Verify hot-reload
	if cfg.HotReload != false {
		t.Errorf("Expected HotReload false (disabled by default), got %v", cfg.HotReload)
	}
}

// T024: Test that when config file doesn't exist at default location, ConfigLoader uses defaults
func TestMissingConfigFileUsesDefaults(t *testing.T) {
	// Per FR-002: Missing config at default location should silently use defaults
	// Do NOT provide explicit ConfigFilePath - let it check default location
	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set LAZYNUGET_CONFIG to a nonexistent path to simulate "no default config found"
	// This simulates the case where the default platform config path doesn't have a file
	os.Setenv("LAZYNUGET_CONFIG", filepath.Join(t.TempDir(), "nonexistent.yml"))
	defer os.Unsetenv("LAZYNUGET_CONFIG")

	opts := config.LoadOptions{
		// NO ConfigFilePath set - let it discover from env/default
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
		Logger:       nil,
	}

	// Load config - should not return error for missing file, just use defaults
	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() returned error for missing config file at default location: %v", err)
	}

	// Verify we got defaults
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.Theme != "default" {
		t.Errorf("Expected default Theme 'default', got '%s'", cfg.Theme)
	}
	if cfg.MaxConcurrentOps != 4 {
		t.Errorf("Expected default MaxConcurrentOps 4, got %d", cfg.MaxConcurrentOps)
	}
}

// T025: Test that empty config file is treated as a syntax error
func TestEmptyConfigFileUsesDefaults(t *testing.T) {
	// Per Phase 4 implementation: empty YAML file (EOF) is a syntax error
	// This changed from Phase 3 where we expected it to fall back to defaults
	emptyConfigPath := filepath.Join("..", "fixtures", "configs", "empty.yml")

	// Verify fixture exists
	if _, err := os.Stat(emptyConfigPath); os.IsNotExist(err) {
		t.Skipf("Skipping test: fixture file %s does not exist yet", emptyConfigPath)
	}

	// Create ConfigLoader
	loader := config.NewConfigLoader()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: emptyConfigPath,
		EnvVarPrefix:   "LAZYNUGET_",
		StrictMode:     false,
		Logger:         nil,
	}

	// Load config - should return error for empty file (EOF is syntax error)
	cfg, err := loader.Load(ctx, opts)
	if err == nil {
		t.Fatal("Expected error for empty config file, got nil")
	}

	// Config should be nil on syntax error
	if cfg != nil {
		t.Error("Expected nil config on empty file error")
	}

	// Error should mention YAML parsing
	if !strings.Contains(err.Error(), "YAML parsing error") && !strings.Contains(err.Error(), "EOF") {
		t.Errorf("Expected error to mention YAML parsing or EOF, got: %v", err)
	}
}
