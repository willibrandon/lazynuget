# Quickstart: Cross-Platform Infrastructure

## Overview

This guide shows how to use the platform abstraction layer for common tasks. All examples assume you've imported the necessary packages.

## Platform Detection

```go
import "github.com/yourusername/lazynuget/internal/platform"

func main() {
    p := platform.New()

    if p.IsWindows() {
        // Windows-specific logic
    } else if p.IsDarwin() {
        // macOS-specific logic
    } else if p.IsLinux() {
        // Linux-specific logic
    }

    fmt.Printf("Running on %s/%s\n", p.OS(), p.Arch())
}
```

## Path Resolution

```go
import "github.com/yourusername/lazynuget/internal/platform"

func getConfigPath(filename string) (string, error) {
    resolver := platform.NewPathResolver()

    // Get platform-appropriate config directory
    configDir, err := resolver.ConfigDir()
    if err != nil {
        return "", err
    }

    // Combine with filename using platform separators
    path := filepath.Join(configDir, filename)

    // Normalize to platform format
    return resolver.Normalize(path), nil
}
```

## Terminal Capabilities

```go
import "github.com/yourusername/lazynuget/internal/platform"

func setupUI() {
    term := platform.NewTerminalCapabilities()

    // Check color support
    switch term.GetColorDepth() {
    case platform.ColorTrueColor:
        // Use 24-bit RGB colors
    case platform.ColorExtended256:
        // Use 256-color palette
    case platform.ColorBasic16:
        // Use 16 ANSI colors
    default:
        // Monochrome mode
    }

    // Check Unicode support
    checkmark := "√"  // ✓ in Unicode
    if !term.SupportsUnicode() {
        checkmark = "+" // ASCII fallback
    }

    // Get dimensions
    width, height, _ := term.GetSize()
    fmt.Printf("Terminal: %dx%d\n", width, height)

    // Watch for resize
    stop := term.WatchResize(func(w, h int) {
        fmt.Printf("Resized to %dx%d\n", w, h)
        // Redraw UI
    })
    defer stop()
}
```

## Process Spawning

```go
import "github.com/yourusername/lazynuget/internal/platform"

func runDotnetRestore(projectPath string) error {
    spawner := platform.NewProcessSpawner()

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
```

## Testing Platform-Specific Code

```go
// paths_test.go
func TestNormalize(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantUnix string
        wantWin  string
    }{
        {
            name: "mixed separators",
            input: "C:/Users/Dev/config.yml",
            wantUnix: "C:/Users/Dev/config.yml",  // unchanged on Unix
            wantWin: "C:\\Users\\Dev\\config.yml", // normalized on Windows
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resolver := platform.NewPathResolver()
            got := resolver.Normalize(tt.input)

            want := tt.wantUnix
            if runtime.GOOS == "windows" {
                want = tt.wantWin
            }

            if got != want {
                t.Errorf("Normalize() = %q, want %q", got, want)
            }
        })
    }
}
```

## Common Patterns

### Graceful Degradation (Read-Only Cache)

```go
func ensureCacheDir() error {
    resolver := platform.NewPathResolver()
    cacheDir, err := resolver.CacheDir()
    if err != nil {
        return err
    }

    // Try to create cache directory
    if err := os.MkdirAll(cacheDir, 0755); err != nil {
        // Log warning but continue (cache operations will fail gracefully)
        log.Warn("Cache directory is read-only, using in-memory cache")
        return nil
    }

    return nil
}
```

### Environment Variable Precedence (Windows vs XDG)

```go
func resolveConfigDir() (string, error) {
    p := platform.New()

    if p.IsWindows() {
        // Windows: APPDATA always wins (even if XDG vars set in WSL)
        if appdata := os.Getenv("APPDATA"); appdata != "" {
            return filepath.Join(appdata, "lazynuget"), nil
        }
        return "", errors.New("APPDATA not set")
    }

    if p.IsLinux() {
        // Linux: Check XDG first, fall back to ~/.config
        if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
            return filepath.Join(xdgConfig, "lazynuget"), nil
        }

        homeDir, _ := os.UserHomeDir()
        return filepath.Join(homeDir, ".config", "lazynuget"), nil
    }

    // macOS: Use ~/Library/Application Support
    homeDir, _ := os.UserHomeDir()
    return filepath.Join(homeDir, "Library", "Application Support", "lazynuget"), nil
}
```
