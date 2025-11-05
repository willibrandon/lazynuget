package platform

import (
	"testing"
)

// BenchmarkColorDepthDetection tests color depth detection performance
// Target: <10ms per FR-015
// See: T099, FR-015
func BenchmarkColorDepthDetection(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = detectColorDepth()
	}
}

// BenchmarkUnicodeSupport tests Unicode support detection performance
// See: T099, FR-016
func BenchmarkUnicodeSupport(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = detectUnicodeSupport()
	}
}

// BenchmarkTerminalSize tests terminal size detection performance
// See: T099, FR-017
func BenchmarkTerminalSize(b *testing.B) {
	// Create minimal terminal capabilities struct for benchmarking
	caps := &terminalCapabilities{}
	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = caps.GetSize()
	}
}

// BenchmarkIsTTY tests TTY detection performance
// See: T099
func BenchmarkIsTTY(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = IsTTY()
	}
}

// BenchmarkDetermineRunMode tests run mode determination performance
// See: T099
func BenchmarkDetermineRunMode(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = DetermineRunMode(false)
	}
}

// BenchmarkTerminalCapabilitiesFull tests full terminal capabilities detection
// This simulates the startup detection overhead
// See: T099, FR-015, FR-016, FR-017
func BenchmarkTerminalCapabilitiesFull(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		// Simulate full detection as done during app startup
		caps := &terminalCapabilities{}
		_ = detectColorDepth()
		_ = detectUnicodeSupport()
		_, _, _ = caps.GetSize()
		_ = IsTTY()
	}
}
