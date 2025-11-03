// Package config provides configuration loading, merging, validation, and hot-reload
// functionality for the LazyNuGet application.
package config

import (
	"fmt"
	"reflect"
	"time"
)

// Config is the root configuration object containing all application settings.
// See: specs/002-config-management/data-model.md entity #1
type Config struct {
	// Meta
	Version    string    `yaml:"version" toml:"version"`
	LoadedFrom string    `yaml:"-" toml:"-"`
	LoadedAt   time.Time `yaml:"-" toml:"-"`

	// UI Settings (FR-020 through FR-025)
	Theme           string      `yaml:"theme" toml:"theme" validate:"oneof=default dark light solarized" default:"default"`
	ColorScheme     ColorScheme `yaml:"colorScheme" toml:"color_scheme"`
	CompactMode     bool        `yaml:"compactMode" toml:"compact_mode" default:"false"`
	ShowHints       bool        `yaml:"showHints" toml:"show_hints" default:"true"`
	ShowLineNumbers bool        `yaml:"showLineNumbers" toml:"show_line_numbers" default:"false"`
	DateFormat      string      `yaml:"dateFormat" toml:"date_format" validate:"dateformat" default:"2006-01-02"`

	// Keybindings (FR-026 through FR-030)
	Keybindings       map[string]KeyBinding `yaml:"keybindings" toml:"keybindings"`
	KeybindingProfile string                `yaml:"keybindingProfile" toml:"keybinding_profile" validate:"oneof=default vim emacs" default:"default"`

	// Performance (FR-031 through FR-034)
	MaxConcurrentOps int           `yaml:"maxConcurrentOps" toml:"max_concurrent_ops" validate:"min=1,max=16" default:"4"`
	CacheSize        int           `yaml:"cacheSize" toml:"cache_size" validate:"min=0" default:"50"` // MB
	RefreshInterval  time.Duration `yaml:"refreshInterval" toml:"refresh_interval" validate:"min=0" default:"0"`
	Timeouts         Timeouts      `yaml:"timeouts" toml:"timeouts"`

	// Dotnet CLI Integration (FR-035 through FR-038)
	DotnetPath      string `yaml:"dotnetPath" toml:"dotnet_path" default:""`
	DotnetVerbosity string `yaml:"dotnetVerbosity" toml:"dotnet_verbosity" validate:"oneof=quiet minimal normal detailed diagnostic" default:"minimal"`

	// Logging (FR-039 through FR-042)
	LogLevel    string      `yaml:"logLevel" toml:"log_level" validate:"oneof=debug info warn error" default:"info"`
	LogDir      string      `yaml:"logDir" toml:"log_dir" default:""`
	LogFormat   string      `yaml:"logFormat" toml:"log_format" validate:"oneof=text json" default:"text"`
	LogRotation LogRotation `yaml:"logRotation" toml:"log_rotation"`

	// Hot-Reload (FR-043 through FR-049)
	HotReload bool `yaml:"hotReload" toml:"hot_reload" default:"false"`
}

// ColorScheme defines customizable colors for UI elements.
// See: specs/002-config-management/data-model.md entity #2
type ColorScheme struct {
	Border      string `yaml:"border" toml:"border" validate:"hexcolor" default:"#FFFFFF"`
	BorderFocus string `yaml:"borderFocus" toml:"border_focus" validate:"hexcolor" default:"#00FF00"`
	Text        string `yaml:"text" toml:"text" validate:"hexcolor" default:"#FFFFFF"`
	TextDim     string `yaml:"textDim" toml:"text_dim" validate:"hexcolor" default:"#808080"`
	Background  string `yaml:"background" toml:"background" validate:"hexcolor" default:"#000000"`
	Highlight   string `yaml:"highlight" toml:"highlight" validate:"hexcolor" default:"#FFFF00"`
	Error       string `yaml:"error" toml:"error" validate:"hexcolor" default:"#FF0000"`
	Warning     string `yaml:"warning" toml:"warning" validate:"hexcolor" default:"#FFA500"`
	Success     string `yaml:"success" toml:"success" validate:"hexcolor" default:"#00FF00"`
	Info        string `yaml:"info" toml:"info" validate:"hexcolor" default:"#00FFFF"`
}

// KeyBinding maps an action to a key combination.
// See: specs/002-config-management/data-model.md entity #3
type KeyBinding struct {
	Action      string `yaml:"action" toml:"action"`
	Key         string `yaml:"key" toml:"key"`
	Description string `yaml:"description" toml:"description"`
	Context     string `yaml:"context" toml:"context"`
}

// Timeouts defines timeout durations for different operation types.
// See: specs/002-config-management/data-model.md entity #4
type Timeouts struct {
	NetworkRequest time.Duration `yaml:"networkRequest" toml:"network_request" validate:"min=1s" default:"30s"`
	DotnetCLI      time.Duration `yaml:"dotnetCLI" toml:"dotnet_cli" validate:"min=1s" default:"60s"`
	FileOperation  time.Duration `yaml:"fileOperation" toml:"file_operation" validate:"min=100ms" default:"5s"`
}

// LogRotation configures log file rotation.
// See: specs/002-config-management/data-model.md entity #5
type LogRotation struct {
	MaxSize    int  `yaml:"maxSize" toml:"max_size" validate:"min=1" default:"10"`
	MaxAge     int  `yaml:"maxAge" toml:"max_age" validate:"min=1" default:"30"`
	MaxBackups int  `yaml:"maxBackups" toml:"max_backups" validate:"min=0" default:"3"`
	Compress   bool `yaml:"compress" toml:"compress" default:"true"`
}

// ConfigSource represents one of the four configuration sources.
// See: specs/002-config-management/data-model.md entity #6
type ConfigSource struct {
	Name       string
	Precedence int
	Data       map[string]interface{}
	FilePath   string
	LoadedAt   time.Time
}

// MergedConfig tracks which source provided each setting (for debugging).
// See: specs/002-config-management/data-model.md entity #7
type MergedConfig struct {
	Final      Config
	Provenance map[string]ConfigSource
	MergedAt   time.Time
}

// ValidationError describes a configuration validation failure.
// See: specs/002-config-management/data-model.md entity #8
type ValidationError struct {
	Key          string
	Value        interface{}
	Constraint   string
	SuggestedFix string
	Severity     string
	DefaultUsed  interface{}
}

// Error implements the error interface for ValidationError.
func (ve ValidationError) Error() string {
	if ve.Severity == "error" {
		return ve.Key + ": " + ve.Constraint
	}
	return fmt.Sprintf("%s: %s (using default: %v)", ve.Key, ve.Constraint, ve.DefaultUsed)
}

// EncryptedValue represents an encrypted configuration value.
// See: specs/002-config-management/data-model.md entity #9
type EncryptedValue struct {
	Ciphertext  []byte
	Nonce       []byte
	KeyID       string
	Algorithm   string
	EncryptedAt time.Time
}

// SettingSchema defines the schema for a single configuration setting.
// See: specs/002-config-management/data-model.md entity #10
type SettingSchema struct {
	Path          string
	Type          reflect.Type
	Constraints   []Constraint
	Default       interface{}
	HotReloadable bool
	Description   string
}

// Constraint defines a validation constraint for a setting.
type Constraint struct {
	Type    string
	Params  interface{}
	Message string
}

// ConfigSchema is the complete schema for all configuration settings.
type ConfigSchema struct {
	Settings map[string]SettingSchema
}
