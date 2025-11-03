package integration

import (
	"context"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/bootstrap"
)

// TestRunMethodExitsOnShutdown tests that Run() method properly waits for shutdown
func TestRunMethodExitsOnShutdown(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags := &bootstrap.Flags{
		NonInteractive: true,
	}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	// Run in goroutine since it blocks until shutdown
	runComplete := make(chan struct{})
	go func() {
		app.Run()
		close(runComplete)
	}()

	// Give Run a moment to start
	time.Sleep(50 * time.Millisecond)

	// Trigger shutdown
	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

	// Wait for Run to complete
	select {
	case <-runComplete:
		// Success - Run exited
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after Shutdown()")
	}
}

// TestRunMethodContextCancellation tests that Run() respects context cancellation
func TestRunMethodContextCancellation(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	flags := &bootstrap.Flags{
		NonInteractive: true,
	}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	// Run in goroutine
	runComplete := make(chan struct{})
	go func() {
		app.Run()
		close(runComplete)
	}()

	// Wait for context timeout
	<-ctx.Done()

	// Trigger shutdown
	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

	// Wait for Run to complete
	select {
	case <-runComplete:
		// Success - Run exited
	case <-time.After(2 * time.Second):
		t.Fatal("Run() did not exit after context cancellation")
	}
}

// TestMultipleRunCallsAreIdempotent tests that calling Run() multiple times is safe
func TestMultipleRunCallsAreIdempotent(t *testing.T) {
	app, err := bootstrap.NewApp("test", "test-commit", "2025-01-01")
	if err != nil {
		t.Fatalf("NewApp() failed: %v", err)
	}

	flags := &bootstrap.Flags{
		NonInteractive: true,
	}

	if err := app.Bootstrap(flags); err != nil {
		t.Fatalf("Bootstrap() failed: %v", err)
	}

	// Start first Run
	runComplete1 := make(chan struct{})
	go func() {
		app.Run()
		close(runComplete1)
	}()

	// Give first Run a moment to start
	time.Sleep(50 * time.Millisecond)

	// Try to start second Run (should be safe/idempotent)
	runComplete2 := make(chan struct{})
	go func() {
		app.Run()
		close(runComplete2)
	}()

	// Shutdown
	time.Sleep(50 * time.Millisecond)
	if err := app.Shutdown(); err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}

	// Both should complete
	select {
	case <-runComplete1:
	case <-time.After(2 * time.Second):
		t.Fatal("First Run() did not exit")
	}

	select {
	case <-runComplete2:
	case <-time.After(2 * time.Second):
		t.Fatal("Second Run() did not exit")
	}
}
