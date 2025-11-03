# Quickstart Guide: Configuration Management System

**Feature**: 002-config-management
**Audience**: Developers implementing or using the configuration system
**Date**: 2025-11-02

## Overview

This guide provides practical examples for implementing and using the LazyNuGet configuration system. It covers common scenarios from basic usage to advanced features like encryption and hot-reload.

## Table of Contents

1. [Basic Usage](#basic-usage)
2. [Loading Configuration](#loading-configuration)
3. [Creating Config Files](#creating-config-files)
4. [Environment Variable Overrides](#environment-variable-overrides)
5. [CLI Flag Overrides](#cli-flag-overrides)
6. [Encrypting Sensitive Values](#encrypting-sensitive-values)
7. [Hot-Reload](#hot-reload)
8. [Validation](#validation)
9. [Testing](#testing)
10. [Troubleshooting](#troubleshooting)

---

## Basic Usage

### Using Defaults (Zero Configuration)

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/willibrandon/lazynuget/internal/config"
)

func main() {
	ctx := context.Background()
	loader := config.NewLoader()

	// Load with all defaults (no config file, no env vars, no CLI flags)
	cfg, err := loader.Load(ctx, config.LoadOptions{})
	if err != nil {
		// Only blocking errors reach here (syntax errors, file too large, etc.)
		fmt.Printf("Fatal config error: %v\n", err)
		os.Exit(1)
	}

	// Config is ready to use
	fmt.Printf("Theme: %s\n", cfg.Theme)                   // "default"
	fmt.Printf("Log Level: %s\n", cfg.LogLevel)            // "info"
	fmt.Printf("Max Concurrent Ops: %d\n", cfg.MaxConcurrentOps) // 4
}
```

---

## Loading Configuration

### Full Load with All Sources

```go
package main

import (
	"context"
	"github.com/willibrandon/lazynuget/internal/config"
	"github.com/willibrandon/lazynuget/internal/logging"
)

func main() {
	ctx := context.Background()
	logger := logging.New("info", "") // Create logger (empty path = stdout only)

	loader := config.NewLoader()

	cfg, err := loader.Load(ctx, config.LoadOptions{
		ConfigFilePath: "", // Empty = use platform default location
		CLIFlags: config.CLIFlags{
			LogLevel:       "debug",   // Override log level via CLI flag
			NonInteractive: false,
			NoColor:        false,
		},
		EnvVarPrefix: "LAZYNUGET_", // Default prefix
		StrictMode:   false,        // false = semantic errors log warnings
		Logger:       logger,
	})

	if err != nil {
		logger.Error("Config load failed", "error", err)
		os.Exit(1)
	}

	// Configuration is merged: defaults < file < env vars < CLI flags
	logger.Info("Config loaded",
		"from", cfg.LoadedFrom,
		"theme", cfg.Theme,
		"log_level", cfg.LogLevel)
}
```

### Custom Config File Location

```go
cfg, err := loader.Load(ctx, config.LoadOptions{
	ConfigFilePath: "/custom/path/to/config.yml", // Explicit path
	Logger:         logger,
})
```

---

## Creating Config Files

### Example: YAML Config (`~/.config/lazynuget/config.yml` on Linux)

```yaml
version: "1.0"

# UI Settings
theme: dark
colorScheme:
  border: "#333333"
  borderFocus: "#00FF00"
  text: "#FFFFFF"
  textDim: "#808080"
  background: "#1E1E1E"
  highlight: "#FFFF00"
  error: "#FF0000"
  warning: "#FFA500"
  success: "#00FF00"
  info: "#00FFFF"

compactMode: false
showHints: true
showLineNumbers: true
dateFormat: "2006-01-02 15:04:05"

# Keybindings
keybindingProfile: vim
keybindings:
  quit:
    action: "quit"
    key: "q"
    description: "Quit application"
    context: "global"
  refresh:
    action: "refresh"
    key: "Ctrl+R"
    description: "Refresh package list"
    context: "package-list"
  search:
    action: "search"
    key: "/"
    description: "Search packages"
    context: "package-list"

# Performance
maxConcurrentOps: 8
cacheSize: 100
refreshInterval: 5m
timeouts:
  networkRequest: 30s
  dotnetCLI: 60s
  fileOperation: 5s

# Dotnet CLI
dotnetPath: "" # Empty = auto-detect from PATH
dotnetVerbosity: normal

# Logging
logLevel: debug
logDir: "" # Empty = platform default
logFormat: json
logRotation:
  maxSize: 10
  maxAge: 30
  maxBackups: 5
  compress: true

# Hot-Reload
hotReload: true
```

### Example: TOML Config (`~/Library/Application Support/lazynuget/config.toml` on macOS)

```toml
version = "1.0"

# UI Settings
theme = "dark"
compactMode = false
showHints = true
showLineNumbers = true
dateFormat = "2006-01-02 15:04:05"

[colorScheme]
border = "#333333"
borderFocus = "#00FF00"
text = "#FFFFFF"
textDim = "#808080"
background = "#1E1E1E"
highlight = "#FFFF00"
error = "#FF0000"
warning = "#FFA500"
success = "#00FF00"
info = "#00FFFF"

# Keybindings
keybindingProfile = "vim"

[keybindings.quit]
action = "quit"
key = "q"
description = "Quit application"
context = "global"

[keybindings.refresh]
action = "refresh"
key = "Ctrl+R"
description = "Refresh package list"
context = "package-list"

# Performance
maxConcurrentOps = 8
cacheSize = 100
refreshInterval = "5m"

[timeouts]
networkRequest = "30s"
dotnetCLI = "60s"
fileOperation = "5s"

# Dotnet CLI
dotnetPath = ""
dotnetVerbosity = "normal"

# Logging
logLevel = "debug"
logDir = ""
logFormat = "json"

[logRotation]
maxSize = 10
maxAge = 30
maxBackups = 5
compress = true

# Hot-Reload
hotReload = true
```

---

## Environment Variable Overrides

### Setting Environment Variables

```bash
# Override log level
export LAZYNUGET_LOG_LEVEL=debug

# Override nested config (use underscores for nesting)
export LAZYNUGET_COLOR_SCHEME_BORDER="#FF0000"

# Override dotnet CLI path
export LAZYNUGET_DOTNET_PATH="/custom/path/dotnet"

# Override max concurrent operations
export LAZYNUGET_MAX_CONCURRENT_OPS=16

# Specify custom config file location
export LAZYNUGET_CONFIG="/custom/config.yml"

# Run application (env vars override config file)
./lazynuget
```

### In Go Code (Parsing Env Vars)

```go
// Environment variable parsing is automatic during config.Load()
// The loader reads all LAZYNUGET_* environment variables and merges them

cfg, err := loader.Load(ctx, config.LoadOptions{
	EnvVarPrefix: "LAZYNUGET_", // Default
	Logger:       logger,
})

// cfg.LogLevel will be "debug" if LAZYNUGET_LOG_LEVEL=debug was set
// cfg.ColorScheme.Border will be "#FF0000" if LAZYNUGET_COLOR_SCHEME_BORDER was set
```

---

## CLI Flag Overrides

### Command-Line Flags

```bash
# Override log level temporarily
./lazynuget --log-level=debug

# Use custom config file
./lazynuget --config=/custom/config.yml

# Disable colors (useful for piping output)
./lazynuget --no-color

# Print merged config for debugging
./lazynuget --print-config

# Validate config without starting application
./lazynuget --validate-config

# Combine multiple flags (CLI flags have highest precedence)
./lazynuget --config=/custom/config.yml --log-level=debug --no-color
```

### In Go Code (CLI Flag Integration)

```go
import "flag"

func main() {
	// Define CLI flags
	var (
		configPath = flag.String("config", "", "Config file path")
		logLevel   = flag.String("log-level", "", "Log level (debug, info, warn, error)")
		noColor    = flag.Bool("no-color", false, "Disable colored output")
		printConfig = flag.Bool("print-config", false, "Print merged config and exit")
		validateConfig = flag.Bool("validate-config", false, "Validate config and exit")
	)
	flag.Parse()

	ctx := context.Background()
	loader := config.NewLoader()

	cfg, err := loader.Load(ctx, config.LoadOptions{
		ConfigFilePath: *configPath,
		CLIFlags: config.CLIFlags{
			LogLevel: *logLevel,
			NoColor:  *noColor,
		},
		Logger: logger,
	})

	if err != nil {
		logger.Error("Config load failed", "error", err)
		os.Exit(1)
	}

	// Handle --print-config flag
	if *printConfig {
		fmt.Println(loader.PrintConfig(cfg))
		os.Exit(0)
	}

	// Handle --validate-config flag
	if *validateConfig {
		errors, err := loader.Validate(ctx, cfg)
		if err != nil {
			logger.Error("Validation failed", "error", err)
			os.Exit(1)
		}
		if len(errors) > 0 {
			for _, e := range errors {
				logger.Warn("Validation error", "key", e.Key, "constraint", e.Constraint)
			}
			os.Exit(1)
		}
		fmt.Println("Config is valid")
		os.Exit(0)
	}

	// Continue with application startup
	// ...
}
```

---

## Encrypting Sensitive Values

### Step 1: Generate and Store Encryption Key

```bash
# Option 1: Use lazynuget CLI to generate key and store in keychain
./lazynuget encrypt-value --generate-key --key-id=prod

# Output:
# Generated 256-bit encryption key for ID 'prod'
# Stored in platform keychain (macOS Keychain / Windows Credential Manager / Linux Secret Service)
# Key ID: prod

# Option 2: Provide your own key via environment variable (less secure, for CI/headless)
export LAZYNUGET_ENCRYPTION_KEY_PROD="hex-encoded-256-bit-key"
```

### Step 2: Encrypt a Sensitive Value

```bash
# Encrypt an API key using the stored key
./lazynuget encrypt-value --key-id=prod "my-secret-api-key-12345"

# Output:
# !encrypted AES256GCM:prod:AbCdEf1234567890+nonce_and_ciphertext_base64==
```

### Step 3: Use Encrypted Value in Config File

```yaml
# config.yml
nugetApiKey: !encrypted AES256GCM:prod:AbCdEf1234567890+nonce_and_ciphertext_base64==
dotnetPath: "/usr/local/share/dotnet/dotnet" # Regular unencrypted setting
```

### Step 4: Load Config (Decryption Happens Automatically)

```go
cfg, err := loader.Load(ctx, config.LoadOptions{Logger: logger})

// cfg.NugetApiKey will contain the decrypted value "my-secret-api-key-12345"
// Decryption happens automatically during Load() using key from keychain

// If decryption fails (key not found, wrong key, corrupted ciphertext):
//   - Warning logged
//   - Default value used for setting
//   - Application continues (non-blocking error)
```

### Encryption Implementation Example

```go
package main

import (
	"context"
	"fmt"
	"github.com/willibrandon/lazynuget/internal/config"
)

func encryptValue(plaintext string, keyID string) {
	ctx := context.Background()

	// Create encryptor with keychain and key derivation
	keychain := config.NewKeychainManager()
	keyDerivation := config.NewKeyDerivation()
	encryptor := config.NewEncryptor(keychain, keyDerivation)

	// Convert to string format for config file
	encryptedStr, err := encryptor.EncryptToString(ctx, plaintext, keyID)
	if err != nil {
		fmt.Printf("Encryption failed: %v\n", err)
		return
	}

	fmt.Printf("Encrypted value:\n%s\n", encryptedStr)
	// Output: !encrypted AES256GCM:prod:base64-ciphertext
}
```

---

## Hot-Reload

### Enabling Hot-Reload in Config

```yaml
# config.yml
hotReload: true
```

Or via environment variable:
```bash
export LAZYNUGET_HOT_RELOAD=true
./lazynuget
```

### Setting Up Hot-Reload in Application

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
	"github.com/willibrandon/lazynuget/internal/logging"
)

func main() {
	ctx := context.Background()
	loader := config.NewLoader()
	logger := logging.New("info", "")

	// Load initial config
	cfg, err := loader.Load(ctx, config.LoadOptions{Logger: logger})
	if err != nil {
		logger.Error("Config load failed: %v", err)
		os.Exit(1)
	}

	// Check if hot-reload is enabled
	if !cfg.HotReload {
		logger.Info("Hot-reload disabled, config is static")
		// Continue with static config
		return
	}

	// Create file watcher for hot-reload with callbacks
	watcher, err := config.NewConfigWatcher(config.WatchOptions{
		ConfigFilePath: cfg.LoadedFrom,
		LoadOptions:    config.LoadOptions{Logger: logger},
		DebounceDelay:  100 * time.Millisecond, // Wait 100ms after last change

		OnReload: func(newConfig *config.Config) {
			logger.Info("Config reloaded successfully")
			// Apply new config to application
			// NOTE: Some settings may require restart (flagged as non-hot-reloadable)
		},

		OnError: func(err error) {
			logger.Warn("Config reload failed: %v", err)
			// Keep previous config when reload fails
		},

		OnFileDeleted: func() {
			logger.Warn("Config file deleted, falling back to defaults")
			// Fall back to defaults
		},
	}, loader)

	if err != nil {
		logger.Error("Failed to create config watcher: %v", err)
		// Continue without hot-reload (not fatal)
		return
	}

	// Start watching for changes
	eventCh, errCh, err := watcher.Watch(ctx)
	if err != nil {
		logger.Error("Failed to start watcher: %v", err)
		return
	}

	// Process events in background goroutine
	go func() {
		for {
			select {
			case event, ok := <-eventCh:
				if !ok {
					return
				}
				// Event has NewConfig, Type, Timestamp, Error fields
				logger.Info("Config change detected: %s", event.Type)

			case err, ok := <-errCh:
				if !ok {
					return
				}
				logger.Error("Watcher error: %v", err)
			}
		}
	}()

	// Application runs with hot-reload enabled
	// Watcher goroutine monitors config file in background
	// ...

	// Clean up on shutdown
	defer watcher.Stop()
}
```

### Hot-Reload Behavior

- **Debouncing**: Waits 100ms after last file change before reloading (handles editors saving temp files)
- **Validation**: New config is validated before applying. If invalid, previous config is retained
- **Notification**: User is notified of successful reload, validation errors, or file deletion
- **Non-reloadable settings**: Settings like window size, keybinding profile may require restart. Watcher detects these and notifies user

---

## Validation

### Validating Config Without Loading

```go
// Useful for testing or CI validation
cfg := &config.Config{
	Theme: "invalid-theme", // Will fail validation
	MaxConcurrentOps: 999,  // Out of range [1, 16]
	LogLevel: "debug",      // Valid
}

errors, err := loader.Validate(ctx, cfg)
if err != nil {
	fmt.Printf("System error: %v\n", err)
	return
}

for _, e := range errors {
	fmt.Printf("[%s] %s: %s (suggested: %s)\n",
		e.Severity, e.Key, e.Constraint, e.SuggestedFix)
}

// Output:
// [warning] theme: must be one of: default, dark, light, solarized (suggested: use 'default')
// [warning] maxConcurrentOps: must be between 1 and 16 (suggested: use value between 1 and 16)
```

### Using --validate-config CLI Flag

```bash
# Validate config file without starting application
./lazynuget --validate-config

# Output (if errors exist):
# [WARNING] maxConcurrentOps: value 999 out of range [1, 16], using default 4
# [WARNING] colorScheme.border: invalid hex color '#GGGGGG', using default #FFFFFF
# Config has 2 warnings but is loadable (will use defaults for invalid settings)

# Exit code: 0 if loadable, 1 if blocking errors
```

---

## Testing

### Unit Test: Load Config with Defaults

```go
func TestLoadConfigDefaults(t *testing.T) {
	ctx := context.Background()
	loader := config.NewLoader()

	cfg, err := loader.Load(ctx, config.LoadOptions{})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify defaults
	if cfg.Theme != "default" {
		t.Errorf("Expected theme 'default', got '%s'", cfg.Theme)
	}
	if cfg.MaxConcurrentOps != 4 {
		t.Errorf("Expected maxConcurrentOps 4, got %d", cfg.MaxConcurrentOps)
	}
}
```

### Integration Test: Load from File

```go
func TestLoadConfigFromFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	configContent := `
version: "1.0"
theme: dark
maxConcurrentOps: 8
logLevel: debug
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Load config from file
	ctx := context.Background()
	loader := config.NewLoader()

	cfg, err := loader.Load(ctx, config.LoadOptions{
		ConfigFilePath: configPath,
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify file values override defaults
	if cfg.Theme != "dark" {
		t.Errorf("Expected theme 'dark', got '%s'", cfg.Theme)
	}
	if cfg.MaxConcurrentOps != 8 {
		t.Errorf("Expected maxConcurrentOps 8, got %d", cfg.MaxConcurrentOps)
	}
}
```

### Integration Test: Precedence Order

```go
func TestConfigPrecedence(t *testing.T) {
	// Setup: File says logLevel=info, env says logLevel=warn, CLI says logLevel=debug
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	os.WriteFile(configPath, []byte("logLevel: info\n"), 0644)

	os.Setenv("LAZYNUGET_LOG_LEVEL", "warn")
	defer os.Unsetenv("LAZYNUGET_LOG_LEVEL")

	ctx := context.Background()
	loader := config.NewLoader()

	cfg, err := loader.Load(ctx, config.LoadOptions{
		ConfigFilePath: configPath,
		CLIFlags: config.CLIFlags{
			LogLevel: "debug", // Highest precedence
		},
	})
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// CLI flag should win (highest precedence)
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected logLevel 'debug' (from CLI), got '%s'", cfg.LogLevel)
	}
}
```

---

## Troubleshooting

### Problem: Config file not found

**Symptom**: Application uses defaults, no error message

**Solution**:
- Check platform-specific default location:
  - macOS: `~/Library/Application Support/lazynuget/config.yml`
  - Linux: `~/.config/lazynuget/config.yml`
  - Windows: `%APPDATA%\lazynuget\config.yml`
- Create config file in default location or use `--config` flag to specify custom path
- Run `./lazynuget --print-config` to see which file is being loaded

### Problem: Config file has syntax error

**Symptom**: Application fails to start with error message

**Solution**:
- Check error message for line number: "Config file syntax error at line 42"
- Validate YAML syntax using online validator or `yamllint`
- Common issues: incorrect indentation, missing quotes, invalid YAML tags

### Problem: Setting not taking effect

**Symptom**: Changed setting in config file but application still uses old value

**Solution**:
- Check precedence: env var or CLI flag may be overriding file value
- Run `./lazynuget --print-config` to see final merged config and provenance
- If hot-reload is enabled: check that setting is hot-reloadable (some settings require restart)
- If hot-reload is disabled: restart application to pick up changes

### Problem: Both YAML and TOML files exist

**Symptom**: Application fails to start with error "Both config.yml and config.toml found"

**Solution**:
- Remove one of the config files (keep YAML or TOML, not both)
- Recommended: Use YAML for better comment support and readability

### Problem: Encrypted value decryption fails

**Symptom**: Warning logged "Failed to decrypt value for key X, using default"

**Solution**:
- Verify encryption key exists in keychain: `./lazynuget encrypt-value --list-keys`
- Check key ID matches: encrypted value format is `AES256GCM:<key-id>:...`
- For CI/headless environments: Set `LAZYNUGET_ENCRYPTION_KEY_<KEYID>` environment variable
- Re-encrypt the value if key was rotated: `./lazynuget encrypt-value --key-id=prod "new-value"`

### Problem: Validation warnings for unknown keys

**Symptom**: Warnings like "Unknown config key 'myCustomSetting', ignoring"

**Solution**:
- Check for typos in config file (e.g., `maxConcurentOps` instead of `maxConcurrentOps`)
- Remove deprecated/unused keys from config file
- These warnings are non-blocking; unknown keys are simply ignored

### Problem: Hot-reload not working

**Symptom**: Changed config file but application didn't reload

**Solution**:
- Verify `hotReload: true` in config file or `LAZYNUGET_HOT_RELOAD=true` env var
- Check logs for watcher errors (permissions, file system issues)
- Wait 100ms+ after saving file (debounce period)
- For network file systems (NFS, SMB): Hot-reload may not work reliably (known limitation)

---

## Additional Resources

- [data-model.md](./data-model.md) - Complete entity definitions and validation rules
- [research.md](./research.md) - Technical decisions and alternatives considered
- [contracts/](./contracts/) - Interface definitions for all config components
- [Constitution](../../.specify/memory/constitution.md) - Project principles guiding design

---

**Last Updated**: 2025-11-02
