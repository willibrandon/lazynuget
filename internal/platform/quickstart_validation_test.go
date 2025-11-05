package platform

import (
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
)

// TestQuickstartExamples validates that all code examples from quickstart.md compile and run
// See: T108, specs/003-platform-abstraction/quickstart.md
func TestQuickstartExamples(t *testing.T) {
	t.Run("example_1_platform_detection", func(t *testing.T) {
		p, err := New()
		if err != nil {
			t.Fatalf("Failed to detect platform: %v", err)
		}

		if p.IsWindows() {
			// Windows-specific logic
			t.Logf("Running on Windows")
		} else if p.IsDarwin() {
			// macOS-specific logic
			t.Logf("Running on macOS")
		} else if p.IsLinux() {
			// Linux-specific logic
			t.Logf("Running on Linux")
		}

		t.Logf("Running on %s/%s", p.OS(), p.Arch())
	})

	t.Run("example_2_path_resolution", func(t *testing.T) {
		p, err := New()
		if err != nil {
			t.Fatalf("platform detection failed: %v", err)
		}

		resolver, err := NewPathResolver(p)
		if err != nil {
			t.Fatalf("failed to create path resolver: %v", err)
		}

		// Get platform-appropriate config directory
		configDir, err := resolver.ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() failed: %v", err)
		}

		// Combine with filename using platform separators
		path := filepath.Join(configDir, "config.yml")

		// Normalize to platform format
		normalized := resolver.Normalize(path)

		t.Logf("Config path: %s", normalized)

		if normalized == "" {
			t.Error("Normalized path is empty")
		}
	})

	t.Run("example_3_terminal_capabilities", func(t *testing.T) {
		term := NewTerminalCapabilities()

		// Check color support
		switch term.GetColorDepth() {
		case ColorTrueColor:
			t.Log("Using 24-bit RGB colors")
		case ColorExtended256:
			t.Log("Using 256-color palette")
		case ColorBasic16:
			t.Log("Using 16 ANSI colors")
		default:
			t.Log("Monochrome mode")
		}

		// Check Unicode support
		checkmark := "√" // ✓ in Unicode
		if !term.SupportsUnicode() {
			checkmark = "+" // ASCII fallback
		}
		t.Logf("Checkmark: %s", checkmark)

		// Get dimensions
		width, height, _ := term.GetSize()
		t.Logf("Terminal: %dx%d", width, height)
	})

	t.Run("example_4_process_spawning_type_check", func(t *testing.T) {
		// Just verify the API compiles (don't actually run dotnet)
		spawner := NewProcessSpawner()

		// Type check: verify spawner has Run method with correct signature
		if spawner == nil {
			t.Fatal("NewProcessSpawner() returned nil")
		}

		// This just verifies the code compiles
		_ = func(projectPath string) error {
			result, err := spawner.Run(
				"dotnet",
				[]string{"restore", projectPath},
				filepath.Dir(projectPath),
				nil, // inherit environment
			)
			if err != nil {
				return fmt.Errorf("failed to spawn dotnet: %w", err)
			}

			if result.ExitCode != 0 {
				return fmt.Errorf("dotnet restore failed: %s", result.Stderr)
			}

			fmt.Println(result.Stdout)
			return nil
		}

		t.Log("Process spawning API compiles correctly")
	})

	t.Run("example_5_testing_pattern", func(t *testing.T) {
		p, err := New()
		if err != nil {
			t.Fatalf("Failed to create platform: %v", err)
		}

		resolver, err := NewPathResolver(p)
		if err != nil {
			t.Fatalf("Failed to create path resolver: %v", err)
		}

		input := "C:/Users/Dev/config.yml"
		got := resolver.Normalize(input)

		wantUnix := "C:/Users/Dev/config.yml"   // unchanged on Unix
		wantWin := "C:\\Users\\Dev\\config.yml" // normalized on Windows

		want := wantUnix
		if runtime.GOOS == "windows" {
			want = wantWin
		}

		if got != want {
			t.Errorf("Normalize() = %q, want %q", got, want)
		}
	})
}
