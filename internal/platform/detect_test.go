package platform

import (
	"runtime"
	"testing"
)

func TestDetectPlatform(t *testing.T) {
	p, err := detectPlatform()
	if err != nil {
		t.Fatalf("detectPlatform() failed: %v", err)
	}

	// Verify OS detection
	expectedOS := runtime.GOOS
	if p.OS() != expectedOS {
		t.Errorf("OS() = %q, want %q", p.OS(), expectedOS)
	}

	// Verify architecture detection
	expectedArch := runtime.GOARCH
	if p.Arch() != expectedArch {
		t.Errorf("Arch() = %q, want %q", p.Arch(), expectedArch)
	}

	// Verify version is not empty
	if p.Version() == "" {
		t.Error("Version() returned empty string")
	}

	// Test convenience methods
	switch runtime.GOOS {
	case "windows":
		if !p.IsWindows() {
			t.Error("IsWindows() should return true on Windows")
		}
		if p.IsDarwin() || p.IsLinux() {
			t.Error("IsDarwin() and IsLinux() should return false on Windows")
		}

	case "darwin":
		if !p.IsDarwin() {
			t.Error("IsDarwin() should return true on macOS")
		}
		if p.IsWindows() || p.IsLinux() {
			t.Error("IsWindows() and IsLinux() should return false on macOS")
		}

	case "linux":
		if !p.IsLinux() {
			t.Error("IsLinux() should return true on Linux")
		}
		if p.IsWindows() || p.IsDarwin() {
			t.Error("IsWindows() and IsDarwin() should return false on Linux")
		}
	}
}

func TestNew(t *testing.T) {
	// Test singleton behavior
	p1, err1 := New()
	if err1 != nil {
		t.Fatalf("New() failed: %v", err1)
	}

	p2, err2 := New()
	if err2 != nil {
		t.Fatalf("New() failed on second call: %v", err2)
	}

	// Should return the same instance
	if p1 != p2 {
		t.Error("New() should return singleton instance")
	}
}

func TestColorDepthString(t *testing.T) {
	tests := []struct {
		depth ColorDepth
		want  string
	}{
		{ColorNone, "none"},
		{ColorBasic16, "16-color"},
		{ColorExtended256, "256-color"},
		{ColorTrueColor, "true-color"},
		{ColorDepth(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.depth.String()
			if got != tt.want {
				t.Errorf("ColorDepth(%d).String() = %q, want %q", tt.depth, got, tt.want)
			}
		})
	}
}
