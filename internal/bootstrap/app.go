package bootstrap

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
	"github.com/willibrandon/lazynuget/internal/lifecycle"
	"github.com/willibrandon/lazynuget/internal/logging"
	"github.com/willibrandon/lazynuget/internal/platform"
)

// App represents the running LazyNuGet application instance.
type App struct {
	// config holds merged configuration from all sources
	config *config.AppConfig

	// logger provides structured logging
	logger logging.Logger

	// platform provides platform detection and utilities
	platform platform.Platform

	// lifecycle manages application state transitions
	lifecycle *lifecycle.Manager

	// runMode determines if the app runs interactively or non-interactively
	runMode platform.RunMode

	// gui is the Bubbletea TUI program (only initialized in interactive mode)
	gui interface{}

	// ctx is the root cancellation context
	ctx context.Context

	// cancel is the cancellation function for shutdown
	cancel context.CancelFunc

	// version contains build information
	version VersionInfo

	// startTime is when the application was created
	startTime time.Time

	// guiOnce ensures GUI is initialized only once (lazy initialization)
	guiOnce sync.Once

	// phase tracks the current initialization phase for error context
	phase string
}

// NewApp creates a new application instance with version information.
func NewApp(version, commit, date string) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Create lifecycle manager with 30-second shutdown timeout
	lifecycleMgr := lifecycle.NewManager(30 * time.Second)

	app := &App{
		ctx:       ctx,
		cancel:    cancel,
		version:   VersionInfo{Version: version, Commit: commit, Date: date},
		startTime: time.Now(),
		lifecycle: lifecycleMgr,
		phase:     "uninitialized",
	}

	return app, nil
}

// Bootstrap initializes all application subsystems in the correct order.
// This method implements Layer 2 panic recovery with phase tracking.
func (app *App) Bootstrap(flags *Flags) error {
	// Layer 2 panic recovery: catch panics and add phase context
	defer func() {
		if r := recover(); r != nil {
			// Mark lifecycle as failed
			if app.lifecycle != nil {
				app.lifecycle.SetState(lifecycle.StateFailed)
			}
			if app.logger != nil {
				app.logger.Error("PANIC during bootstrap (phase: %s): %v", app.phase, r)
			}
			// Re-panic for main() to handle
			panic(r)
		}
	}()

	// Transition to initializing state
	if err := app.lifecycle.SetState(lifecycle.StateInitializing); err != nil {
		return fmt.Errorf("failed to enter initializing state: %w", err)
	}

	// Phase: Config loading
	app.phase = "config"

	// Convert bootstrap.Flags to config.Flags
	var configFlags *config.Flags
	if flags != nil {
		configFlags = &config.Flags{
			ShowVersion:    flags.ShowVersion,
			ShowHelp:       flags.ShowHelp,
			ConfigPath:     flags.ConfigPath,
			LogLevel:       flags.LogLevel,
			NonInteractive: flags.NonInteractive,
		}
	}

	cfg, err := config.Load(configFlags)
	if err != nil {
		app.lifecycle.SetState(lifecycle.StateFailed)
		return fmt.Errorf("configuration loading failed: %w", err)
	}
	app.config = cfg

	// Phase: Logging setup
	app.phase = "logging"
	// For now, log to stdout only (file logging can be added later)
	app.logger = logging.New(app.config.LogLevel, "")

	// Phase: Directory permission checking
	app.phase = "directory-permissions"
	app.checkDirectoryPermissions()

	// Phase: Platform detection
	app.phase = "platform"
	app.platform = platform.New(app.config, app.logger)

	// Phase: Determine run mode (interactive vs non-interactive)
	app.phase = "runmode"
	app.runMode = platform.DetermineRunMode(app.config.NonInteractive)
	app.logger.Info("Run mode determined: %s", app.runMode)

	// Phase: Dotnet CLI validation (async, non-blocking)
	app.phase = "dotnet-validation"
	// Launch dotnet validation in background - don't block startup
	go func() {
		if err := platform.ValidateDotnetCLI(); err != nil {
			app.logger.Warn("Dotnet CLI validation warning: %v", err)
			// Don't fail startup - just warn the user
		} else {
			app.logger.Debug("Dotnet CLI validated successfully")
		}
	}()

	// Transition to running state
	app.phase = "ready"
	if err := app.lifecycle.SetState(lifecycle.StateRunning); err != nil {
		return fmt.Errorf("failed to enter running state: %w", err)
	}

	app.logger.Info("Bootstrap complete, application is running")
	return nil
}

// GetConfig returns the application configuration.
func (app *App) GetConfig() *config.AppConfig {
	return app.config
}

// GetLogger returns the application logger.
func (app *App) GetLogger() logging.Logger {
	return app.logger
}

// GetPlatform returns the platform utilities.
func (app *App) GetPlatform() platform.Platform {
	return app.platform
}

