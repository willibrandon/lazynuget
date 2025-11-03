// Package contracts defines the core interfaces for LazyNuGet's bootstrap system.
package contracts

import (
	"context"
	"time"
)

// Lifecycle manages application state transitions and coordinates startup/shutdown.
// It ensures orderly initialization of components, tracks application state, and
// enforces clean resource cleanup during shutdown.
//
// Responsibilities:
//   - Manage state transitions (Uninitialized → Initializing → Running → ShuttingDown → Stopped)
//   - Track running background goroutines
//   - Enforce shutdown timeout to prevent hangs
//   - Provide state visibility to other components
//
// Thread Safety: All methods are safe for concurrent use.
//
// Constitutional Alignment:
//   - Principle III (Safety): Guaranteed cleanup within timeout bounds
//   - Principle V (Performance): Non-blocking operations, efficient state checks
//
// Example Usage:
//
//	lifecycle := NewLifecycle(3 * time.Second, logger)
//	if err := lifecycle.Start(ctx, appContext); err != nil {
//	    log.Fatal(err)
//	}
//	// Application runs...
//	if err := lifecycle.Shutdown(ctx); err != nil {
//	    log.Error(err)
//	}
type Lifecycle interface {
	// Start initializes the application and transitions to Running state.
	//
	// This method:
	//   - Transitions from Uninitialized to Initializing
	//   - Invokes app.Bootstrap() to initialize all subsystems
	//   - Registers signal handlers (SIGINT, SIGTERM)
	//   - Transitions to Running on success, Stopped on failure
	//
	// Parameters:
	//   - ctx: Cancellation context for the overall application
	//   - app: Application context to bootstrap
	//
	// Returns error if:
	//   - Invalid state transition (e.g., already Running)
	//   - Bootstrap fails (config invalid, dependency missing, etc.)
	//   - Signal handler registration fails
	//
	// Corresponds to: FR-007, FR-008, FR-015
	Start(ctx context.Context, app any) error

	// Stop initiates graceful shutdown and waits for completion.
	//
	// This method:
	//   - Transitions from Running to ShuttingDown
	//   - Closes internal shutdown channel to notify all listeners
	//   - Waits for all tracked goroutines to complete (via WaitGroup)
	//   - Executes registered shutdown handlers in order
	//   - Enforces shutdownTimeout to prevent indefinite hangs
	//   - Transitions to Stopped when complete or timeout exceeded
	//
	// Parameters:
	//   - ctx: Timeout context for shutdown operations
	//
	// Returns error if:
	//   - Invalid state transition (e.g., not Running)
	//   - Shutdown handlers fail
	//   - Timeout exceeded before completion
	//   - Context canceled before completion
	//
	// Behavior on timeout:
	//   - Logs warning about incomplete shutdown
	//   - Forces transition to Stopped
	//   - Returns timeout error
	//
	// Corresponds to: FR-009, FR-016
	Stop(ctx context.Context) error

	// State returns the current lifecycle state.
	//
	// Thread-safe read operation using RWMutex. Useful for:
	//   - Conditional logic based on application state
	//   - Debugging and logging
	//   - Health checks and monitoring
	//
	// Returns: One of StateUninitialized, StateInitializing, StateRunning,
	//          StateShuttingDown, or StateStopped.
	State() State

	// Go launches a goroutine with lifecycle tracking and panic recovery.
	//
	// This method:
	//   - Increments internal WaitGroup counter
	//   - Launches goroutine with panic recovery (Layer 3 from research.md)
	//   - Decrements WaitGroup counter when fn returns
	//   - Logs errors returned by fn
	//
	// Parameters:
	//   - fn: Function to execute in background goroutine
	//
	// Panic Handling:
	//   - Panics are caught, logged, and NOT propagated
	//   - Allows graceful degradation instead of application crash
	//   - Goroutine WaitGroup still decremented properly
	//
	// Usage:
	//
	//	lifecycle.Go(func() error {
	//	    // Background work here
	//	    return nil
	//	})
	//
	// Corresponds to: FR-016 (panic recovery)
	Go(fn func() error)

	// Wait blocks until shutdown is initiated.
	//
	// This method:
	//   - Blocks until Stop() is called or context canceled
	//   - Returns immediately if already ShuttingDown or Stopped
	//
	// Useful for main goroutine coordination:
	//
	//	lifecycle.Start(ctx, app)
	//	lifecycle.Wait(ctx)  // Block until shutdown signal
	//	lifecycle.Stop(shutdownCtx)
	//
	// Returns error if:
	//   - Context canceled before shutdown initiated
	Wait(ctx context.Context) error

	// RegisterShutdownHandler adds a cleanup function to be called during shutdown.
	//
	// Handlers are called in the ORDER REGISTERED during shutdown. This allows
	// proper cleanup sequencing (e.g., close connections before closing logger).
	//
	// Parameters:
	//   - handler: Function accepting timeout context, returning error
	//
	// Handler Contract:
	//   - MUST respect context cancellation/timeout
	//   - SHOULD complete quickly (<500ms typical)
	//   - SHOULD NOT panic (but panics will be recovered)
	//   - Errors are logged but do NOT stop other handlers
	//
	// Usage:
	//
	//	lifecycle.RegisterShutdownHandler(func(ctx context.Context) error {
	//	    return database.Close()
	//	})
	//
	// Corresponds to: FR-009 (resource cleanup)
	RegisterShutdownHandler(handler func(context.Context) error)
}

