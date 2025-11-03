package platform

import (
	"testing"
)

// TestRunModeString tests the String method
func TestRunModeString(t *testing.T) {
	tests := []struct {
		name string
		want string
		mode RunMode
	}{
		{
			name: "interactive mode",
			mode: RunModeInteractive,
			want: "interactive",
		},
		{
			name: "non-interactive mode",
			mode: RunModeNonInteractive,
			want: "non-interactive",
		},
		{
			name: "unknown mode",
			mode: RunMode(999),
			want: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.String()
			if got != tt.want {
				t.Errorf("RunMode.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRunModeIsInteractive tests the IsInteractive method
func TestRunModeIsInteractive(t *testing.T) {
	tests := []struct {
		name string
		mode RunMode
		want bool
	}{
		{
			name: "interactive mode is interactive",
			mode: RunModeInteractive,
			want: true,
		},
		{
			name: "non-interactive mode is not interactive",
			mode: RunModeNonInteractive,
			want: false,
		},
		{
			name: "unknown mode is not interactive",
			mode: RunMode(999),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.IsInteractive()
			if got != tt.want {
				t.Errorf("RunMode.IsInteractive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRunModeConstants verifies the constant values
func TestRunModeConstants(t *testing.T) {
	// Interactive should be 0 (iota starts at 0)
	if RunModeInteractive != 0 {
		t.Errorf("Expected RunModeInteractive to be 0, got %d", RunModeInteractive)
	}

	// NonInteractive should be 1
	if RunModeNonInteractive != 1 {
		t.Errorf("Expected RunModeNonInteractive to be 1, got %d", RunModeNonInteractive)
	}

	// They should be different
	if RunModeInteractive == RunModeNonInteractive {
		t.Error("RunModeInteractive and RunModeNonInteractive should have different values")
	}
}

// TestRunModeComparison verifies RunMode comparison works correctly
func TestRunModeComparison(t *testing.T) {
	interactive := RunModeInteractive
	nonInteractive := RunModeNonInteractive

	if interactive == nonInteractive {
		t.Error("Interactive and non-interactive modes should not be equal")
	}

	if interactive != RunModeInteractive {
		t.Error("RunMode comparison with constant failed")
	}

	if nonInteractive != RunModeNonInteractive {
		t.Error("RunMode comparison with constant failed")
	}
}

// TestRunModeZeroValue verifies zero value is interactive
func TestRunModeZeroValue(t *testing.T) {
	var mode RunMode

	// Zero value should be interactive (since it's the first iota value)
	if mode != RunModeInteractive {
		t.Errorf("Zero value RunMode should be RunModeInteractive, got %v", mode)
	}

	if !mode.IsInteractive() {
		t.Error("Zero value RunMode should be interactive")
	}

	if mode.String() != "interactive" {
		t.Errorf("Zero value RunMode.String() should be 'interactive', got %q", mode.String())
	}
}
