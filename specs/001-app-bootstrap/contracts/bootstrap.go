// Package contracts defines the core interfaces for LazyNuGet's bootstrap system.
// These interfaces establish the contract between components without coupling to implementations.
//
// Feature: Application Bootstrap and Lifecycle Management (001-app-bootstrap)
// Phase: 1 - Design & Contracts
package contracts

import (
	"context"
	"time"
)

// Bootstrapper defines the contract for application initialization and startup.
// Implementations coordinate the complete bootstrap sequence from binary launch
// to fully operational TUI or non-interactive mode.
//
// Responsibilities:
//   - Parse command-line arguments and environment variables
//   - Initialize all application subsystems in correct dependency order
//   - Validate configuration and external dependencies (dotnet CLI)
//   - Launch interactive TUI or non-interactive mode based on environment
//   - Coordinate graceful shutdown on signals or user command
//
// Constitutional Alignment:
//   - Principle II (Simplicity): Zero-config operation with sensible defaults
//   - Principle V (Performance): <200ms p95 startup target
//   - Principle VII (Testability): Non-interactive mode for automated testing
//
// Example Usage:
//
//	bootstrapper := bootstrap.New(version, commit, date)
//	if err := bootstrapper.Initialize(); err != nil {
//	    log.Fatalf("Initialization failed: %v", err)
//	}
//	if err := bootstrapper.Run(); err != nil {
//	    log.Fatalf("Runtime error: %v", err)
//	}
type Bootstrapper interface {
	// Initialize prepares the application for execution.
	//
	// This method:
	//   - Parses command-line flags (--version, --help, --config, etc.)
	//   - Loads and validates configuration from all sources
	//   - Initializes logger (first, to capture all events)
	//   - Initializes platform utilities
	//   - Validates external dependencies (dotnet CLI availability)
	//   - Registers signal handlers (SIGINT, SIGTERM)
	//
	// Returns error if:
	//   - Configuration validation fails (invalid values, conflicting settings)
	//   - Required external dependencies missing (dotnet CLI not found)
	//   - File system operations fail (cannot create config/log directories)
	//
	// Performance: MUST complete in <100ms to meet <200ms total startup budget.
	// This excludes GUI initialization which is deferred to Run().
	//
	// Corresponds to: FR-004, FR-005, FR-006, FR-007, FR-015
	Initialize() error

	// Run executes the main application loop.
	//
	// This method:
	//   - Detects interactive vs non-interactive mode (TTY detection)
	//   - Initializes GUI (lazy, only if interactive mode)
	//   - Starts the Bubbletea event loop or non-interactive worker
	//   - Blocks until shutdown signal received or user quits
	//   - Returns when application is ready to exit
	//
	// Returns error if:
	//   - GUI initialization fails in interactive mode
	//   - Runtime panic occurs (after recovery and logging)
	//   - Background operations fail critically
	//
	// Shutdown Behavior:
	//   - First SIGINT/SIGTERM: Graceful shutdown (cleanup within timeout)
	//   - Second signal: Force quit (restore default signal behavior)
	//
	// Performance: GUI initialization (if needed) should complete in <100ms
	// to meet total <200ms startup budget.
	//
	// Corresponds to: FR-008, FR-011, FR-012, FR-018
	Run() error

	// Shutdown initiates graceful application termination.
	//
	// This method:
	//   - Cancels the application's root context
	//   - Signals all background goroutines to stop
	//   - Waits for GUI cleanup (if running)
	//   - Executes shutdown handlers in reverse initialization order
	//   - Enforces timeout to prevent indefinite hangs
	//
	// The ctx parameter provides overall timeout control. If ctx is canceled
	// or times out before shutdown completes, cleanup is abandoned and the
	// method returns immediately.
	//
	// Returns error if:
	//   - Shutdown handlers fail
	//   - Timeout is exceeded (from FR-009: <3 seconds)
	//   - Context is canceled before completion
	//
	// Guarantees:
	//   - Resource cleanup attempted (files closed, buffers flushed)
	//   - Exit code set appropriately (0=success, 1=user error, 2=system error)
	//   - No zombie goroutines left running
	//
	// Corresponds to: FR-009, FR-010, FR-016
	Shutdown(ctx context.Context) error
}

// VersionInfo contains build and release information displayed to users.
// This data is typically injected at build time via -ldflags.
//
// Example injection:
//
//	go build -ldflags="-X main.version=1.0.0 -X main.commit=abc123 -X main.date=2025-11-02"
//
// Corresponds to: FR-002
type VersionInfo struct {
	// Version is the semantic version string (e.g., "1.0.0", "1.2.3-beta").
	Version string

	// Commit is the Git commit SHA (short or full) identifying the exact source.
	Commit string

	// Date is the build timestamp in RFC3339 format (e.g., "2025-11-02T10:30:00Z").
	Date string
}

// String formats version info for display to users.
// Format: "LazyNuGet version {Version} ({Commit}) built on {Date}"
func (v VersionInfo) String() string {
	return "LazyNuGet version " + v.Version + " (" + v.Commit + ") built on " + v.Date
}

// ConfigSource represents a source of configuration data with its precedence.
// Multiple sources are merged following the precedence order defined in FR-005:
// CLI flags (highest) > Environment > File > Defaults (lowest).
type ConfigSource int