// State represents the current application lifecycle state.
// States transition in a strict order: Uninitialized → Initializing → Running
// → ShuttingDown → Stopped.
//
// State transitions are protected by mutex and are atomic.
// Corresponds to: Entity 3 (Lifecycle Manager) state machine from data-model.md
type State int

const (
	// StateUninitialized indicates the application has not started bootstrap.
	// Valid transitions: → Initializing (via Start)
	StateUninitialized State = iota

	// StateInitializing indicates bootstrap is in progress.
	// Valid transitions: → Running (success), → Stopped (failure)
	StateInitializing

	// StateRunning indicates the application is fully operational.
	// Valid transitions: → ShuttingDown (via Stop or signal)
	StateRunning

	// StateShuttingDown indicates graceful shutdown in progress.
	// Valid transitions: → Stopped (cleanup complete or timeout)
	StateShuttingDown

	// StateStopped indicates the application has fully stopped.
	// Valid transitions: None (terminal state)
	StateStopped
)

// String returns human-readable name for State.
func (s State) String() string {
	switch s {
	case StateUninitialized:
		return "uninitialized"
	case StateInitializing:
		return "initializing"
	case StateRunning:
		return "running"
	case StateShuttingDown:
		return "shutting_down"
	case StateStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

// IsTransitionValid checks if transitioning from state 'from' to state 'to' is allowed.
// Useful for validation before attempting state changes.
func IsTransitionValid(from, to State) bool {
	switch from {
	case StateUninitialized:
		return to == StateInitializing
	case StateInitializing:
		return to == StateRunning || to == StateStopped
	case StateRunning:
		return to == StateShuttingDown
	case StateShuttingDown:
		return to == StateStopped
	case StateStopped:
		return false // Terminal state
	default:
		return false
	}
}

// ShutdownHandler is a function called during graceful shutdown.
// Handlers receive a context with timeout and should:
//   - Complete cleanup within the context deadline
//   - Return error only for critical failures
//   - Not panic (panics are recovered but logged as errors)
//
// Example:
//
//	func cleanupDatabase(ctx context.Context) error {
//	    if err := db.Close(); err != nil {
//	        return fmt.Errorf("failed to close database: %w", err)
//	    }
//	    return nil
//	}
type ShutdownHandler func(ctx context.Context) error

// LifecycleConfig contains configuration for Lifecycle behavior.
// This allows customization of timeouts and limits.
type LifecycleConfig struct {
	// ShutdownTimeout is the maximum time allowed for graceful shutdown.
	// Default: 3 seconds (from FR-009)
	// Range: 1s to 10s (validated by VR-009 from data-model.md)
	ShutdownTimeout time.Duration

	// MaxConcurrentOps limits how many background operations can run simultaneously.
	// This prevents resource exhaustion from runaway goroutine spawning.
	// Default: 4 (from data-model.md)
	// Range: 1 to 16 (validated by VR-010 from data-model.md)
	MaxConcurrentOps int

	// EnableMetrics enables lifecycle timing instrumentation.
	// When true, tracks duration of each lifecycle phase for performance analysis.
	// Typically enabled via DEBUG_LIFECYCLE=1 environment variable.
	EnableMetrics bool
}

// DefaultLifecycleConfig returns a LifecycleConfig with sensible defaults.
func DefaultLifecycleConfig() LifecycleConfig {
	return LifecycleConfig{
		ShutdownTimeout:  3 * time.Second,
		MaxConcurrentOps: 4,
		EnableMetrics:    false,
	}
}

// Validate checks if LifecycleConfig values are within acceptable ranges.
// Returns error describing the first validation failure found.
func (cfg LifecycleConfig) Validate() error {
	if cfg.ShutdownTimeout < 1*time.Second || cfg.ShutdownTimeout > 10*time.Second {
		return ErrInvalidShutdownTimeout
	}
	if cfg.MaxConcurrentOps < 1 || cfg.MaxConcurrentOps > 16 {
		return ErrInvalidMaxOps
	}
	return nil
}

// Common lifecycle errors.
var (
	// ErrInvalidState indicates an invalid state transition was attempted.
	ErrInvalidState = newLifecycleError("invalid state transition")

	// ErrAlreadyStarted indicates Start was called when already Running.
	ErrAlreadyStarted = newLifecycleError("lifecycle already started")

	// ErrNotRunning indicates Stop was called before Start completed.
	ErrNotRunning = newLifecycleError("lifecycle not running")

	// ErrShutdownTimeout indicates graceful shutdown exceeded the timeout.
	ErrShutdownTimeout = newLifecycleError("shutdown timeout exceeded")

	// ErrInvalidShutdownTimeout indicates ShutdownTimeout is out of range (1s-10s).
	ErrInvalidShutdownTimeout = newLifecycleError("shutdown timeout must be 1s-10s")

	// ErrInvalidMaxOps indicates MaxConcurrentOps is out of range (1-16).
	ErrInvalidMaxOps = newLifecycleError("max concurrent ops must be 1-16")
)

// LifecycleError is an error type for lifecycle-specific failures.
type LifecycleError struct {
	msg string
}

func newLifecycleError(msg string) *LifecycleError {
	return &LifecycleError{msg: msg}
}

func (e *LifecycleError) Error() string {
	return "lifecycle: " + e.msg
}

// LifecycleMetrics captures timing data for lifecycle operations.
// Collected when LifecycleConfig.EnableMetrics is true.
type LifecycleMetrics struct {
	// StartDuration is time spent in Start() from call to Running state.
	StartDuration time.Duration

	// RunDuration is time spent in Running state.
	RunDuration time.Duration

	// StopDuration is time spent in Stop() from call to Stopped state.
	StopDuration time.Duration

	// GoroutinesPeakCount is the maximum number of tracked goroutines at once.
	GoroutinesPeakCount int

	// ShutdownHandlerCount is the number of registered shutdown handlers.
	ShutdownHandlerCount int

	// ShutdownHandlerErrors is the number of handlers that returned errors.
	ShutdownHandlerErrors int
}

// LifecycleObserver receives notifications about lifecycle state changes.
// Useful for monitoring, health checks, and coordinating dependent systems.
//
// All methods are called synchronously during state transitions and MUST:
//   - Return quickly (<10ms)
//   - Not block or perform I/O
//   - Not panic (panics will crash the state machine)
type LifecycleObserver interface {
	// OnStateChange is called whenever the lifecycle state transitions.
	// Parameters:
	//   - from: Previous state
	//   - to: New state
	OnStateChange(from, to State)
}
