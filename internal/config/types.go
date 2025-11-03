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
	LoadedAt          time.Time             `yaml:"-" toml:"-"`
	Keybindings       map[string]KeyBinding `yaml:"keybindings" toml:"keybindings"`
	ColorScheme       ColorScheme           `yaml:"colorScheme" toml:"color_scheme"`
	DotnetPath        string                `yaml:"dotnetPath" toml:"dotnet_path" default:""`
	DotnetVerbosity   string                `yaml:"dotnetVerbosity" toml:"dotnet_verbosity" validate:"oneof=quiet minimal normal detailed diagnostic" default:"minimal"`
	LogFormat         string                `yaml:"logFormat" toml:"log_format" validate:"oneof=text json" default:"text"`
	LogDir            string                `yaml:"logDir" toml:"log_dir" default:""`
	LogLevel          string                `yaml:"logLevel" toml:"log_level" validate:"oneof=debug info warn error" default:"info"`
	DateFormat        string                `yaml:"dateFormat" toml:"date_format" validate:"dateformat" default:"2006-01-02"`
	LoadedFrom        string                `yaml:"-" toml:"-"`
	KeybindingProfile string                `yaml:"keybindingProfile" toml:"keybinding_profile" validate:"oneof=default vim emacs" default:"default"`
	Theme             string                `yaml:"theme" toml:"theme" validate:"oneof=default dark light solarized" default:"default"`
	Version           string                `yaml:"version" toml:"version"`
	LogRotation       LogRotation           `yaml:"logRotation" toml:"log_rotation"`
	Timeouts          Timeouts              `yaml:"timeouts" toml:"timeouts"`
	RefreshInterval   time.Duration         `yaml:"refreshInterval" toml:"refresh_interval" validate:"min=0" default:"0"`
	CacheSize         int                   `yaml:"cacheSize" toml:"cache_size" validate:"min=0" default:"50"`
	MaxConcurrentOps  int                   `yaml:"maxConcurrentOps" toml:"max_concurrent_ops" validate:"min=1,max=16" default:"4"`
	ShowLineNumbers   bool                  `yaml:"showLineNumbers" toml:"show_line_numbers" default:"false"`
	ShowHints         bool                  `yaml:"showHints" toml:"show_hints" default:"true"`
	CompactMode       bool                  `yaml:"compactMode" toml:"compact_mode" default:"false"`
	HotReload         bool                  `yaml:"hotReload" toml:"hot_reload" default:"false"`
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
	LoadedAt   time.Time
	Data       map[string]any
	Name       string
	FilePath   string
	Precedence int
}

// MergedConfig tracks which source provided each setting (for debugging).
// See: specs/002-config-management/data-model.md entity #7
type MergedConfig struct {
	MergedAt   time.Time
	Provenance map[string]ConfigSource
	Final      Config
}

// ValidationError describes a configuration validation failure.
// See: specs/002-config-management/data-model.md entity #8
type ValidationError struct {
	Value        any
	DefaultUsed  any
	Key          string
	Constraint   string
	SuggestedFix string
	Severity     string
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
	EncryptedAt time.Time
	KeyID       string
	Algorithm   string
	Ciphertext  []byte
	Nonce       []byte
}

// String returns a safe string representation that never exposes plaintext.
// See: FR-018
func (ev *EncryptedValue) String() string {
	return fmt.Sprintf("EncryptedValue{KeyID=%s, Algorithm=%s, CiphertextLen=%d, NonceLen=%d}",
		ev.KeyID, ev.Algorithm, len(ev.Ciphertext), len(ev.Nonce))
}

// SettingSchema defines the schema for a single configuration setting.
// See: specs/002-config-management/data-model.md entity #10
type SettingSchema struct {
	Type          reflect.Type
	Default       any
	Path          string
	Description   string
	Constraints   []Constraint
	HotReloadable bool
}

// Constraint defines a validation constraint for a setting.
type Constraint struct {
	Type    string
	Params  any
	Message string
}

// ConfigSchema is the complete schema for all configuration settings.
type ConfigSchema struct {
	Settings map[string]SettingSchema
}
