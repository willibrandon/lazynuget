//go:build windows

package platform

import (
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// resizeWatcher manages terminal resize event notifications on Windows
// Uses polling to detect terminal resize events since Windows doesn't have SIGWINCH
// See: T066, T067, T068, FR-016
type resizeWatcher struct {
	callbacks  []func(width, height int)
	mu         sync.RWMutex
	stopChan   chan struct{}
	stopped    bool
	lastWidth  int
	lastHeight int
}

// newResizeWatcher creates a new resize watcher for Windows
func newResizeWatcher() *resizeWatcher {
	// Get initial size
	width, height, _ := term.GetSize(int(os.Stdout.Fd()))

	w := &resizeWatcher{
		callbacks:  make([]func(width, height int), 0),
		stopChan:   make(chan struct{}),
		lastWidth:  width,
		lastHeight: height,
	}

	// Start polling goroutine
	go w.pollForResize()

	return w
}

// pollForResize polls terminal size and notifies callbacks on change
func (w *resizeWatcher) pollForResize() {
	ticker := time.NewTicker(500 * time.Millisecond) // Poll every 500ms
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get current terminal size
			width, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil {
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

			// Check if size changed
			w.mu.RLock()
			changed := width != w.lastWidth || height != w.lastHeight
			w.mu.RUnlock()

			if changed {
				// Update last known size
				w.mu.Lock()
				w.lastWidth = width
				w.lastHeight = height

				// Copy callbacks while holding lock
				callbacks := make([]func(int, int), len(w.callbacks))
				copy(callbacks, w.callbacks)
				w.mu.Unlock()

				// Notify all registered callbacks
				for _, callback := range callbacks {
					if callback != nil {
						callback(width, height)
					}
				}
			}

		case <-w.stopChan:
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

// watchResize implements the platform-specific resize watching for Windows
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
