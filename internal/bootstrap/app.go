package bootstrap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
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

	// lifecycle manages application state transitions (to be implemented in US2)
	lifecycle interface{}

	// gui is the Bubbletea TUI program (to be implemented in US4)
	gui interface{}

	// ctx is the root cancellation context
	ctx context.Context

	// cancel is the cancellation function for shutdown
	cancel context.CancelFunc

	// version contains build information
	version VersionInfo

	// startTime is when the application was created
	startTime time.Time

	// shutdownHandlers are called during graceful shutdown
	shutdownHandlers []func(context.Context) error

	// guiOnce ensures GUI is initialized only once (lazy initialization)
	guiOnce sync.Once

	// phase tracks the current initialization phase for error context
	phase string
}

// NewApp creates a new application instance with version information.
func NewApp(version, commit, date string) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		ctx:              ctx,
		cancel:           cancel,
		version:          VersionInfo{Version: version, Commit: commit, Date: date},
		startTime:        time.Now(),
		shutdownHandlers: make([]func(context.Context) error, 0),
		phase:            "uninitialized",
	}

	return app, nil
}

// Bootstrap initializes all application subsystems in the correct order.
// This method implements Layer 2 panic recovery with phase tracking.
func (app *App) Bootstrap() error {
	// Layer 2 panic recovery: catch panics and add phase context
	defer func() {
		if r := recover(); r != nil {
			if app.logger != nil {
				app.logger.Error("PANIC during bootstrap (phase: %s): %v", app.phase, r)
			}
			// Re-panic for main() to handle
			panic(r)
		}
	}()

	// Phase: Config loading
	app.phase = "config"
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	app.config = cfg

	// Phase: Logging setup
	app.phase = "logging"
	app.logger = logging.New(app.config.LogLevel, "")

	// Phase: Platform detection
	app.phase = "platform"
	app.platform = platform.New(app.config, app.logger)

	app.phase = "ready"
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
