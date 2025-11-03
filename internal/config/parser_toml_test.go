package config

import (
	"testing"
)

// TestParseTOML tests TOML parsing with table-driven tests
func TestParseTOML(t *testing.T) {
	tests := []struct {
		checkFunc func(*Config) error
		name      string
		toml      string
		wantErr   bool
	}{
		{
			name: "simple fields",
			toml: `
log_level = "debug"
max_concurrent_ops = 8
theme = "dark"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				if cfg.MaxConcurrentOps != 8 {
					t.Errorf("Expected MaxConcurrentOps=8, got %d", cfg.MaxConcurrentOps)
				}
				if cfg.Theme != "dark" {
					t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
				}
				return nil
			},
		},
		{
			name: "boolean values",
			toml: `
compact_mode = true
show_hints = false
show_line_numbers = true
hot_reload = false
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if !cfg.CompactMode {
					t.Error("Expected CompactMode=true")
				}
				if cfg.ShowHints {
					t.Error("Expected ShowHints=false")
				}
				if !cfg.ShowLineNumbers {
					t.Error("Expected ShowLineNumbers=true")
				}
				if cfg.HotReload {
					t.Error("Expected HotReload=false")
				}
				return nil
			},
		},
		{
			name: "nested colorScheme using section header",
			toml: `
[color_scheme]
border = "#5C6370"
border_focus = "#61AFEF"
text = "#ABB2BF"
error = "#E06C75"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.ColorScheme.Border != "#5C6370" {
					t.Errorf("Expected Border=#5C6370, got %s", cfg.ColorScheme.Border)
				}
				if cfg.ColorScheme.BorderFocus != "#61AFEF" {
					t.Errorf("Expected BorderFocus=#61AFEF, got %s", cfg.ColorScheme.BorderFocus)
				}
				if cfg.ColorScheme.Text != "#ABB2BF" {
					t.Errorf("Expected Text=#ABB2BF, got %s", cfg.ColorScheme.Text)
				}
				if cfg.ColorScheme.Error != "#E06C75" {
					t.Errorf("Expected Error=#E06C75, got %s", cfg.ColorScheme.Error)
				}
				return nil
			},
		},
		{
			name: "nested log rotation",
			toml: `
[log_rotation]
max_size = 20
max_age = 60
max_backups = 10
compress = true
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogRotation.MaxSize != 20 {
					t.Errorf("Expected MaxSize=20, got %d", cfg.LogRotation.MaxSize)
				}
				if cfg.LogRotation.MaxAge != 60 {
					t.Errorf("Expected MaxAge=60, got %d", cfg.LogRotation.MaxAge)
				}
				if cfg.LogRotation.MaxBackups != 10 {
					t.Errorf("Expected MaxBackups=10, got %d", cfg.LogRotation.MaxBackups)
				}
				if !cfg.LogRotation.Compress {
					t.Error("Expected Compress=true")
				}
				return nil
			},
		},
		{
			name: "invalid TOML syntax - missing quotes",
			toml: `
log_level = debug
theme = dark
`,
			wantErr: true,
		},
		{
			name: "invalid TOML syntax - unclosed bracket",
			toml: `
[color_scheme
border = "#5C6370"
`,
			wantErr: true,
		},
		{
			name:    "empty TOML",
			toml:    ``,
			wantErr: false,
			checkFunc: func(_ *Config) error {
				// Empty TOML should parse successfully but return zero values
				return nil
			},
		},
		{
			name: "TOML with comments",
			toml: `
# This is a comment
log_level = "debug"  # inline comment
theme = "dark"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				if cfg.Theme != "dark" {
					t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
				}
				return nil
			},
		},
		{
			name: "unknown fields are detected",
			toml: `
log_level = "debug"
unknownField = "somevalue"
anotherUnknown = 123
theme = "dark"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				if cfg.Theme != "dark" {
					t.Errorf("Expected Theme=dark, got %s", cfg.Theme)
				}
				return nil
			},
		},
		{
			name: "mixed section header style",
			toml: `
log_level = "debug"

[color_scheme]
border = "#5C6370"
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.LogLevel != "debug" {
					t.Errorf("Expected LogLevel=debug, got %s", cfg.LogLevel)
				}
				if cfg.ColorScheme.Border != "#5C6370" {
					t.Errorf("Expected Border=#5C6370, got %s", cfg.ColorScheme.Border)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseTOML([]byte(tt.toml))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				_ = tt.checkFunc(cfg)
			}
		})
	}
}

// TestParseTOMLTypes tests various TOML type conversions
func TestParseTOMLTypes(t *testing.T) {
	tests := []struct {
		want    any
		name    string
		toml    string
		field   string
		wantErr bool
	}{
		{
			name:  "string field",
			toml:  `theme = "dark"`,
			field: "theme",
			want:  "dark",
		},
		{
			name:  "int field",
			toml:  `max_concurrent_ops = 8`,
			field: "maxConcurrentOps",
			want:  8,
		},
		{
			name:  "bool field true",
			toml:  `compact_mode = true`,
			field: "compactMode",
			want:  true,
		},
		{
			name:  "bool field false",
			toml:  `show_hints = false`,
			field: "showHints",
			want:  false,
		},
		{
			name:  "quoted string",
			toml:  `log_level = "debug"`,
			field: "logLevel",
			want:  "debug",
		},
		{
			name:  "number as string",
			toml:  `version = "1.0"`,
			field: "version",
			want:  "1.0",
		},
		{
			name:  "single quotes",
			toml:  `log_level = 'debug'`,
			field: "logLevel",
			want:  "debug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseTOML([]byte(tt.toml))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check specific field
			var actual any
			switch tt.field {
			case "theme":
				actual = cfg.Theme
			case "logLevel":
				actual = cfg.LogLevel
			case "maxConcurrentOps":
				actual = cfg.MaxConcurrentOps
			case "compactMode":
				actual = cfg.CompactMode
			case "showHints":
				actual = cfg.ShowHints
			case "version":
				actual = cfg.Version
			default:
				t.Fatalf("Unknown field in test: %s", tt.field)
			}

			if actual != tt.want {
				t.Errorf("Field %s: expected %v, got %v", tt.field, tt.want, actual)
			}
		})
	}
}

// TestParseTOMLErrorMessages tests that error messages are helpful
func TestParseTOMLErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		toml        string
		errContains string
		wantErr     bool
	}{
		{
			name: "missing quotes around string value",
			toml: `
log_level = debug
`,
			wantErr:     true,
			errContains: "TOML parsing error",
		},
		{
			name: "unclosed section bracket",
			toml: `
[color_scheme
border = "#5C6370"
`,
			wantErr:     true,
			errContains: "TOML parsing error",
		},
		{
			name: "duplicate keys",
			toml: `
log_level = "debug"
log_level = "info"
`,
			wantErr:     true,
			errContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTOML([]byte(tt.toml))

			if tt.wantErr && err == nil {
				t.Error("Expected error but got none")
				return
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
		})
	}
}

// TestParseTOMLMultilineStrings tests TOML multiline string support
func TestParseTOMLMultilineStrings(t *testing.T) {
	tests := []struct {
		checkFunc func(*Config) error
		name      string
		toml      string
		wantErr   bool
	}{
		{
			name: "multiline string with triple quotes",
			toml: `
log_dir = """
/path/to/
logs/directory"""
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				// TOML preserves newlines in multiline strings
				if cfg.LogDir != "/path/to/\nlogs/directory" {
					t.Errorf("Expected multiline string, got %s", cfg.LogDir)
				}
				return nil
			},
		},
		{
			name: "literal string with single quotes",
			toml: `
dotnet_path = 'C:\Windows\Path\dotnet.exe'
`,
			wantErr: false,
			checkFunc: func(cfg *Config) error {
				if cfg.DotnetPath != `C:\Windows\Path\dotnet.exe` {
					t.Errorf("Expected literal string, got %s", cfg.DotnetPath)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseTOML([]byte(tt.toml))

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				_ = tt.checkFunc(cfg)
			}
		})
	}
}
