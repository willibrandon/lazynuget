package bootstrap

import (
	"testing"
)

// TestParseFlags tests command-line flag parsing
func TestParseFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		want       Flags
		shouldExit bool
	}{
		{
			name: "no flags",
			args: []string{},
			want: Flags{
				ShowVersion:    false,
				ShowHelp:       false,
				ConfigPath:     "",
				LogLevel:       "info",
				NonInteractive: false,
			},
			shouldExit: false,
		},
		{
			name: "version flag",
			args: []string{"-version"},
			want: Flags{
				ShowVersion: true,
			},
			shouldExit: true,
		},
		{
			name: "help flag",
			args: []string{"-help"},
			want: Flags{
				ShowHelp: true,
			},
			shouldExit: true,
		},
		{
			name: "config path",
			args: []string{"-config", "/path/to/config.yml"},
			want: Flags{
				ConfigPath: "/path/to/config.yml",
			},
			shouldExit: false,
		},
		{
			name: "log level",
			args: []string{"-log-level", "debug"},
			want: Flags{
				LogLevel: "debug",
			},
			shouldExit: false,
		},
		{
			name: "non-interactive",
			args: []string{"-non-interactive"},
			want: Flags{
				NonInteractive: true,
			},
			shouldExit: false,
		},
		{
			name: "multiple flags",
			args: []string{"-log-level", "warn", "-non-interactive", "-config", "/custom/config.toml"},
			want: Flags{
				LogLevel:       "warn",
				NonInteractive: true,
				ConfigPath:     "/custom/config.toml",
			},
			shouldExit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewApp("test", "test-commit", "2025-01-01")
			if err != nil {
				t.Fatalf("NewApp() failed: %v", err)
			}
			defer app.cancel()

			flags, shouldExit, err := app.ParseFlags(tt.args)
			if err != nil {
				t.Fatalf("ParseFlags() error = %v", err)
			}

			if shouldExit != tt.shouldExit {
				t.Errorf("shouldExit = %v, want %v", shouldExit, tt.shouldExit)
			}

			if flags.ShowVersion != tt.want.ShowVersion {
				t.Errorf("ShowVersion = %v, want %v", flags.ShowVersion, tt.want.ShowVersion)
			}
			if flags.ShowHelp != tt.want.ShowHelp {
				t.Errorf("ShowHelp = %v, want %v", flags.ShowHelp, tt.want.ShowHelp)
			}
			if tt.want.ConfigPath != "" && flags.ConfigPath != tt.want.ConfigPath {
				t.Errorf("ConfigPath = %v, want %v", flags.ConfigPath, tt.want.ConfigPath)
			}
			if tt.want.LogLevel != "" && flags.LogLevel != tt.want.LogLevel {
				t.Errorf("LogLevel = %v, want %v", flags.LogLevel, tt.want.LogLevel)
			}
			if flags.NonInteractive != tt.want.NonInteractive {
				t.Errorf("NonInteractive = %v, want %v", flags.NonInteractive, tt.want.NonInteractive)
			}
		})
	}
}

// TestParseFlagsDefaults tests that defaults are applied correctly
func TestParseFlagsDefaults(t *testing.T) {
	app, err := NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}
	defer app.cancel()

	flags, _, err := app.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if flags.ShowVersion {
		t.Error("ShowVersion should default to false")
	}
	if flags.ShowHelp {
		t.Error("ShowHelp should default to false")
	}
	if flags.ConfigPath != "" {
		t.Error("ConfigPath should default to empty string")
	}
	if flags.LogLevel != "info" {
		t.Errorf("LogLevel should default to 'info', got %q", flags.LogLevel)
	}
	if flags.NonInteractive {
		t.Error("NonInteractive should default to false")
	}
}

// TestShowHelp tests the help display function
func TestShowHelp(_ *testing.T) {
	// ShowHelp should not panic
	ShowHelp()
}
