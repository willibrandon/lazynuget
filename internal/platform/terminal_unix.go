//go:build !windows

package platform

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/term"
)

// resizeWatcher manages terminal resize event notifications on Unix systems
// Uses SIGWINCH signal to detect terminal resize events
// See: T065, T067, T068, FR-016
type resizeWatcher struct {
	sigChan   chan os.Signal
	stopChan  chan struct{}
	callbacks []func(width, height int)
	mu        sync.RWMutex
	stopped   bool
}

// newResizeWatcher creates a new resize watcher for Unix platforms
func newResizeWatcher() *resizeWatcher {
	w := &resizeWatcher{
		callbacks: make([]func(width, height int), 0),
		sigChan:   make(chan os.Signal, 1),
		stopChan:  make(chan struct{}),
	}

	// Register for SIGWINCH (window size change) signal
	signal.Notify(w.sigChan, syscall.SIGWINCH)

	// Start goroutine to handle resize signals
	go w.handleSignals()

	return w
}

// handleSignals processes SIGWINCH signals and notifies callbacks
func (w *resizeWatcher) handleSignals() {
	for {
		select {
		case <-w.sigChan:
			// Get new terminal size
			width, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
				// If we can't get size, skip this event
				continue
			}

			// Apply clamping (same as GetSize method)
			const (
				MinWidth  = 40
				MinHeight = 10
				MaxWidth  = 500
				MaxHeight = 200
			)

			if width < MinWidth {
				width = MinWidth
			} else if width > MaxWidth {
				width = MaxWidth
			}

			if height < MinHeight {
				height = MinHeight
			} else if height > MaxHeight {
				height = MaxHeight
			}

			// Notify all registered callbacks
			w.mu.RLock()
			callbacks := make([]func(int, int), len(w.callbacks))
			copy(callbacks, w.callbacks)
			w.mu.RUnlock()

			for _, callback := range callbacks {
				if callback != nil {
					callback(width, height)
				}
			}

		case <-w.stopChan:
			// Stop signal received, clean up and exit
			signal.Stop(w.sigChan)
			close(w.sigChan)
			return
		}
	}
}

// addCallback registers a callback for resize events
func (w *resizeWatcher) addCallback(callback func(width, height int)) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.callbacks = append(w.callbacks, callback)
}

// stop stops the resize watcher and cleans up resources
func (w *resizeWatcher) stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.stopped {
		w.stopped = true
		close(w.stopChan)
	}
}

// Global resize watcher instance (singleton)
var (
	globalWatcher     *resizeWatcher
	globalWatcherOnce sync.Once
	globalWatcherMu   sync.Mutex
)

// watchResize implements the platform-specific resize watching for Unix
// Returns a stop function that unregisters the callback
func watchResize(callback func(width, height int)) (stop func()) {
	globalWatcherMu.Lock()
	defer globalWatcherMu.Unlock()

	// Initialize global watcher once
	globalWatcherOnce.Do(func() {
		globalWatcher = newResizeWatcher()
	})

	// Add callback to watcher
	globalWatcher.addCallback(callback)

	// Return stop function
	return func() {
		// For now, stopping removes the entire watcher
		// A more sophisticated implementation would remove just this callback
		globalWatcher.stop()
	}
}
