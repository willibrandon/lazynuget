package config

import (
	"strings"
	"testing"
)

// TestApplyTopLevelSetting tests applying top-level settings from env vars
func TestApplyTopLevelSetting(t *testing.T) {
	tests := []struct {
		wantValue any
		initial   *Config
		name      string
		value     string
		wantField string
		parts     []string
	}{
		{
			name:      "set log level",
			parts:     []string{"logLevel"},
			value:     "debug",
			initial:   &Config{LogLevel: "info"},
			wantField: "LogLevel",
			wantValue: "debug",
		},
		{
			name:      "set theme",
			parts:     []string{"theme"},
			value:     "dark",
			initial:   &Config{Theme: "default"},
			wantField: "Theme",
			wantValue: "dark",
		},
		{
			name:      "set compact mode",
			parts:     []string{"compactMode"},
			value:     "true",
			initial:   &Config{CompactMode: false},
			wantField: "CompactMode",
			wantValue: true,
		},
		{
			name:      "set show hints",
			parts:     []string{"showHints"},
			value:     "false",
			initial:   &Config{ShowHints: true},
			wantField: "ShowHints",
			wantValue: false,
		},
		{
			name:      "set max concurrent ops",
			parts:     []string{"maxConcurrentOps"},
			value:     "8",
			initial:   &Config{MaxConcurrentOps: 4},
			wantField: "MaxConcurrentOps",
			wantValue: 8,
		},
		{
			name:      "set dotnet path",
			parts:     []string{"dotnetPath"},
			value:     "/custom/dotnet",
			initial:   &Config{DotnetPath: "dotnet"},
			wantField: "DotnetPath",
			wantValue: "/custom/dotnet",
		},
		{
			name:      "set log format",
			parts:     []string{"logFormat"},
			value:     "json",
			initial:   &Config{LogFormat: "text"},
			wantField: "LogFormat",
			wantValue: "json",
		},
		{
			name:      "set hot reload",
			parts:     []string{"hotReload"},
			value:     "true",
			initial:   &Config{HotReload: false},
			wantField: "HotReload",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.initial
			path := strings.Join(tt.parts, ".")
			applyEnvVarValue(cfg, path, tt.value)

			// Check the field was set correctly
			switch tt.wantField {
			case "LogLevel":
				if cfg.LogLevel != tt.wantValue.(string) {
					t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.wantValue)
				}
			case "Theme":
				if cfg.Theme != tt.wantValue.(string) {
					t.Errorf("Theme = %v, want %v", cfg.Theme, tt.wantValue)
				}
			case "CompactMode":
				if cfg.CompactMode != tt.wantValue.(bool) {
					t.Errorf("CompactMode = %v, want %v", cfg.CompactMode, tt.wantValue)
				}
			case "ShowHints":
				if cfg.ShowHints != tt.wantValue.(bool) {
					t.Errorf("ShowHints = %v, want %v", cfg.ShowHints, tt.wantValue)
				}
			case "MaxConcurrentOps":
				if cfg.MaxConcurrentOps != tt.wantValue.(int) {
					t.Errorf("MaxConcurrentOps = %v, want %v", cfg.MaxConcurrentOps, tt.wantValue)
				}
			case "DotnetPath":
				if cfg.DotnetPath != tt.wantValue.(string) {
					t.Errorf("DotnetPath = %v, want %v", cfg.DotnetPath, tt.wantValue)
				}
			case "LogFormat":
				if cfg.LogFormat != tt.wantValue.(string) {
					t.Errorf("LogFormat = %v, want %v", cfg.LogFormat, tt.wantValue)
				}
			case "HotReload":
				if cfg.HotReload != tt.wantValue.(bool) {
					t.Errorf("HotReload = %v, want %v", cfg.HotReload, tt.wantValue)
				}
			}
		})
	}
}

