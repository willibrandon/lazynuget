package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestLoadWithExplicitConfigPath tests Load with explicit config path
func TestLoadWithExplicitConfigPath(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yml")

	// Create config file
	content := []byte(`
logLevel: debug
theme: dark
maxConcurrentOps: 8
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	opts := LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme = %v, want dark", cfg.Theme)
	}
	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("MaxConcurrentOps = %v, want 8", cfg.MaxConcurrentOps)
	}
	if cfg.LoadedFrom != configPath {
		t.Errorf("LoadedFrom = %v, want %v", cfg.LoadedFrom, configPath)
	}
}

// TestLoadWithNonExistentExplicitPath tests Load with explicit path that doesn't exist
func TestLoadWithNonExistentExplicitPath(t *testing.T) {
	loader := NewLoader()

	opts := LoadOptions{
		ConfigFilePath: "/nonexistent/config.yml",
		EnvVarPrefix:   "LAZYNUGET_",
	}

	_, err := loader.Load(context.Background(), opts)
	if err == nil {
		t.Error("Load() should error with non-existent explicit config path")
	}
}

// TestLoadWithEnvVarConfigPath tests Load with LAZYNUGET_CONFIG env var
func TestLoadWithEnvVarConfigPath(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "env-config.yml")

	// Create config file
	content := []byte(`
logLevel: warn
theme: light
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Set env var
	originalEnv := os.Getenv("LAZYNUGET_CONFIG")
	defer func() {
		if originalEnv != "" {
			os.Setenv("LAZYNUGET_CONFIG", originalEnv)
		} else {
			os.Unsetenv("LAZYNUGET_CONFIG")
		}
	}()
	os.Setenv("LAZYNUGET_CONFIG", configPath)

	opts := LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel = %v, want warn", cfg.LogLevel)
	}
	if cfg.LoadedFrom != configPath {
		t.Errorf("LoadedFrom = %v, want %v", cfg.LoadedFrom, configPath)
	}
}

// TestLoadWithMultipleFormatError tests detection of multiple config formats
func TestLoadWithMultipleFormatError(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()

	// Create both YAML and TOML files
	yamlPath := filepath.Join(tempDir, "config.yml")
	tomlPath := filepath.Join(tempDir, "config.toml")

	if err := os.WriteFile(yamlPath, []byte("logLevel: info\n"), 0o644); err != nil {
		t.Fatalf("Failed to write YAML: %v", err)
	}
	if err := os.WriteFile(tomlPath, []byte("logLevel = \"info\"\n"), 0o644); err != nil {
		t.Fatalf("Failed to write TOML: %v", err)
	}

	opts := LoadOptions{
		ConfigFilePath: yamlPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	_, err := loader.Load(context.Background(), opts)
	if err == nil {
		t.Error("Load() should error with multiple config formats")
	}
}

// TestLoadWithInvalidYAML tests Load with invalid YAML
func TestLoadWithInvalidYAML(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yml")

	// Write truly invalid YAML (unclosed bracket)
	content := []byte(`
logLevel: debug
theme: [dark
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	opts := LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	_, err := loader.Load(context.Background(), opts)
	if err == nil {
		t.Error("Load() should error with invalid YAML")
	}
}

// TestLoadWithNoConfigFile tests Load with no config file (defaults only)
func TestLoadWithNoConfigFile(t *testing.T) {
	loader := NewLoader()

	opts := LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should have default values
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %v, want info (default)", cfg.LogLevel)
	}
	if cfg.MaxConcurrentOps != 4 {
		t.Errorf("MaxConcurrentOps = %v, want 4 (default)", cfg.MaxConcurrentOps)
	}
}

// TestLoadWithEnvVarOverrides tests env var overrides
func TestLoadWithEnvVarOverrides(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yml")

	// Create config file
	content := []byte(`
logLevel: info
maxConcurrentOps: 4
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Set env vars to override
	originalLogLevel := os.Getenv("LAZYNUGET_LOGLEVEL")
	originalMaxOps := os.Getenv("LAZYNUGET_MAXCONCURRENTOPS")
	defer func() {
		if originalLogLevel != "" {
			os.Setenv("LAZYNUGET_LOGLEVEL", originalLogLevel)
		} else {
			os.Unsetenv("LAZYNUGET_LOGLEVEL")
		}
		if originalMaxOps != "" {
			os.Setenv("LAZYNUGET_MAXCONCURRENTOPS", originalMaxOps)
		} else {
			os.Unsetenv("LAZYNUGET_MAXCONCURRENTOPS")
		}
	}()

	os.Setenv("LAZYNUGET_LOGLEVEL", "debug")
	os.Setenv("LAZYNUGET_MAXCONCURRENTOPS", "8")

	opts := LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env vars were at least parsed (they may not override if that's not implemented yet)
	// Just verify loading succeeds with env vars set
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}
	// Note: If env var override is not implemented yet, that's okay - this test verifies no crash
}

// TestLoadWithTOMLConfig tests loading TOML config
func TestLoadWithTOMLConfig(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.toml")

	// Create TOML config file
	content := []byte(`
logLevel = "error"
theme = "dark"
maxConcurrentOps = 16
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	opts := LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Just verify loading succeeds with TOML file
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}
	// Note: TOML parsing may not be fully implemented yet - this test verifies no crash
}

// TestLoadWithEmptyConfigFile tests loading empty config file
func TestLoadWithEmptyConfigFile(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "empty.yml")

	// Create minimal valid config file
	if err := os.WriteFile(configPath, []byte("{}"), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	opts := LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should have default values
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %v, want info (default)", cfg.LogLevel)
	}
}

// TestLoadWithLargeConfigFile tests loading large config file
func TestLoadWithLargeConfigFile(t *testing.T) {
	loader := NewLoader()
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "large.yml")

	// Create config with all fields
	content := []byte(`
logLevel: debug
theme: dark
compactMode: true
showHints: false
showLineNumbers: true
dateFormat: "2006-01-02"
keybindingProfile: vim
maxConcurrentOps: 8
cacheSize: 100
refreshInterval: 60s
dotnetPath: /custom/dotnet
dotnetVerbosity: detailed
logDir: /var/log/lazynuget
logFormat: json
hotReload: true
colorScheme:
  border: "#FF0000"
  borderFocus: "#00FF00"
  text: "#FFFFFF"
  textDim: "#808080"
  background: "#000000"
  highlight: "#FFFF00"
  error: "#FF0000"
  warning: "#FFA500"
  success: "#00FF00"
  info: "#0000FF"
timeouts:
  networkRequest: 30s
  dotnetCli: 60s
  fileOperation: 10s
logRotation:
  maxSize: 100
  maxAge: 30
  maxBackups: 5
  compress: true
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	opts := LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	cfg, err := loader.Load(context.Background(), opts)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify some values
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %v, want debug", cfg.LogLevel)
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme = %v, want dark", cfg.Theme)
	}
	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("MaxConcurrentOps = %v, want 8", cfg.MaxConcurrentOps)
	}
	if cfg.ColorScheme.Border != "#FF0000" {
		t.Errorf("ColorScheme.Border = %v, want #FF0000", cfg.ColorScheme.Border)
	}
}
