package platform

import (
	"os"
	"testing"
)

// TestDetectColorDepth tests color depth detection
// Note: Most tests will return ColorNone when running in non-TTY environments (e.g., CI)
func TestDetectColorDepth(t *testing.T) {
	tests := []struct {
		name          string
		colorterm     string
		term          string
		expectedDepth ColorDepth
		noColor       bool
	}{
		{
			name:          "NO_COLOR set",
			noColor:       true,
			term:          "xterm-256color",
			expectedDepth: ColorNone,
		},
		{
			name:          "dumb terminal",
			term:          "dumb",
			expectedDepth: ColorNone,
		},
		{
			name:          "empty TERM",
			term:          "",
			expectedDepth: ColorNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origNoColor := os.Getenv("NO_COLOR")
			origColorterm := os.Getenv("COLORTERM")
			origTerm := os.Getenv("TERM")

			// Clean up after test
			defer func() {
				if origNoColor != "" {
					os.Setenv("NO_COLOR", origNoColor)
				} else {
					os.Unsetenv("NO_COLOR")
				}
				if origColorterm != "" {
					os.Setenv("COLORTERM", origColorterm)
				} else {
					os.Unsetenv("COLORTERM")
				}
				if origTerm != "" {
					os.Setenv("TERM", origTerm)
				} else {
					os.Unsetenv("TERM")
				}
			}()

			// Set test env vars
			os.Unsetenv("NO_COLOR")
			os.Unsetenv("COLORTERM")
			os.Setenv("TERM", tt.term)

			if tt.noColor {
				os.Setenv("NO_COLOR", "1")
			}
			if tt.colorterm != "" {
				os.Setenv("COLORTERM", tt.colorterm)
			}

			// Detect color depth
			depth := detectColorDepth()

			// Verify result
			if depth != tt.expectedDepth {
				t.Errorf("detectColorDepth() = %v (%s), want %v (%s)",
					depth, depth.String(), tt.expectedDepth, tt.expectedDepth.String())
			}
		})
	}
}

// TestDetectUnicodeSupport tests Unicode support detection
func TestDetectUnicodeSupport(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		lcAll    string
		lcCtype  string
		expected bool
	}{
		{
			name:     "UTF-8 in LANG",
			lang:     "en_US.UTF-8",
			expected: true,
		},
		{
			name:     "UTF8 in LANG (no dash)",
			lang:     "en_US.UTF8",
			expected: true,
		},
		{
			name:     "UTF-8 in LC_ALL",
			lcAll:    "en_US.UTF-8",
			expected: true,
		},
		{
			name:     "UTF-8 in LC_CTYPE",
			lcCtype:  "en_US.UTF-8",
			expected: true,
		},
		{
			name:     "no UTF-8",
			lang:     "C",
			expected: false,
		},
		{
			name:     "empty locale",
			lang:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars
			origLang := os.Getenv("LANG")
			origLcAll := os.Getenv("LC_ALL")
			origLcCtype := os.Getenv("LC_CTYPE")

			// Clean up after test
			defer func() {
				if origLang != "" {
					os.Setenv("LANG", origLang)
				} else {
					os.Unsetenv("LANG")
				}
				if origLcAll != "" {
					os.Setenv("LC_ALL", origLcAll)
				} else {
					os.Unsetenv("LC_ALL")
				}
				if origLcCtype != "" {
					os.Setenv("LC_CTYPE", origLcCtype)
				} else {
					os.Unsetenv("LC_CTYPE")
				}
			}()

			// Set test env vars
			os.Setenv("LANG", tt.lang)
			if tt.lcAll != "" {
				os.Setenv("LC_ALL", tt.lcAll)
			} else {
				os.Unsetenv("LC_ALL")
			}
			if tt.lcCtype != "" {
				os.Setenv("LC_CTYPE", tt.lcCtype)
			} else {
				os.Unsetenv("LC_CTYPE")
			}

			// Detect Unicode support
			supported := detectUnicodeSupport()

			// Verify result
			if supported != tt.expected {
				t.Errorf("detectUnicodeSupport() = %v, want %v", supported, tt.expected)
			}
		})
	}
}

// TestNewTerminalCapabilities tests TerminalCapabilities factory
func TestNewTerminalCapabilities(t *testing.T) {
	caps := NewTerminalCapabilities()

	if caps == nil {
		t.Fatal("NewTerminalCapabilities() returned nil")
	}

	// Verify GetColorDepth returns a valid value
	depth := caps.GetColorDepth()
	if depth < ColorNone || depth > ColorTrueColor {
		t.Errorf("GetColorDepth() returned invalid value: %v", depth)
	}

	// Verify SupportsUnicode returns a boolean
	unicode := caps.SupportsUnicode()
	t.Logf("SupportsUnicode() = %v", unicode)

	// Verify IsTTY returns a boolean
	isTTY := caps.IsTTY()
	t.Logf("IsTTY() = %v", isTTY)
}

// TestGetSize tests terminal size detection
func TestGetSize(t *testing.T) {
	caps := NewTerminalCapabilities()

	width, height, err := caps.GetSize()
	// Size detection may fail in non-TTY environments (e.g., CI)
	if err != nil {
		t.Logf("GetSize() returned error (expected in non-TTY): %v", err)
		// Verify we got default values
		if width != 80 || height != 24 {
			t.Errorf("GetSize() with error should return defaults (80, 24), got (%d, %d)", width, height)
		}
		return
	}

	// If no error, verify we got reasonable values
	if width <= 0 || height <= 0 {
		t.Errorf("GetSize() = (%d, %d), expected positive values", width, height)
	}

	t.Logf("Terminal size: %dx%d", width, height)
}

// TestWatchResize tests resize watcher (stub implementation)
func TestWatchResize(t *testing.T) {
	caps := NewTerminalCapabilities()

	called := false
	callback := func(_, _ int) {
		called = true
	}

	stop := caps.WatchResize(callback)
	if stop == nil {
		t.Error("WatchResize() returned nil stop function")
	}

	// Call stop function (should not panic)
	stop()

	// Note: callback won't be called in stub implementation
	// This is expected behavior for now
	t.Logf("Resize callback called: %v (expected: false for stub)", called)
}
