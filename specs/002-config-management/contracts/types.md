# Configuration Types

Type definitions for all configuration entities.

## Root Configuration Type

```go
package contracts

import (
	"time"
)

// Config is the root configuration object containing all application settings.
// See: data-model.md for complete field documentation and validation rules.
type Config struct {
	// Meta
	Version    string
	LoadedFrom string
	LoadedAt   time.Time

	// UI Settings (FR-020 through FR-025)
	Theme           string
	ColorScheme     ColorScheme
	CompactMode     bool
	ShowHints       bool
	ShowLineNumbers bool
	DateFormat      string

	// Keybindings (FR-026 through FR-030)
	Keybindings       map[string]KeyBinding
	KeybindingProfile string

	// Performance (FR-031 through FR-034)
	MaxConcurrentOps int
	CacheSize        int
	RefreshInterval  time.Duration
	Timeouts         Timeouts

	// Dotnet CLI Integration (FR-035 through FR-038)
	DotnetPath      string
	DotnetVerbosity string

	// Logging (FR-039 through FR-042)
	LogLevel    string
	LogDir      string
	LogFormat   string
	LogRotation LogRotation

	// Hot-Reload (FR-043)
	HotReload bool
}
```

## UI Types

```go
// ColorScheme defines customizable colors for UI elements.
// See: FR-021
type ColorScheme struct {
	Border      string
	BorderFocus string
	Text        string
	TextDim     string
	Background  string
	Highlight   string
	Error       string
	Warning     string
	Success     string
	Info        string
}

// KeyBinding maps an action to a key combination.
// See: FR-026 through FR-030
type KeyBinding struct {
	Action      string
	Key         string
	Description string
	Context     string
}
```

## Performance Types

```go
// Timeouts defines timeout durations for different operation types.
// See: FR-034
type Timeouts struct {
	NetworkRequest time.Duration
	DotnetCLI      time.Duration
	FileOperation  time.Duration
}
```

## Logging Types

```go
// LogRotation configures log file rotation.
// See: FR-042
type LogRotation struct {
	MaxSize    int  // Megabytes
	MaxAge     int  // Days
	MaxBackups int  // Number of old logs to keep
	Compress   bool // Compress rotated logs
}
```

## Validation Types

```go
// ValidationError describes a configuration validation failure.
// See: data-model.md entity #8
type ValidationError struct {
	Key          string      // Config key path (e.g., "maxConcurrentOps")
	Value        interface{} // Invalid value provided
	Constraint   string      // Constraint that failed
	SuggestedFix string      // Suggested correction
	Severity     string      // "error" (blocks startup) or "warning" (falls back to default)
	DefaultUsed  interface{} // Default value used (if severity is "warning")
}

// Error implements the error interface for ValidationError.
func (ve ValidationError) Error() string {
	if ve.Severity == "error" {
		return ve.Key + ": " + ve.Constraint
	}
	return ve.Key + ": " + ve.Constraint + " (using default: " + ve.DefaultUsed.(string) + ")"
}
```

## Source Tracking Types

```go
// ConfigSource represents one of the four configuration sources.
// See: data-model.md entity #6
type ConfigSource struct {
	Name       string
	Precedence int
	Data       map[string]interface{}
	FilePath   string
	LoadedAt   time.Time
}

// MergedConfig tracks which source provided each setting (for debugging).
// See: data-model.md entity #7
type MergedConfig struct {
	Final      Config
	Provenance map[string]ConfigSource
	MergedAt   time.Time
}
```

## Encryption Types

```go
// EncryptedValue represents an encrypted configuration value.
// See: FR-015 through FR-019, data-model.md entity #9
type EncryptedValue struct {
	Ciphertext  []byte
	Nonce       []byte
	KeyID       string
	Algorithm   string
	EncryptedAt time.Time
}
```

## Schema Types

```go
// SettingSchema defines the schema for a single configuration setting.
// See: data-model.md entity #10
type SettingSchema struct {
	Path          string
	Type          string // Type name (e.g., "int", "string", "time.Duration")
	Constraints   []Constraint
	Default       interface{}
	HotReloadable bool
	Description   string
}

// Constraint defines a validation constraint for a setting.
type Constraint struct {
	Type    string      // "range", "enum", "regex", "min", "max", "hexcolor", etc.
	Params  interface{} // Constraint parameters
	Message string      // Error message if constraint fails
}

// ConfigSchema is the complete schema for all configuration settings.
type ConfigSchema struct {
	Settings map[string]SettingSchema
}
```
