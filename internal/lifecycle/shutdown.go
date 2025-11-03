package lifecycle

import (
	"context"
	"fmt"
	"runtime/debug"
	"sort"

	"github.com/willibrandon/lazynuget/internal/logging"
)

// Shutdown executes all registered shutdown handlers with timeout
func (m *Manager) Shutdown(ctx context.Context, logger logging.Logger) error {
	// Layer 5 panic recovery: Protect shutdown process
	defer func() {
		if r := recover(); r != nil {
			if logger != nil {
				logger.Error("PANIC during shutdown: %v\nStack: %s", r, debug.Stack())
			}
			// Don't re-panic during shutdown
		}
	}()

	// Transition to shutting down state
	if err := m.SetState(StateShuttingDown); err != nil {
		return fmt.Errorf("failed to transition to shutdown state: %w", err)
	}

	if logger != nil {
		logger.Info("Beginning graceful shutdown (timeout: %s)", m.shutdownTimeout)
	}

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, m.shutdownTimeout)
	defer cancel()

	// Sort handlers by priority (lower numbers first)
	handlers := m.getSortedHandlers()

	// Execute handlers sequentially
	var shutdownErrors []error
	for _, handler := range handlers {
		if logger != nil {
			logger.Debug("Running shutdown handler: %s (priority: %d)", handler.Name, handler.Priority)
		}

		// Wrap handler execution with panic recovery
		err := m.executeHandlerSafely(shutdownCtx, handler, logger)
		if err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("%s: %w", handler.Name, err))
			if logger != nil {
				logger.Warn("Shutdown handler failed: %s: %v", handler.Name, err)
			}
		}

		// Check if context expired
		if shutdownCtx.Err() != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("shutdown timeout exceeded"))
			if logger != nil {
				logger.Error("Shutdown timeout exceeded")
			}
			break
		}
	}

	// Transition to shutdown complete
	if err := m.SetState(StateShutdownComplete); err != nil {
		shutdownErrors = append(shutdownErrors, err)
	}

	if logger != nil {
		uptime := m.GetUptime()
		logger.Info("Shutdown complete (uptime: %s)", uptime)
	}

	// Return combined errors if any
	if len(shutdownErrors) > 0 {
		return fmt.Errorf("shutdown completed with %d errors: %v", len(shutdownErrors), shutdownErrors)
	}

	return nil
}

// executeHandlerSafely runs a shutdown handler with panic recovery
func (m *Manager) executeHandlerSafely(ctx context.Context, handler ShutdownHandler, logger logging.Logger) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in shutdown handler: %v", r)
			if logger != nil {
				logger.Error("PANIC in shutdown handler %s: %v\nStack: %s", handler.Name, r, debug.Stack())
			}
		}
	}()

	return handler.Handler(ctx)
}

// getSortedHandlers returns handlers sorted by priority
func (m *Manager) getSortedHandlers() []ShutdownHandler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create a copy to avoid holding lock during execution
	handlers := make([]ShutdownHandler, len(m.shutdownHandlers))
	copy(handlers, m.shutdownHandlers)

	// Sort by priority (lower numbers first)
	sort.Slice(handlers, func(i, j int) bool {
		return handlers[i].Priority < handlers[j].Priority
	})

	return handlers
}