// GetRunMode returns the determined run mode.
func (app *App) GetRunMode() platform.RunMode {
	return app.runMode
}

// GetGUI returns the GUI instance, initializing it lazily if in interactive mode.
// Returns nil if in non-interactive mode.
func (app *App) GetGUI() interface{} {
	if !app.runMode.IsInteractive() {
		return nil
	}

	app.guiOnce.Do(func() {
		// TODO: Initialize Bubbletea TUI here when GUI is implemented
		app.logger.Debug("GUI initialization deferred (not yet implemented)")
	})

	return app.gui
}

// Run starts the application and waits for shutdown signal
func (app *App) Run() error {
	// Verify we're in running state
	if app.lifecycle.GetState() != lifecycle.StateRunning {
		return fmt.Errorf("cannot run: application not in running state (current: %s)", app.lifecycle.GetState())
	}

	app.logger.Info("Application started, waiting for shutdown signal...")

	// Create signal handler
	signalHandler := lifecycle.NewSignalHandler(app.lifecycle, app.logger)

	// Wait for shutdown signal (this blocks)
	shutdownCtx := signalHandler.WaitForShutdownSignal(app.ctx)

	// Block until context is cancelled
	<-shutdownCtx.Done()

	app.logger.Info("Shutdown signal received")

	// Perform graceful shutdown
	return app.Shutdown()
}

// Shutdown performs graceful shutdown of all subsystems
func (app *App) Shutdown() error {
	app.logger.Info("Beginning graceful shutdown...")

	// Create a fresh context for shutdown (not the cancelled app context)
	shutdownCtx := context.Background()

	// Execute lifecycle shutdown with all registered handlers
	if err := app.lifecycle.Shutdown(shutdownCtx, app.logger); err != nil {
		app.logger.Error("Shutdown completed with errors: %v", err)
		// Cancel the app context even if shutdown had errors
		app.cancel()
		return err
	}

	// Cancel the application context after successful shutdown
	app.cancel()

	app.logger.Info("Shutdown complete")
	return nil
}

// RegisterShutdownHandler registers a function to be called during shutdown
func (app *App) RegisterShutdownHandler(name string, priority int, handler func(context.Context) error) {
	app.lifecycle.RegisterShutdownHandler(lifecycle.ShutdownHandler{
		Name:     name,
		Priority: priority,
		Handler:  handler,
	})
}

// checkDirectoryPermissions verifies that config directories are writable
// If permissions are insufficient, warns and attempts to use temp directory fallback
func (app *App) checkDirectoryPermissions() {
	directories := []struct {
		name string
		path string
	}{
		{"config", app.config.ConfigDir},
		{"log", app.config.LogDir},
		{"cache", app.config.CacheDir},
	}

	for _, dir := range directories {
		// Check if directory exists
		info, err := os.Stat(dir.path)
		if err != nil {
			if os.IsNotExist(err) {
				// Try to create the directory
				if err := os.MkdirAll(dir.path, 0755); err != nil {
					app.logger.Warn("Cannot create %s directory %s: %v\nFalling back to temp directory", dir.name, dir.path, err)
					app.useTempDirectoryFallback(dir.name)
					continue
				}
				app.logger.Debug("Created %s directory: %s", dir.name, dir.path)
			} else {
				app.logger.Warn("Cannot access %s directory %s: %v", dir.name, dir.path, err)
				continue
			}
		} else if !info.IsDir() {
			app.logger.Warn("%s path %s exists but is not a directory\nFalling back to temp directory", dir.name, dir.path)
			app.useTempDirectoryFallback(dir.name)
			continue
		}

		// Test write permissions by creating a temp file
		testFile := filepath.Join(dir.path, ".lazynuget-write-test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			app.logger.Warn("Cannot write to %s directory %s: %v\nFalling back to temp directory", dir.name, dir.path, err)
			app.useTempDirectoryFallback(dir.name)
			continue
		}
		// Clean up test file
		os.Remove(testFile)

		app.logger.Debug("%s directory verified: %s", dir.name, dir.path)
	}
}

// useTempDirectoryFallback updates config to use temp directory for the specified type
func (app *App) useTempDirectoryFallback(dirType string) {
	tempBase := os.TempDir()
	fallbackPath := filepath.Join(tempBase, "lazynuget", dirType)

	// Create the fallback directory
	if err := os.MkdirAll(fallbackPath, 0755); err != nil {
		app.logger.Error("Cannot create fallback %s directory %s: %v", dirType, fallbackPath, err)
		return
	}

	// Update config with fallback path
	switch dirType {
	case "config":
		app.config.ConfigDir = fallbackPath
	case "log":
		app.config.LogDir = fallbackPath
	case "cache":
		app.config.CacheDir = fallbackPath
	}

	app.logger.Info("Using fallback %s directory: %s", dirType, fallbackPath)
}
