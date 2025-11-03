package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkConfigLoadStartup tests FR-042: Startup config load <500ms
func BenchmarkConfigLoadStartup(b *testing.B) {
	ctx := context.Background()
	loader := NewLoader()

	// Create a typical config file
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: "1.0"
theme: dark
logLevel: debug
maxConcurrentOps: 8
compactMode: false
showHints: true
showLineNumbers: true

colorScheme:
  border: "#333333"
  borderFocus: "#00FF00"
  text: "#FFFFFF"
  background: "#1E1E1E"

timeouts:
  networkRequest: 30s
  dotnetCLI: 60s
  fileOperation: 5s

logRotation:
  maxSize: 10
  maxAge: 30
  maxBackups: 5
  compress: true
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	b.ResetTimer()

	// Run benchmark
	for b.Loop() {
		_, err := loader.Load(ctx, LoadOptions{
			ConfigFilePath: configPath,
		})
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}

	b.StopTimer()

	// Verify performance: <500ms per FR-042
	avgTime := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
	b.Logf("Average config load time: %v", avgTime)

	if avgTime > 500*time.Millisecond {
		b.Errorf("Config load time %v exceeds 500ms target (FR-042)", avgTime)
	}
}

// BenchmarkConfigFileParsing tests SC-010: Config file parsing <100ms for typical files
func BenchmarkConfigFileParsing(b *testing.B) {
	// Create a typical config file
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	configContent := `
version: "1.0"
theme: dark
logLevel: debug
maxConcurrentOps: 8
compactMode: false
showHints: true

colorScheme:
  border: "#333333"
  borderFocus: "#00FF00"
  text: "#FFFFFF"

timeouts:
  networkRequest: 30s
  dotnetCLI: 60s
`
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	if err != nil {
		b.Fatalf("Failed to create test config: %v", err)
	}

	// Run benchmark
	for b.Loop() {
		_, err := parseConfigFile(configPath)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}

	b.StopTimer()

	// Verify performance: <100ms per SC-010
	avgTime := time.Duration(b.Elapsed().Nanoseconds() / int64(b.N))
	b.Logf("Average parse time: %v", avgTime)

	if avgTime > 100*time.Millisecond {
		b.Errorf("Config parsing time %v exceeds 100ms target (SC-010)", avgTime)
	}
}

// BenchmarkConfigLoadDefaults tests loading with defaults only
func BenchmarkConfigLoadDefaults(b *testing.B) {
	ctx := context.Background()
	loader := NewLoader()

	for b.Loop() {
		_, err := loader.Load(ctx, LoadOptions{})
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

// BenchmarkConfigValidation tests validation performance
func BenchmarkConfigValidation(b *testing.B) {
	ctx := context.Background()
	loader := NewLoader()

	cfg := &Config{
		Theme:            "dark",
		LogLevel:         "debug",
		MaxConcurrentOps: 8,
		CompactMode:      false,
		ShowHints:        true,
	}

	for b.Loop() {
		_, err := loader.Validate(ctx, cfg)
		if err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}

// BenchmarkEnvVarParsing tests environment variable parsing performance
func BenchmarkEnvVarParsing(b *testing.B) {
	// Set up environment variables
	os.Setenv("LAZYNUGET_LOG_LEVEL", "debug")
	os.Setenv("LAZYNUGET_THEME", "dark")
	os.Setenv("LAZYNUGET_MAX_CONCURRENT_OPS", "8")
	os.Setenv("LAZYNUGET_COLOR_SCHEME_BORDER", "#333333")
	defer func() {
		os.Unsetenv("LAZYNUGET_LOG_LEVEL")
		os.Unsetenv("LAZYNUGET_THEME")
		os.Unsetenv("LAZYNUGET_MAX_CONCURRENT_OPS")
		os.Unsetenv("LAZYNUGET_COLOR_SCHEME_BORDER")
	}()

	for b.Loop() {
		_ = parseEnvVars("LAZYNUGET_")
	}
}

// BenchmarkYAMLParsing tests YAML parsing performance
func BenchmarkYAMLParsing(b *testing.B) {
	yamlContent := []byte(`
version: "1.0"
theme: dark
logLevel: debug
maxConcurrentOps: 8

colorScheme:
  border: "#333333"
  borderFocus: "#00FF00"
  text: "#FFFFFF"

timeouts:
  networkRequest: 30s
  dotnetCLI: 60s
`)

	for b.Loop() {
		_, err := parseYAML(yamlContent)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}

// BenchmarkTOMLParsing tests TOML parsing performance
func BenchmarkTOMLParsing(b *testing.B) {
	tomlContent := []byte(`
version = "1.0"
theme = "dark"
log_level = "debug"
max_concurrent_ops = 8

[color_scheme]
border = "#333333"
border_focus = "#00FF00"
text = "#FFFFFF"

[timeouts]
network_request = "30s"
dotnet_cli = "60s"
`)

	for b.Loop() {
		_, err := parseTOML(tomlContent)
		if err != nil {
			b.Fatalf("Parse failed: %v", err)
		}
	}
}
