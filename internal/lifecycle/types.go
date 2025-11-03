package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// State represents the application lifecycle state
type State int

const (
	StateUninitialized State = iota // Before bootstrap
	StateInitializing               // During bootstrap
	StateRunning                    // Normal operation
	StateShuttingDown               // Graceful shutdown in progress
	StateShutdownComplete           // Shutdown finished
	StateFailed                     // Fatal error occurred
)

func (s State) String() string {
	switch s {
	case StateUninitialized:
		return "Uninitialized"
	case StateInitializing:
		return "Initializing"
	case StateRunning:
		return "Running"
	case StateShuttingDown:
		return "ShuttingDown"
	case StateShutdownComplete:
		return "ShutdownComplete"
	case StateFailed:
		return "Failed"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// Manager manages the application lifecycle state machine
type Manager struct {
	mu               sync.RWMutex
	state            State
	startTime        time.Time
	shutdownHandlers []ShutdownHandler
	shutdownTimeout  time.Duration
}

// ShutdownHandler is a function called during graceful shutdown
type ShutdownHandler struct {
	Name     string
	Priority int // Lower numbers run first
	Handler  func(context.Context) error
}

// NewManager creates a new lifecycle manager
func NewManager(shutdownTimeout time.Duration) *Manager {
	return &Manager{
		state:            StateUninitialized,
		shutdownHandlers: make([]ShutdownHandler, 0),
		shutdownTimeout:  shutdownTimeout,
	}
}

// GetState returns the current lifecycle state
func (m *Manager) GetState() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

// SetState transitions to a new state with validation
func (m *Manager) SetState(newState State) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate state transition
	if !m.isValidTransition(m.state, newState) {
		return fmt.Errorf("invalid state transition from %s to %s", m.state, newState)
	}

	m.state = newState

	// Record start time when entering running state
	if newState == StateRunning {
		m.startTime = time.Now()
	}

	return nil
}

// isValidTransition checks if a state transition is allowed
func (m *Manager) isValidTransition(from, to State) bool {
	// Valid transitions based on lifecycle flow
	validTransitions := map[State][]State{
		StateUninitialized:    {StateInitializing, StateFailed},
		StateInitializing:     {StateRunning, StateFailed},
		StateRunning:          {StateShuttingDown, StateFailed},
		StateShuttingDown:     {StateShutdownComplete, StateFailed},
		StateShutdownComplete: {},        // Terminal state
		StateFailed:           {},        // Terminal state
	}

	allowedStates, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, allowed := range allowedStates {
		if allowed == to {
			return true
		}
	}

	return false
}

// RegisterShutdownHandler adds a handler to be called during shutdown
func (m *Manager) RegisterShutdownHandler(handler ShutdownHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shutdownHandlers = append(m.shutdownHandlers, handler)
}

// GetUptime returns the duration since the app entered running state
func (m *Manager) GetUptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.state == StateRunning || m.state == StateShuttingDown || m.state == StateShutdownComplete {
		return time.Since(m.startTime)
	}

	return 0
}
