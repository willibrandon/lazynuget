package bootstrap

import (
	"context"
	"testing"
	"time"
)

func TestNewApp(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
		wantErr bool
	}{
		{
			name:    "valid version info",
			version: "1.0.0",
			commit:  "abc123",
			date:    "2025-01-01",
			wantErr: false,
		},
		{
			name:    "empty version info",
			version: "",
			commit:  "",
			date:    "",
			wantErr: false,
		},
		{
			name:    "partial version info",
			version: "dev",
			commit:  "unknown",
			date:    "unknown",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewApp(tt.version, tt.commit, tt.date)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewApp() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewApp() unexpected error: %v", err)
				return
			}

			if app == nil {
				t.Fatal("NewApp() returned nil app")
			}

			if app.version.Version != tt.version {
				t.Errorf("version = %v, want %v", app.version.Version, tt.version)
			}
			if app.version.Commit != tt.commit {
				t.Errorf("commit = %v, want %v", app.version.Commit, tt.commit)
			}
			if app.version.Date != tt.date {
				t.Errorf("date = %v, want %v", app.version.Date, tt.date)
			}

			if app.lifecycle == nil {
				t.Error("lifecycle manager not initialized")
			}

			if app.ctx == nil {
				t.Error("context not initialized")
			}
			if app.cancel == nil {
				t.Error("cancel func not initialized")
			}

			app.cancel()
		})
	}
}

func TestBootstrap(t *testing.T) {
	tests := []struct {
		flags   *Flags
		name    string
		wantErr bool
	}{
		{
			name:    "bootstrap with nil flags",
			flags:   nil,
			wantErr: false,
		},
		{
			name: "bootstrap with default flags",
			flags: &Flags{
				ShowVersion:    false,
				ShowHelp:       false,
				ConfigPath:     "",
				LogLevel:       "",
				NonInteractive: false,
			},
			wantErr: false,
		},
		{
			name: "bootstrap with custom log level",
			flags: &Flags{
				LogLevel:       "debug",
				NonInteractive: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, err := NewApp("test", "test-commit", "2025-01-01")
			if err != nil {
				t.Fatalf("NewApp() failed: %v", err)
			}
			defer app.cancel()

			err = app.Bootstrap(tt.flags)

			if tt.wantErr {
				if err == nil {
					t.Error("Bootstrap() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Bootstrap() unexpected error: %v", err)
				return
			}

			if app.config == nil {
				t.Error("config not loaded")
			}

			if app.logger == nil {
				t.Error("logger not initialized")
			}

			if app.platform == nil {
				t.Error("platform not initialized")
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	app, err := NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	stateStr := app.lifecycle.GetState().String()
	if stateStr != "running" && stateStr != "Running" {
		t.Errorf("expected app to be in running state after bootstrap, got: %v", stateStr)
	}

	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

	select {
	case <-app.ctx.Done():
		// Context was cancelled as expected
	case <-time.After(100 * time.Millisecond):
		t.Error("context was not cancelled after shutdown")
	}
}

func TestGetConfig(t *testing.T) {
	app, err := NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}
	defer app.cancel()

	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	cfg := app.GetConfig()
	if cfg == nil {
		t.Error("GetConfig() returned nil")
	}
}

func TestGetLogger(t *testing.T) {
	app, err := NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}
	defer app.cancel()

	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	logger := app.GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}
}

func TestGetRunMode(t *testing.T) {
	app, err := NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}
	defer app.cancel()

	if err := app.Bootstrap(&Flags{NonInteractive: true}); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	runMode := app.GetRunMode()
	if !runMode.IsInteractive() && runMode.String() != "non-interactive" {
		t.Errorf("expected non-interactive mode, got: %v", runMode)
	}
}

func TestRegisterShutdownHandler(t *testing.T) {
	app, err := NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}
	defer app.cancel()

	handlerCalled := false
	app.RegisterShutdownHandler("test-handler", 100, func(_ context.Context) error {
		handlerCalled = true
		return nil
	})

	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	if err := app.Shutdown(); err != nil {
		t.Fatalf("Shutdown() failed: %v", err)
	}

	if !handlerCalled {
		t.Error("shutdown handler was not called")
	}
}