const (
	// SourceDefault represents hardcoded default values (lowest precedence).
	SourceDefault ConfigSource = iota

	// SourceFile represents values loaded from config file (~/.config/lazynuget/config.yml).
	SourceFile

	// SourceEnvironment represents values from environment variables (LAZYNUGET_*).
	SourceEnvironment

	// SourceCLI represents values from command-line flags (highest precedence).
	SourceCLI
)

// String returns human-readable name for ConfigSource.
func (cs ConfigSource) String() string {
	switch cs {
	case SourceDefault:
		return "default"
	case SourceFile:
		return "file"
	case SourceEnvironment:
		return "environment"
	case SourceCLI:
		return "cli"
	default:
		return "unknown"
	}
}

// StartupPhase represents a stage in the bootstrap sequence.
// Useful for progress reporting and debugging slow startups.
type StartupPhase int

const (
	// PhaseInit represents initial setup (flags parsing, context creation).
	PhaseInit StartupPhase = iota

	// PhaseConfig represents configuration loading and validation.
	PhaseConfig

	// PhaseLogging represents logger initialization.
	PhaseLogging

	// PhasePlatform represents platform detection and path setup.
	PhasePlatform

	// PhaseDependencies represents external dependency validation (dotnet CLI).
	PhaseDependencies

	// PhaseSignals represents signal handler registration.
	PhaseSignals

	// PhaseGUI represents GUI initialization (lazy, only if interactive).
	PhaseGUI

	// PhaseReady represents fully operational state.
	PhaseReady
)

// String returns human-readable name for StartupPhase.
func (sp StartupPhase) String() string {
	switch sp {
	case PhaseInit:
		return "init"
	case PhaseConfig:
		return "config"
	case PhaseLogging:
		return "logging"
	case PhasePlatform:
		return "platform"
	case PhaseDependencies:
		return "dependencies"
	case PhaseSignals:
		return "signals"
	case PhaseGUI:
		return "gui"
	case PhaseReady:
		return "ready"
	default:
		return "unknown"
	}
}

// ShutdownPhase represents a stage in the graceful shutdown sequence.
// Shutdown proceeds in reverse order of initialization.
type ShutdownPhase int

const (
	// PhaseShutdownInit represents the start of shutdown sequence.
	PhaseShutdownInit ShutdownPhase = iota

	// PhaseShutdownGUI represents GUI cleanup (stop event loop, restore terminal).
	PhaseShutdownGUI

	// PhaseShutdownWorkers represents canceling and waiting for background goroutines.
	PhaseShutdownWorkers

	// PhaseShutdownPlatform represents platform resource cleanup.
	PhaseShutdownPlatform

	// PhaseShutdownConfig represents configuration finalization (flush writes).
	PhaseShutdownConfig

	// PhaseShutdownLogging represents logger cleanup (flush buffers, close files).
	PhaseShutdownLogging

	// PhaseShutdownComplete represents fully stopped state.
	PhaseShutdownComplete
)

// String returns human-readable name for ShutdownPhase.
func (sp ShutdownPhase) String() string {
	switch sp {
	case PhaseShutdownInit:
		return "init"
	case PhaseShutdownGUI:
		return "gui"
	case PhaseShutdownWorkers:
		return "workers"
	case PhaseShutdownPlatform:
		return "platform"
	case PhaseShutdownConfig:
		return "config"
	case PhaseShutdownLogging:
		return "logging"
	case PhaseShutdownComplete:
		return "complete"
	default:
		return "unknown"
	}
}

// ExitCode represents standard POSIX exit codes for the application.
// These codes communicate success or failure type to the parent process.
//
// Corresponds to: FR-010
type ExitCode int

const (
	// ExitSuccess indicates normal termination (0).
	ExitSuccess ExitCode = 0

	// ExitUserError indicates user-caused error: invalid flags, bad config,
	// missing required input (1).
	ExitUserError ExitCode = 1

	// ExitSystemError indicates internal error: panic, resource exhaustion,
	// external dependency failure (2).
	ExitSystemError ExitCode = 2
)

// StartupMetrics captures performance data for the bootstrap sequence.
// Used to validate SC-001 (<200ms p95 startup) and identify slow phases.
//
// Enable collection via DEBUG_STARTUP=1 environment variable.
type StartupMetrics struct {
	PhaseTimings  map[StartupPhase]time.Duration
	TotalDuration time.Duration
	LazyGUI       bool
}

// IsSlow returns true if total duration exceeds the 200ms target.
// Useful for detecting performance regressions.
func (sm StartupMetrics) IsSlow() bool {
	return sm.TotalDuration > 200*time.Millisecond
}

// SlowestPhase returns the phase with longest duration and its time.
// Useful for identifying optimization targets.
func (sm StartupMetrics) SlowestPhase() (StartupPhase, time.Duration) {
	var slowestPhase StartupPhase
	var slowestDuration time.Duration

	for phase, duration := range sm.PhaseTimings {
		if duration > slowestDuration {
			slowestPhase = phase
			slowestDuration = duration
		}
	}

	return slowestPhase, slowestDuration
}