// TestApplyNestedSetting tests applying nested settings from env vars
func TestApplyNestedSetting(t *testing.T) {
	tests := []struct {
		initial *Config
		check   func(*testing.T, *Config)
		name    string
		value   string
		parts   []string
	}{
		// ColorScheme tests
		{
			name:    "set color scheme border",
			parts:   []string{"colorScheme", "border"},
			value:   "#FF0000",
			initial: &Config{ColorScheme: ColorScheme{Border: "#FFFFFF"}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Border != "#FF0000" {
					t.Errorf("Border = %v, want #FF0000", cfg.ColorScheme.Border)
				}
			},
		},
		{
			name:    "set color scheme error",
			parts:   []string{"colorScheme", "error"},
			value:   "#FF0000",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Error != "#FF0000" {
					t.Errorf("Error = %v, want #FF0000", cfg.ColorScheme.Error)
				}
			},
		},
		{
			name:    "set color scheme warning",
			parts:   []string{"colorScheme", "warning"},
			value:   "#FFA500",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Warning != "#FFA500" {
					t.Errorf("Warning = %v, want #FFA500", cfg.ColorScheme.Warning)
				}
			},
		},
		{
			name:    "set color scheme success",
			parts:   []string{"colorScheme", "success"},
			value:   "#00FF00",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Success != "#00FF00" {
					t.Errorf("Success = %v, want #00FF00", cfg.ColorScheme.Success)
				}
			},
		},
		{
			name:    "set color scheme info",
			parts:   []string{"colorScheme", "info"},
			value:   "#0000FF",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Info != "#0000FF" {
					t.Errorf("Info = %v, want #0000FF", cfg.ColorScheme.Info)
				}
			},
		},
		{
			name:    "set color scheme highlight",
			parts:   []string{"colorScheme", "highlight"},
			value:   "#FFFF00",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Highlight != "#FFFF00" {
					t.Errorf("Highlight = %v, want #FFFF00", cfg.ColorScheme.Highlight)
				}
			},
		},
		{
			name:    "set color scheme background",
			parts:   []string{"colorScheme", "background"},
			value:   "#000000",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Background != "#000000" {
					t.Errorf("Background = %v, want #000000", cfg.ColorScheme.Background)
				}
			},
		},
		{
			name:    "set color scheme text",
			parts:   []string{"colorScheme", "text"},
			value:   "#00FF00",
			initial: &Config{ColorScheme: ColorScheme{Text: "#FFFFFF"}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.Text != "#00FF00" {
					t.Errorf("Text = %v, want #00FF00", cfg.ColorScheme.Text)
				}
			},
		},
		{
			name:    "set color scheme textDim",
			parts:   []string{"colorScheme", "textDim"},
			value:   "#808080",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.TextDim != "#808080" {
					t.Errorf("TextDim = %v, want #808080", cfg.ColorScheme.TextDim)
				}
			},
		},
		{
			name:    "set color scheme borderFocus",
			parts:   []string{"colorScheme", "borderFocus"},
			value:   "#00FFFF",
			initial: &Config{ColorScheme: ColorScheme{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.ColorScheme.BorderFocus != "#00FFFF" {
					t.Errorf("BorderFocus = %v, want #00FFFF", cfg.ColorScheme.BorderFocus)
				}
			},
		},
		// Timeouts tests
		{
			name:    "set timeout networkRequest",
			parts:   []string{"timeouts", "networkRequest"},
			value:   "5s",
			initial: &Config{Timeouts: Timeouts{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Timeouts.NetworkRequest.String() != "5s" {
					t.Errorf("NetworkRequest = %v, want 5s", cfg.Timeouts.NetworkRequest)
				}
			},
		},
		{
			name:    "set timeout dotnetCli",
			parts:   []string{"timeouts", "dotnetCli"},
			value:   "10s",
			initial: &Config{Timeouts: Timeouts{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Timeouts.DotnetCLI.String() != "10s" {
					t.Errorf("DotnetCLI = %v, want 10s", cfg.Timeouts.DotnetCLI)
				}
			},
		},
		{
			name:    "set timeout fileOperation",
			parts:   []string{"timeouts", "fileOperation"},
			value:   "30s",
			initial: &Config{Timeouts: Timeouts{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.Timeouts.FileOperation.String() != "30s" {
					t.Errorf("FileOperation = %v, want 30s", cfg.Timeouts.FileOperation)
				}
			},
		},
		// LogRotation tests
		{
			name:    "set logRotation maxSize",
			parts:   []string{"logRotation", "maxSize"},
			value:   "100",
			initial: &Config{LogRotation: LogRotation{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.LogRotation.MaxSize != 100 {
					t.Errorf("MaxSize = %v, want 100", cfg.LogRotation.MaxSize)
				}
			},
		},
		{
			name:    "set logRotation maxAge",
			parts:   []string{"logRotation", "maxAge"},
			value:   "30",
			initial: &Config{LogRotation: LogRotation{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.LogRotation.MaxAge != 30 {
					t.Errorf("MaxAge = %v, want 30", cfg.LogRotation.MaxAge)
				}
			},
		},
		{
			name:    "set logRotation maxBackups",
			parts:   []string{"logRotation", "maxBackups"},
			value:   "5",
			initial: &Config{LogRotation: LogRotation{}},
			check: func(t *testing.T, cfg *Config) {
				if cfg.LogRotation.MaxBackups != 5 {
					t.Errorf("MaxBackups = %v, want 5", cfg.LogRotation.MaxBackups)
				}
			},
		},
		{
			name:    "set logRotation compress",
			parts:   []string{"logRotation", "compress"},
			value:   "true",
			initial: &Config{LogRotation: LogRotation{}},
			check: func(t *testing.T, cfg *Config) {
				if !cfg.LogRotation.Compress {
					t.Error("Compress should be true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.initial
			path := strings.Join(tt.parts, ".")
			applyEnvVarValue(cfg, path, tt.value)
			tt.check(t, cfg)
		})
	}
}

// TestApplyDoubleNestedSetting tests the double-nested setting function
func TestApplyDoubleNestedSetting(t *testing.T) {
	cfg := &Config{}

	// applyDoubleNestedSetting is a placeholder that doesn't do anything yet
	// Test that it doesn't error
	err := applyDoubleNestedSetting(cfg, "parent", "child", "grandchild", "value")
	if err != nil {
		t.Errorf("applyDoubleNestedSetting should not error: %v", err)
	}
}
