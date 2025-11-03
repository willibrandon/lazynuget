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
	startTime time.Time
	logger    logging.Logger
	platform  platform.Platform
	gui       any
	ctx       context.Context
	config    *config.AppConfig
	lifecycle *lifecycle.Manager
	cancel    context.CancelFunc
	version   VersionInfo
	phase     string
	runMode   platform.RunMode
	guiOnce   sync.Once
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
				if err := app.lifecycle.SetState(lifecycle.StateFailed); err != nil && app.logger != nil {
					app.logger.Error("Failed to set lifecycle state to failed during panic recovery: %v", err)
				}
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
		if setErr := app.lifecycle.SetState(lifecycle.StateFailed); setErr != nil {
			return fmt.Errorf("configuration loading failed: %w (state transition error: %w)", err, setErr)
		}
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
	app.platform = platform.New()

	// Phase: Determine run mode (interactive vs non-interactive)
	app.phase = "runmode"
	app.runMode = platform.DetermineRunMode(app.config.NonInteractive)
	app.logger.Info("Run mode determined: %s", app.runMode)

	// Phase: Dotnet CLI validation (synchronous, non-failing)
	app.phase = "dotnet-validation"
	// Validate dotnet CLI but don't fail startup if missing
	if err := platform.ValidateDotnetCLI(); err != nil {
		app.logger.Warn("Dotnet CLI validation warning: %v", err)
		// Don't fail startup - just warn the user
	} else {
		app.logger.Debug("Dotnet CLI validated successfully")
	}

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
func (app *App) GetGUI() any {
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
			if !os.IsNotExist(err) {
				app.logger.Warn("Cannot access %s directory %s: %v", dir.name, dir.path, err)
				continue
			}
			// Try to create the directory (owner-only permissions for security)
			if err := os.MkdirAll(dir.path, 0o700); err != nil {
				app.logger.Warn("Cannot create %s directory %s: %v\nFalling back to temp directory", dir.name, dir.path, err)
				app.useTempDirectoryFallback(dir.name)
				continue
			}
			app.logger.Debug("Created %s directory: %s", dir.name, dir.path)
		} else if !info.IsDir() {
			app.logger.Warn("%s path %s exists but is not a directory\nFalling back to temp directory", dir.name, dir.path)
			app.useTempDirectoryFallback(dir.name)
			continue
		}

		// Test write permissions by creating a temp file (owner-only for security)
		testFile := filepath.Join(dir.path, ".lazynuget-write-test")
		if err := os.WriteFile(testFile, []byte("test"), 0o600); err != nil {
			app.logger.Warn("Cannot write to %s directory %s: %v\nFalling back to temp directory", dir.name, dir.path, err)
			app.useTempDirectoryFallback(dir.name)
			continue
		}
		// Clean up test file
		if err := os.Remove(testFile); err != nil {
			app.logger.Debug("Failed to remove test file %s: %v (not critical)", testFile, err)
		}

		app.logger.Debug("%s directory verified: %s", dir.name, dir.path)
	}
}

// useTempDirectoryFallback updates config to use temp directory for the specified type
func (app *App) useTempDirectoryFallback(dirType string) {
	tempBase := os.TempDir()
	fallbackPath := filepath.Join(tempBase, "lazynuget", dirType)

	// Create the fallback directory (owner-only permissions for security)
	if err := os.MkdirAll(fallbackPath, 0o700); err != nil {
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
