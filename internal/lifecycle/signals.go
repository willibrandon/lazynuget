package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/willibrandon/lazynuget/internal/logging"
)

// SignalHandler manages OS signal handling for graceful shutdown
type SignalHandler struct {
	manager *Manager
	logger  logging.Logger
	signals []os.Signal
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(manager *Manager, logger logging.Logger) *SignalHandler {
	return &SignalHandler{
		manager: manager,
		logger:  logger,
		signals: []os.Signal{
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination request
		},
	}
}

// WaitForShutdownSignal blocks until a shutdown signal is received
// Returns a context that will be cancelled when shutdown is requested
func (sh *SignalHandler) WaitForShutdownSignal(parentCtx context.Context) context.Context {
	// Create cancellable context from parent
	ctx, cancel := context.WithCancel(parentCtx)

	// Create signal channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sh.signals...)

	// Start goroutine to wait for signals
	go func() {
		// Layer 4 panic recovery: Protect goroutines
		defer func() {
			if r := recover(); r != nil {
				if sh.logger != nil {
					sh.logger.Error("PANIC in signal handler goroutine: %v", r)
				}
			}
		}()

		select {
		case sig := <-sigChan:
			if sh.logger != nil {
				sh.logger.Info("Received signal: %s, initiating shutdown", sig)
			}
			cancel()
		case <-parentCtx.Done():
			// Parent context cancelled
			cancel()
		}

		// Stop receiving signals
		signal.Stop(sigChan)
		close(sigChan)
	}()

	return ctx
}
