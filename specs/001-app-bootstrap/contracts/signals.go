// Package contracts defines the core interfaces for LazyNuGet's bootstrap system.
package contracts

import (
	"context"
	"os"
)

// SignalHandler manages OS signal processing for graceful shutdown.
// It abstracts platform-specific signal handling and integrates with the
// application's context-based cancellation system.
//
// Responsibilities:
//   - Register interest in specific OS signals (SIGINT, SIGTERM)
//   - Notify application when signals are received
//   - Support force-quit on second signal (restore default behavior)
//   - Work consistently across Windows, macOS, Linux
//
// Thread Safety: All methods are safe for concurrent use.
//
// Constitutional Alignment:
//   - Principle III (Safety): Graceful shutdown prevents data loss
//   - Principle IV (Cross-Platform): Identical behavior on all platforms
//
// Example Usage:
//
//	handler := signal.New(logger)
//	ctx := handler.Register(os.Interrupt, syscall.SIGTERM)
//
//	// Application runs...
//	<-ctx.Done()  // Blocks until signal received
//	log.Info("Shutdown signal received")
//
//	handler.Shutdown()  // Cleanup signal handler
//
// Corresponds to: FR-008 (signal handling), research.md section 2
type SignalHandler interface {
	// Register configures the handler to listen for specific OS signals.
	//
	// This method:
	//   - Creates a cancellable context derived from context.Background()
	//   - Configures signal.NotifyContext for the specified signals
	//   - Returns the context that will be canceled when any signal arrives
	//
	// Parameters:
	//   - signals: OS signals to monitor (typically os.Interrupt, syscall.SIGTERM)
	//
	// Returns:
	//   - Context that is canceled when first signal arrives
	//
	// Behavior:
	//   - First signal: Cancels context for graceful shutdown
	//   - Second signal: Calls Shutdown() to restore default behavior (force quit)
	//
	// Platform Notes:
	//   - Windows: os.Interrupt on Ctrl+C, syscall.SIGTERM on console close
	//   - macOS/Linux: Full POSIX signal support
	//
	// Example:
	//
	//	ctx := handler.Register(os.Interrupt, syscall.SIGTERM)
	//	go func() {
	//	    <-ctx.Done()
	//	    log.Info("Graceful shutdown initiated")
	//	    // Cleanup...
	//	}()
	//
	// Corresponds to: FR-008, research.md signal.NotifyContext pattern
	Register(signals ...os.Signal) context.Context

	// Wait blocks until a registered signal is received.
	//
	// This method:
	//   - Blocks until signal arrives or context is canceled
	//   - Returns the signal that triggered the wakeup
	//   - Returns nil if context was canceled without signal
	//
	// Returns:
	//   - The received OS signal, or nil if context canceled
	//
	// Usage:
	//
	//	sig := handler.Wait()
	//	if sig != nil {
	//	    log.Infof("Received signal: %v", sig)
	//	}
	//
	// Note: Most applications should use the context returned by Register()
	// instead of calling Wait() directly, as it integrates better with
	// context-based cancellation patterns.
	Wait() os.Signal

	// Shutdown stops signal monitoring and restores default signal behavior.
	//
	// This method:
	//   - Stops forwarding signals to the application
	//   - Restores OS default handlers for all registered signals
	//   - Allows forced termination on next signal
	//
	// Call this when:
	//   - Graceful shutdown completes successfully
	//   - Shutdown timeout is exceeded (allow force quit)
	//   - User sends second signal during shutdown
	//
	// After calling Shutdown(), next signal will use OS default behavior
	// (typically immediate termination without cleanup).
	//
	// Example shutdown sequence:
	//
	//	<-signalCtx.Done()  // First signal
	//	go func() {
	//	    time.Sleep(3 * time.Second)
	//	    handler.Shutdown()  // Allow force quit after timeout
	//	}()
	//	// Perform cleanup...
	//
	// Corresponds to: FR-008 (signal handling with force-quit support)
	Shutdown()
}

// SignalHandlerConfig contains configuration for signal handling behavior.
type SignalHandlerConfig struct {
	ShutdownGracePeriod     os.Signal
	ForceQuitOnSecondSignal bool
}

// DefaultSignalHandlerConfig returns a SignalHandlerConfig with sensible defaults.
func DefaultSignalHandlerConfig() SignalHandlerConfig {
	return SignalHandlerConfig{
		ForceQuitOnSecondSignal: true,
		ShutdownGracePeriod:     nil, // Will be set to shutdownTimeout duration
	}
}

// SignalNotification contains details about a received signal.
// Useful for logging and metrics.
type SignalNotification struct {
	// Signal is the OS signal that was received.
	Signal os.Signal

	// ReceivedAt is the timestamp when the signal arrived.
	ReceivedAt string // time.Time as RFC3339 string

	// IsSecondSignal indicates if this is the second signal during shutdown.
	// When true, this triggers force-quit behavior.
	IsSecondSignal bool
}

// SignalObserver receives notifications when signals are received.
// Useful for logging, metrics, and coordinating shutdown across systems.
//
// All methods are called synchronously when signals arrive and MUST:
//   - Return immediately (<1ms)
//   - Not block or perform I/O
//   - Not panic (panics will disrupt signal handling)
type SignalObserver interface {
	// OnSignalReceived is called when a registered signal arrives.
	// Parameters:
	//   - notification: Details about the signal
	OnSignalReceived(notification SignalNotification)
}

