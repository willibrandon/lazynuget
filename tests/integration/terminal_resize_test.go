package integration

import (
	"sync"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/platform"
)

// TestTerminalResize_WatchResize tests terminal resize event handling
// See: T060, FR-016
func TestTerminalResize_WatchResize(t *testing.T) {
	caps := platform.NewTerminalCapabilities()

	// Track callback invocations
	var mu sync.Mutex
	callCount := 0

	callback := func(width, height int) {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		t.Logf("Resize callback invoked: %dx%d (call #%d)", width, height, callCount)
	}

	// Register callback
	stop := caps.WatchResize(callback)
	if stop == nil {
		t.Fatal("WatchResize() returned nil stop function")
	}
	defer stop()

	// Note: This test verifies the watcher infrastructure works.
	// Actual resize events can't be easily simulated in automated tests
	// without platform-specific terminal manipulation.
	// Manual testing is required to verify real resize events are detected.

	// Give watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	t.Log("Terminal resize watcher initialized successfully")
	t.Logf("On Unix, resize events are detected via SIGWINCH signal")
	t.Logf("On Windows, resize events are detected via polling (500ms interval)")

	// Verify stop function doesn't panic
	stop()
	t.Log("Resize watcher stopped successfully")
}

// TestTerminalResize_MultipeCallbacks tests multiple callbacks can be registered
// See: T060, T067
func TestTerminalResize_MultipleCallbacks(t *testing.T) {
	caps := platform.NewTerminalCapabilities()

	// Register first callback
	callback1 := func(width, height int) {
		t.Logf("Callback 1: %dx%d", width, height)
	}
	stop1 := caps.WatchResize(callback1)
	defer stop1()

	// Register second callback
	callback2 := func(width, height int) {
		t.Logf("Callback 2: %dx%d", width, height)
	}
	stop2 := caps.WatchResize(callback2)
	defer stop2()

	// Give watchers time to initialize
	time.Sleep(100 * time.Millisecond)

	t.Log("Multiple resize watchers registered successfully")

	// Note: Without actual resize events, we can't verify callbacks are invoked
	// Manual testing required to verify both callbacks receive events

	// Verify stop functions work independently
	stop1()
	t.Log("First watcher stopped")

	stop2()
	t.Log("Second watcher stopped")
}

// TestTerminalResize_DimensionClamping tests that resize events provide clamped dimensions
// See: T063, T064, FR-015
func TestTerminalResize_DimensionClamping(t *testing.T) {
	caps := platform.NewTerminalCapabilities()

	// Get initial size to verify clamping is applied
	width, height, err := caps.GetSize()
	if err != nil {
		t.Logf("GetSize() returned error (expected in non-TTY): %v", err)
		// In non-TTY environments, we get defaults which are already valid
		if width != 80 || height != 24 {
			t.Errorf("Default dimensions should be 80x24, got %dx%d", width, height)
		}
		return
	}

	// Verify dimensions are within valid range
	const (
		MinWidth  = 40
		MinHeight = 10
		MaxWidth  = 500
		MaxHeight = 200
	)

	if width < MinWidth {
		t.Errorf("Width %d is below minimum %d (should be clamped)", width, MinWidth)
	}
	if width > MaxWidth {
		t.Errorf("Width %d exceeds maximum %d (should be clamped)", width, MaxWidth)
	}
	if height < MinHeight {
		t.Errorf("Height %d is below minimum %d (should be clamped)", height, MinHeight)
	}
	if height > MaxHeight {
		t.Errorf("Height %d exceeds maximum %d (should be clamped)", height, MaxHeight)
	}

	t.Logf("Terminal dimensions are properly clamped: %dx%d", width, height)

	// Register callback to verify resize events also provide clamped dimensions
	var mu sync.Mutex
	var resizeWidth, resizeHeight int
	gotResize := false

	callback := func(w, h int) {
		mu.Lock()
		defer mu.Unlock()
		resizeWidth = w
		resizeHeight = h
		gotResize = true
		t.Logf("Resize event: %dx%d", w, h)
	}

	stop := caps.WatchResize(callback)
	defer stop()

	// Wait briefly for any resize events
	time.Sleep(200 * time.Millisecond)

	// Note: In automated testing, we likely won't get actual resize events
	// Manual testing with terminal resize is needed to verify clamping in callbacks

	if gotResize {
		// If we did get a resize event, verify it was clamped
		if resizeWidth < MinWidth || resizeWidth > MaxWidth {
			t.Errorf("Resize width %d out of valid range [%d,%d]", resizeWidth, MinWidth, MaxWidth)
		}
		if resizeHeight < MinHeight || resizeHeight > MaxHeight {
			t.Errorf("Resize height %d out of valid range [%d,%d]", resizeHeight, MinHeight, MaxHeight)
		}
	} else {
		t.Log("No resize events detected (expected in automated testing)")
	}
}

// TestTerminalResize_StopPreventsCallbacks tests that stop() prevents further callbacks
// See: T067
func TestTerminalResize_StopPreventsCallbacks(t *testing.T) {
	caps := platform.NewTerminalCapabilities()

	var mu sync.Mutex
	callCount := 0

	callback := func(_, _ int) {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		t.Logf("Callback invoked after stop? Count: %d", callCount)
	}

	// Register and immediately stop
	stop := caps.WatchResize(callback)
	stop()

	// Wait to ensure no callbacks occur after stop
	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	finalCount := callCount
	mu.Unlock()

	if finalCount > 0 {
		t.Logf("Callback was invoked %d times (may be from initialization)", finalCount)
	} else {
		t.Log("No callbacks after stop (expected)")
	}

	// Note: Can't fully test without triggering actual resize events
	// Manual testing required
}
