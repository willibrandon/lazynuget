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
	startTime    time.Time
	configLoader config.ConfigLoader
	platform     platform.PlatformInfo
	pathResolver platform.PathResolver
	gui          any
	ctx          context.Context
	watcher      config.ConfigWatcher
	logger       logging.Logger
	config       *config.Config
	cancel       context.CancelFunc
	lifecycle    *lifecycle.Manager
	version      VersionInfo
	configPath   string
	phase        string
	runMode      platform.RunMode
	configMu     sync.RWMutex
	guiOnce      sync.Once
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

	// Create config loader
	loader := config.NewLoader()

	// Prepare load options from flags
	loadOpts := config.LoadOptions{
		EnvVarPrefix: "LAZYNUGET_",
		StrictMode:   false,
		Logger:       nil, // Will set up logger after config is loaded
	}

	if flags != nil {
		loadOpts.ConfigFilePath = flags.ConfigPath
		loadOpts.CLIFlags = config.CLIFlags{
			LogLevel:       flags.LogLevel,
			NonInteractive: flags.NonInteractive,
		}
	}

	cfg, err := loader.Load(app.ctx, loadOpts)
	if err != nil {
		if setErr := app.lifecycle.SetState(lifecycle.StateFailed); setErr != nil {
			return fmt.Errorf("configuration loading failed: %w (state transition error: %w)", err, setErr)
		}
		return fmt.Errorf("configuration loading failed: %w", err)
	}
	app.config = cfg
	app.configLoader = loader
	app.configPath = loadOpts.ConfigFilePath

	// Phase: Logging setup
	app.phase = "logging"
	// For now, log to stdout only (file logging can be added later)
	app.logger = logging.New(app.config.LogLevel, "")

	// Phase: Directory permission checking
	app.phase = "directory-permissions"
	app.checkDirectoryPermissions()

	// Phase: Platform detection
	app.phase = "platform"
	platformInfo, err := platform.New()
	if err != nil {
		if setErr := app.lifecycle.SetState(lifecycle.StateFailed); setErr != nil {
			return fmt.Errorf("platform detection failed: %w (state transition error: %w)", err, setErr)
		}
		return fmt.Errorf("platform detection failed: %w", err)
	}
	app.platform = platformInfo

	// Log detected platform information
	app.logger.Debug("Platform detected: OS=%s, Arch=%s, Version=%s",
		platformInfo.OS(), platformInfo.Arch(), platformInfo.Version())

	// Create path resolver for platform-specific path operations
	pathResolver, err := platform.NewPathResolver(platformInfo)
	if err != nil {
		if setErr := app.lifecycle.SetState(lifecycle.StateFailed); setErr != nil {
			return fmt.Errorf("path resolver creation failed: %w (state transition error: %w)", err, setErr)
		}
		return fmt.Errorf("path resolver creation failed: %w", err)
	}
	app.pathResolver = pathResolver

	// Log platform paths
	configDir, configErr := pathResolver.ConfigDir()
	cacheDir, cacheErr := pathResolver.CacheDir()
	if configErr == nil && cacheErr == nil {
		app.logger.Debug("Platform paths: Config=%s, Cache=%s", configDir, cacheDir)
	} else {
		app.logger.Warn("Failed to retrieve platform paths: config=%v, cache=%v", configErr, cacheErr)
	}

	// Detect and log terminal capabilities (T069)
	termCaps := platform.NewTerminalCapabilities()
	app.logger.Debug("Terminal capabilities: ColorDepth=%s, Unicode=%v, TTY=%v",
		termCaps.GetColorDepth(), termCaps.SupportsUnicode(), termCaps.IsTTY())

	// Check terminal dimensions and warn if below minimum (T070, FR-015)
	width, height, err := termCaps.GetSize()
	if err == nil {
		const (
			MinWidth  = 40
			MinHeight = 10
		)
		if width < MinWidth || height < MinHeight {
			app.logger.Warn("Terminal dimensions %dx%d are below recommended minimum %dx%d. "+
				"TUI may not display correctly. Dimensions have been clamped to safe values.",
				width, height, MinWidth, MinHeight)
		}
	}

	// Phase: Determine run mode (interactive vs non-interactive)
	app.phase = "runmode"
	nonInteractive := false
	if flags != nil {
		nonInteractive = flags.NonInteractive
	}
	app.runMode = platform.DetermineRunMode(nonInteractive)
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

	// Phase: Hot-reload watcher setup (if enabled)
	app.phase = "hot-reload"
	if app.config.HotReload && app.configPath != "" {
		app.logger.Info("Hot-reload enabled, starting config file watcher")

		watcher, err := config.NewConfigWatcher(config.WatchOptions{
			ConfigFilePath: app.configPath,
			LoadOptions:    loadOpts,
			OnReload: func(newCfg *config.Config) {
				app.configMu.Lock()
				app.config = newCfg
				app.configMu.Unlock()
				app.logger.Info("Configuration reloaded successfully")
			},
			OnError: func(err error) {
				app.logger.Error("Configuration reload failed: %v", err)
			},
			OnFileDeleted: func() {
				app.logger.Warn("Configuration file deleted, using previous configuration")
			},
		}, loader)

		if err != nil {
			app.logger.Warn("Failed to start config file watcher: %v", err)
		} else {
			app.watcher = watcher

			// Start watching in background
			eventCh, errCh, err := watcher.Watch(app.ctx)
			if err != nil {
				app.logger.Warn("Failed to start watching config file: %v", err)
			} else {
				// Consume events in background goroutine
				go func() {
					for {
						select {
						case <-app.ctx.Done():
							return
						case event := <-eventCh:
							app.logger.Debug("Config change event: type=%s, error=%v", event.Type, event.Error)
						case err := <-errCh:
							app.logger.Error("Config watcher error: %v", err)
						}
					}
				}()

				// Register shutdown handler to stop watcher
				app.RegisterShutdownHandler("config-watcher", 100, func(_ context.Context) error {
					if app.watcher != nil {
						app.logger.Debug("Stopping config file watcher")
						return app.watcher.Stop()
					}
					return nil
				})

				app.logger.Debug("Config file watcher started successfully")
			}
		}
	} else if app.config.HotReload && app.configPath == "" {
		app.logger.Debug("Hot-reload enabled but no config file path available (using defaults)")
	}

	// Register logger cleanup handler (runs last, after all other shutdown handlers)
	app.RegisterShutdownHandler("logger", 999, func(_ context.Context) error {
		app.logger.Debug("Closing logger")
		return app.logger.Close()
	})

	// Transition to running state
	app.phase = "ready"
	if err := app.lifecycle.SetState(lifecycle.StateRunning); err != nil {
		return fmt.Errorf("failed to enter running state: %w", err)
	}

	app.logger.Info("Bootstrap complete, application is running")
	return nil
}

// GetConfig returns the application configuration.
// Thread-safe: uses RLock to allow concurrent reads while hot-reload updates happen.
func (app *App) GetConfig() *config.Config {
	app.configMu.RLock()
	defer app.configMu.RUnlock()
	return app.config
}

// GetLogger returns the application logger.
func (app *App) GetLogger() logging.Logger {
	return app.logger
}

// GetPlatform returns the platform utilities.
func (app *App) GetPlatform() platform.PlatformInfo {
	return app.platform
}

// GetPathResolver returns the platform path resolver.
func (app *App) GetPathResolver() platform.PathResolver {
	return app.pathResolver
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

	// In non-interactive mode, exit after a brief delay to allow async operations to complete
	if !app.runMode.IsInteractive() {
		go func() {
			time.Sleep(200 * time.Millisecond)
			app.logger.Info("Non-interactive mode: triggering automatic shutdown")
			app.cancel() // Trigger shutdown
		}()
	}

	// Block until context is cancelled (either by signal or by auto-shutdown in non-interactive mode)
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
		{"log", app.config.LogDir},
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
	case "log":
		app.config.LogDir = fallbackPath
	}

	app.logger.Info("Using fallback %s directory: %s", dirType, fallbackPath)
}