// Common signal handling patterns from research.md section 2

// IntegrationExample demonstrates the recommended integration pattern combining
// signal handling with errgroup for coordinated shutdown.
//
// This pattern is documented in research.md section 2 and shows:
//   - signal.NotifyContext for cross-platform signals
//   - errgroup.WithContext for coordinated goroutines
//   - Context cancellation for graceful shutdown
//   - Force-quit support via stop() function
//
// Usage:
//
//	func main() {
//	    ctx, stop := signal.NotifyContext(context.Background(),
//	        os.Interrupt, syscall.SIGTERM)
//	    defer stop()
//
//	    g, gCtx := errgroup.WithContext(ctx)
//
//	    g.Go(func() error { return runApplication(gCtx) })
//
//	    g.Go(func() error {
//	        <-gCtx.Done()
//	        stop() // Allows second Ctrl+C to force quit
//
//	        shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
//	        defer cancel()
//	        return performGracefulShutdown(shutdownCtx)
//	    })
//
//	    if err := g.Wait(); err != nil && err != context.Canceled {
//	        os.Exit(1)
//	    }
//	}
//
// This example is provided as documentation of the recommended pattern.
// The actual implementation will be in internal/bootstrap/signals.go.
type IntegrationExample struct{}

// PlatformSignals defines which signals are supported on each platform.
// This helps with cross-platform testing and documentation.
//
// Note: Always use os.Interrupt for portability (works on all platforms).
// Use syscall.SIGTERM only when targeting Unix-like systems specifically.
var PlatformSignals = struct {
	// Windows supports limited signal set
	Windows []os.Signal

	// MacOS supports full POSIX signal set
	MacOS []os.Signal

	// Linux supports full POSIX signal set
	Linux []os.Signal

	// Common signals that work on all platforms
	Portable []os.Signal
}{
	Windows: []os.Signal{
		os.Interrupt, // Ctrl+C
		// syscall.SIGTERM is supported but behavior varies
	},
	MacOS: []os.Signal{
		os.Interrupt, // Ctrl+C (SIGINT)
		os.Kill,      // SIGKILL (cannot be caught)
		// syscall.SIGTERM, SIGHUP, etc. available via syscall package
	},
	Linux: []os.Signal{
		os.Interrupt, // Ctrl+C (SIGINT)
		os.Kill,      // SIGKILL (cannot be caught)
		// syscall.SIGTERM, SIGHUP, etc. available via syscall package
	},
	Portable: []os.Signal{
		os.Interrupt, // Safe on all platforms
	},
}

// BubbleteatIntegration documents how to integrate signal handling with
// the Bubbletea TUI framework.
//
// From research.md section 3, the recommended approach is:
//   - Bootstrap owns signal handling (not Bubbletea)
//   - Use tea.WithoutSignalHandler() to disable Bubbletea's handling
//   - Use tea.WithContext() to pass cancellation context
//   - Send tea.Quit() message when shutdown signal received
//
// Example:
//
//	type Application struct {
//	    ctx     context.Context
//	    cancel  context.CancelFunc
//	    program *tea.Program
//	}
//
//	func (app *Application) Run() error {
//	    app.ctx, app.cancel = context.WithCancel(context.Background())
//	    defer app.cancel()
//
//	    // Signal handling before Bubbletea
//	    sigChan := make(chan os.Signal, 1)
//	    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
//
//	    go func() {
//	        <-sigChan
//	        if app.program != nil {
//	            app.program.Send(tea.Quit())
//	        }
//	    }()
//
//	    model := NewAppModel(app.config, app.logger)
//	    app.program = tea.NewProgram(
//	        model,
//	        tea.WithContext(app.ctx),
//	        tea.WithoutSignalHandler(),  // Bootstrap handles signals
//	        tea.WithAltScreen(),
//	    )
//
//	    finalModel, err := app.program.Run()
//	    app.cleanup(finalModel)
//	    return err
//	}
//
// This pattern ensures:
//   - Consistent signal handling across TUI and non-TUI modes
//   - Proper coordination between bootstrap and GUI layers
//   - Clean terminal restoration on shutdown
type BubbleteatIntegration struct{}

// NonInteractiveMode documents signal handling in non-interactive mode.
// When running without a TTY (CI, tests, automation), signal handling
// behavior should be identical to interactive mode.
//
// From research.md section 4, non-interactive mode is detected via:
//   - golang.org/x/term.IsTerminal() for TTY detection
//   - Environment variables (CI, NO_COLOR, TERM=dumb)
//   - Explicit --non-interactive flag
//
// Signal handling remains the same:
//   - First SIGINT/SIGTERM: Graceful shutdown
//   - Second signal: Force quit
//   - No difference in behavior based on TTY status
//
// Example test:
//
//	func TestNonInteractiveShutdown(t *testing.T) {
//	    // Simulate CI environment
//	    os.Setenv("CI", "true")
//
//	    app := NewApp("1.0.0", "abc123", "2025-11-02")
//	    ctx := app.SignalHandler().Register(os.Interrupt)
//
//	    go func() {
//	        time.Sleep(100 * time.Millisecond)
//	        syscall.Kill(syscall.Getpid(), syscall.SIGINT)
//	    }()
//
//	    <-ctx.Done()
//	    // Verify graceful shutdown occurred
//	}
type NonInteractiveMode struct{}
