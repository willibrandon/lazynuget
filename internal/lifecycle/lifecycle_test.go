package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type mockLogger struct {
	logs []string
}

func (m *mockLogger) Info(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf("INFO: "+format, args...))
}

func (m *mockLogger) Debug(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf("DEBUG: "+format, args...))
}

func (m *mockLogger) Warn(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf("WARN: "+format, args...))
}

func (m *mockLogger) Error(format string, args ...any) {
	m.logs = append(m.logs, fmt.Sprintf("ERROR: "+format, args...))
}

func TestStateTransitions(t *testing.T) {
	tests := []struct {
		name        string
		transitions []State
		wantErr     bool
	}{
		{
			name:        "valid transition sequence",
			transitions: []State{StateInitializing, StateRunning, StateShuttingDown, StateShutdownComplete},
			wantErr:     false,
		},
		{
			name:        "invalid skip to running",
			transitions: []State{StateRunning},
			wantErr:     true,
		},
		{
			name:        "invalid backward transition",
			transitions: []State{StateInitializing, StateRunning, StateInitializing},
			wantErr:     true,
		},
		{
			name:        "failed state from uninitialized",
			transitions: []State{StateFailed},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewManager(30 * time.Second)

			var lastErr error
			for _, state := range tt.transitions {
				err := mgr.SetState(state)
				if err != nil {
					lastErr = err
					break
				}
			}

			if tt.wantErr && lastErr == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantErr && lastErr != nil {
				t.Errorf("unexpected error: %v", lastErr)
			}
		})
	}
}

func TestShutdownTimeout(t *testing.T) {
	mgr := NewManager(100 * time.Millisecond)
	logger := &mockLogger{}

	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "slow-handler",
		Priority: 100,
		Handler: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(1 * time.Second):
				return nil
			}
		},
	})

	if err := mgr.SetState(StateInitializing); err != nil {
		t.Fatalf("SetState(Initializing) failed: %v", err)
	}
	if err := mgr.SetState(StateRunning); err != nil {
		t.Fatalf("SetState(Running) failed: %v", err)
	}

	ctx := context.Background()
	err := mgr.Shutdown(ctx, logger)

	if err == nil {
		t.Error("expected shutdown to fail due to timeout")
	}

	if err != nil && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestShutdownHandlerPriority(t *testing.T) {
	mgr := NewManager(30 * time.Second)
	logger := &mockLogger{}

	executionOrder := []string{}

	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "high-priority",
		Priority: 10,
		Handler: func(context.Context) error {
			executionOrder = append(executionOrder, "high-priority")
			return nil
		},
	})

	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "low-priority",
		Priority: 100,
		Handler: func(context.Context) error {
			executionOrder = append(executionOrder, "low-priority")
			return nil
		},
	})

	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "medium-priority",
		Priority: 50,
		Handler: func(context.Context) error {
			executionOrder = append(executionOrder, "medium-priority")
			return nil
		},
	})

	mgr.SetState(StateInitializing)
	mgr.SetState(StateRunning)

	ctx := context.Background()
	if err := mgr.Shutdown(ctx, logger); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	expected := []string{"high-priority", "medium-priority", "low-priority"}
	if len(executionOrder) != len(expected) {
		t.Errorf("expected %d handlers, got %d", len(expected), len(executionOrder))
	}

	for i, name := range expected {
		if i >= len(executionOrder) || executionOrder[i] != name {
			t.Errorf("expected handler %d to be %s, got %v", i, name, executionOrder)
		}
	}
}

func TestShutdownHandlerErrors(t *testing.T) {
	mgr := NewManager(30 * time.Second)
	logger := &mockLogger{}

	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "failing-handler",
		Priority: 100,
		Handler: func(context.Context) error {
			return errors.New("intentional failure")
		},
	})

	successCalled := false
	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "success-handler",
		Priority: 200,
		Handler: func(context.Context) error {
			successCalled = true
			return nil
		},
	})

	mgr.SetState(StateInitializing)
	mgr.SetState(StateRunning)

	ctx := context.Background()
	err := mgr.Shutdown(ctx, logger)

	if err == nil {
		t.Error("expected error from failed handler")
	}

	if !successCalled {
		t.Error("success handler was not called despite earlier failure")
	}
}

func TestShutdownHandlerPanic(t *testing.T) {
	mgr := NewManager(30 * time.Second)
	logger := &mockLogger{}

	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "panicking-handler",
		Priority: 100,
		Handler: func(context.Context) error {
			panic("intentional panic")
		},
	})

	successCalled := false
	mgr.RegisterShutdownHandler(ShutdownHandler{
		Name:     "success-handler",
		Priority: 200,
		Handler: func(context.Context) error {
			successCalled = true
			return nil
		},
	})

	mgr.SetState(StateInitializing)
	mgr.SetState(StateRunning)

	ctx := context.Background()
	err := mgr.Shutdown(ctx, logger)

	if err == nil {
		t.Error("expected error from panicking handler")
	}

	if !strings.Contains(err.Error(), "panic") {
		t.Errorf("expected error to mention panic, got: %v", err)
	}

	if !successCalled {
		t.Error("success handler was not called despite earlier panic")
	}
}

func TestGetState(t *testing.T) {
	mgr := NewManager(30 * time.Second)

	if mgr.GetState() != StateUninitialized {
		t.Errorf("expected initial state Uninitialized, got %v", mgr.GetState())
	}

	mgr.SetState(StateInitializing)
	if mgr.GetState() != StateInitializing {
		t.Errorf("expected state Initializing, got %v", mgr.GetState())
	}

	mgr.SetState(StateRunning)
	if mgr.GetState() != StateRunning {
		t.Errorf("expected state Running, got %v", mgr.GetState())
	}
}

func TestConcurrentStateAccess(t *testing.T) {
	mgr := NewManager(30 * time.Second)

	if err := mgr.SetState(StateInitializing); err != nil {
		t.Fatalf("Failed to set initializing state: %v", err)
	}
	if err := mgr.SetState(StateRunning); err != nil {
		t.Fatalf("Failed to set running state: %v", err)
	}

	done := make(chan bool, 10)
	for range 10 {
		go func() {
			for range 100 {
				_ = mgr.GetState()
			}
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}
