package integration

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/bootstrap"
	"github.com/willibrandon/lazynuget/internal/lifecycle"
)

func TestGracefulShutdown(t *testing.T) {
	// Create app instance
	app, err := bootstrap.NewApp("test", "test", "test")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Bootstrap
	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Failed to bootstrap: %v", err)
	}

	// Register shutdown handlers
	handler1Called := false
	handler2Called := false

	app.RegisterShutdownHandler("test-handler-1", 10, func(ctx context.Context) error {
		handler1Called = true
		return nil
	})

	app.RegisterShutdownHandler("test-handler-2", 20, func(ctx context.Context) error {
		handler2Called = true
		return nil
	})

	// Perform shutdown
	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify handlers were called
	if !handler1Called {
		t.Error("Expected handler 1 to be called")
	}

	if !handler2Called {
		t.Error("Expected handler 2 to be called")
	}
}

func TestShutdownHandlerPriority(t *testing.T) {
	// Create app instance
	app, err := bootstrap.NewApp("test", "test", "test")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Bootstrap
	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Failed to bootstrap: %v", err)
	}

	// Track execution order
	var executionOrder []string

	// Register handlers in reverse priority order
	app.RegisterShutdownHandler("high-priority", 10, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "high")
		return nil
	})

	app.RegisterShutdownHandler("low-priority", 30, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "low")
		return nil
	})

	app.RegisterShutdownHandler("medium-priority", 20, func(ctx context.Context) error {
		executionOrder = append(executionOrder, "medium")
		return nil
	})

	// Perform shutdown
	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Verify execution order (should be high -> medium -> low)
	if len(executionOrder) != 3 {
		t.Fatalf("Expected 3 handlers to execute, got %d", len(executionOrder))
	}

	if executionOrder[0] != "high" || executionOrder[1] != "medium" || executionOrder[2] != "low" {
		t.Errorf("Expected order [high, medium, low], got %v", executionOrder)
	}
}

func TestShutdownWithTimeout(t *testing.T) {
	// Create app instance
	app, err := bootstrap.NewApp("test", "test", "test")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Bootstrap
	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Failed to bootstrap: %v", err)
	}

	// Register a handler that takes too long (simulating hung shutdown)
	// The handler will block for 100 seconds, but should be cancelled by the
	// shutdown timeout (30 seconds from lifecycle manager)
	app.RegisterShutdownHandler("slow-handler", 10, func(ctx context.Context) error {
		select {
		case <-time.After(100 * time.Second):
			return nil
		case <-ctx.Done():
			// Expected: context timeout after 30 seconds
			return ctx.Err()
		}
	})

	// Perform shutdown (should timeout after 30 seconds per lifecycle manager config)
	start := time.Now()
	err = app.Shutdown()
	elapsed := time.Since(start)

	// Shutdown should complete around the 30 second timeout (±2 seconds tolerance)
	if elapsed < 28*time.Second || elapsed > 32*time.Second {
		t.Errorf("Expected shutdown to take ~30 seconds (timeout), but took: %v", elapsed)
	}

	// Error is expected due to timeout
	if err == nil {
		t.Error("Expected shutdown to return error due to timeout")
	}
}

func TestShutdownStateTransitions(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "../../lazynuget-test", "../../cmd/lazynuget/main.go")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer exec.Command("rm", "../../lazynuget-test").Run()

	// This test verifies that the application goes through proper state transitions
	// We do this by testing in-process
	app, err := bootstrap.NewApp("test", "test", "test")
	if err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Bootstrap should transition to Running state
	if err := app.Bootstrap(nil); err != nil {
		t.Fatalf("Failed to bootstrap: %v", err)
	}

	// Shutdown should transition to ShutdownComplete state
	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Note: We can't directly check the state from outside the package
	// This test validates that the shutdown process completes without panics
}

func TestLifecycleStateValidation(t *testing.T) {
	// Test that lifecycle state machine validates transitions
	mgr := lifecycle.NewManager(5 * time.Second)

	// Valid transition: Uninitialized -> Initializing
	if err := mgr.SetState(lifecycle.StateInitializing); err != nil {
		t.Errorf("Valid transition failed: %v", err)
	}

	// Valid transition: Initializing -> Running
	if err := mgr.SetState(lifecycle.StateRunning); err != nil {
		t.Errorf("Valid transition failed: %v", err)
	}

	// Invalid transition: Running -> Initializing (cannot go back)
	if err := mgr.SetState(lifecycle.StateInitializing); err == nil {
		t.Error("Expected invalid transition to fail")
	}

	// Valid transition: Running -> ShuttingDown
	if err := mgr.SetState(lifecycle.StateShuttingDown); err != nil {
		t.Errorf("Valid transition failed: %v", err)
	}

	// Valid transition: ShuttingDown -> ShutdownComplete
	if err := mgr.SetState(lifecycle.StateShutdownComplete); err != nil {
		t.Errorf("Valid transition failed: %v", err)
	}

	// Invalid transition from terminal state
	if err := mgr.SetState(lifecycle.StateRunning); err == nil {
		t.Error("Expected transition from terminal state to fail")
	}
}
