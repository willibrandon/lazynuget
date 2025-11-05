# Contract: Terminal Capabilities

## Type: ColorDepth

Represents the terminal's color support level.

```go
package contracts

// ColorDepth represents terminal color support level
type ColorDepth int

const (
	ColorNone        ColorDepth = 0        // No color support
	ColorBasic16     ColorDepth = 16       // 16 ANSI colors
	ColorExtended256 ColorDepth = 256      // 256-color palette
	ColorTrueColor   ColorDepth = 16777216 // 24-bit true color
)
```

### Color Depth Values

| Value | Name | Description | Use Case |
|-------|------|-------------|----------|
| `0` | `ColorNone` | No color support | `TERM=dumb`, redirected output, basic terminals |
| `16` | `ColorBasic16` | 16 ANSI colors | Standard terminal emulators, SSH sessions |
| `256` | `ColorExtended256` | 256-color palette | Modern terminals (iTerm2, Windows Terminal) |
| `16777216` | `ColorTrueColor` | 24-bit RGB colors | Latest terminals with `COLORTERM=truecolor` |

### Detection Logic

Color depth is detected from environment variables in priority order:

1. **TrueColor**: `COLORTERM=truecolor` or `COLORTERM=24bit`
2. **256-color**: `TERM` contains `256color` (e.g., `xterm-256color`)
3. **16-color**: `TERM` is set and not `dumb`
4. **No color**: `TERM=dumb`, `NO_COLOR` env var set, or not a TTY

## Interface: TerminalCapabilities

Provides terminal feature detection for adaptive UI rendering.

```go
// TerminalCapabilities provides terminal feature detection
type TerminalCapabilities interface {
	// GetColorDepth returns detected color support level
	GetColorDepth() ColorDepth

	// SupportsUnicode returns true if terminal can display Unicode
	SupportsUnicode() bool

	// GetSize returns terminal dimensions (width, height in characters)
	GetSize() (width int, height int, err error)

	// IsTTY returns true if stdout is connected to an interactive terminal
	IsTTY() bool

	// WatchResize registers a callback for terminal resize events
	// Returns a stop function to unregister the watcher
	WatchResize(callback func(width, height int)) (stop func())
}
```

## Methods

### GetColorDepth() ColorDepth

Returns the detected color support level.

**Usage**:
```go
term := platform.NewTerminalCapabilities()

switch term.GetColorDepth() {
case platform.ColorTrueColor:
    // Use RGB colors: \x1b[38;2;255;128;0m
    fmt.Println("\x1b[38;2;255;128;0mOrange text\x1b[0m")

case platform.ColorExtended256:
    // Use 256-color palette: \x1b[38;5;214m
    fmt.Println("\x1b[38;5;214mOrange text\x1b[0m")

case platform.ColorBasic16:
    // Use basic ANSI colors: \x1b[33m (yellow, closest to orange)
    fmt.Println("\x1b[33mYellow text\x1b[0m")

default: // ColorNone
    // Plain text only
    fmt.Println("Plain text")
}
```

**Note**: Color depth is detected once at startup and cached.

### SupportsUnicode() bool

Returns true if the terminal can display Unicode characters.

**Detection logic**:
- **Unix**: Check if `LANG` or `LC_ALL` contains `UTF-8` (case-insensitive)
- **Windows**: Check if console code page is 65001 (UTF-8)
- **Fallback**: Return `false` if detection uncertain

**Usage**:
```go
term := platform.NewTerminalCapabilities()

var checkmark string
if term.SupportsUnicode() {
    checkmark = "✓"  // U+2713 CHECK MARK
} else {
    checkmark = "+"  // ASCII fallback
}

fmt.Printf("%s Task completed\n", checkmark)
```

**Common Unicode → ASCII fallbacks**:
- `✓` → `+` (checkmark)
- `✗` → `-` (cross)
- `▶` → `>` (arrow)
- `│` → `|` (box drawing vertical)
- `─` → `-` (box drawing horizontal)

### GetSize() (width int, height int, err error)

Returns the current terminal dimensions in characters.

**Behavior**:
- Queries terminal via `golang.org/x/term.GetSize()`
- Falls back to `COLUMNS` and `LINES` environment variables if query fails
- Returns error if terminal size cannot be determined

**Validation**:
- Width: minimum 40, maximum 500 (per assumption A-005)
- Height: minimum 10, maximum 200
- Values outside range are clamped to limits

**Usage**:
```go
term := platform.NewTerminalCapabilities()

width, height, err := term.GetSize()
if err != nil {
    // Handle error (not a TTY, or detection failed)
    width, height = 80, 24 // Use defaults
}

fmt.Printf("Terminal size: %dx%d\n", width, height)

if width < 100 {
    // Use compact layout
} else {
    // Use full-width layout
}
```

**Note**: For dynamic resize handling, use `WatchResize()` instead of polling `GetSize()`.

### IsTTY() bool

Returns true if stdout is connected to an interactive terminal (not redirected or piped).

**Detection**:
- Uses `golang.org/x/term.IsTerminal(os.Stdout.Fd())`
- Returns `false` if output is redirected to file or piped to another command

**Usage**:
```go
term := platform.NewTerminalCapabilities()

if term.IsTTY() {
    // Show interactive TUI
    runInteractiveMode()
} else {
    // Output is redirected, use non-interactive mode
    runNonInteractiveMode()
}
```

**Note**: This is also used by `internal/platform/detect.go` for `RunMode` determination (see spec 001).

### WatchResize(callback func(width, height int)) (stop func())

Registers a callback function to be invoked when the terminal is resized.

**Platform-specific implementation**:
- **Unix**: Listens for `SIGWINCH` signal (per FR-028)
- **Windows**: Monitors console events via Windows API

**Callback behavior**:
- Called immediately with current dimensions when registered
- Called again whenever terminal is resized
- Receives new width and height as parameters
- Should complete quickly (UI redraw triggers should be async)

**Returns**: A `stop` function that unregisters the callback and cleans up resources.

**Usage**:
```go
term := platform.NewTerminalCapabilities()

// Register resize handler
stop := term.WatchResize(func(width, height int) {
    log.Printf("Terminal resized to %dx%d\n", width, height)

    // Trigger UI redraw (should be async)
    go redrawUI(width, height)
})

// Unregister when done
defer stop()

// Run application
select {} // Keep running
```

**Important notes**:
- Multiple callbacks can be registered (all will be invoked)
- Callbacks should not block (schedule async work if needed)
- Call `stop()` to prevent goroutine leaks
- Resize events may fire multiple times during a single resize operation

## Implementation Notes

### Caching Strategy

- `GetColorDepth()` and `SupportsUnicode()` are detected once and cached
- `GetSize()` queries fresh dimensions each time (unless watching for resize)
- `IsTTY()` is detected once and cached (output destination doesn't change)

### Graceful Degradation

Per spec clarification: "What happens when terminal reports capabilities incorrectly?"

**Strategy**: Accept reported capabilities and provide user override mechanism if needed (future feature). If UI renders incorrectly, users can set environment variables:

```bash
# Force basic 16-color mode
export TERM=xterm

# Disable color entirely
export NO_COLOR=1

# Force Unicode off
export LANG=C
```

### Performance

- Terminal detection: <10ms total (startup overhead)
- GetSize(): <1ms (system call)
- Resize callbacks: Triggered by OS signal (negligible overhead)

## Example: Adaptive UI Rendering

```go
import "github.com/yourusername/lazynuget/internal/platform"

func renderStatus(status string, success bool) string {
    term := platform.NewTerminalCapabilities()

    // Choose symbol based on Unicode support
    var symbol string
    if success {
        symbol = "✓"
        if !term.SupportsUnicode() {
            symbol = "+"
        }
    } else {
        symbol = "✗"
        if !term.SupportsUnicode() {
            symbol = "-"
        }
    }

    // Choose color based on color depth
    var colorStart, colorEnd string
    if term.GetColorDepth() >= platform.ColorBasic16 {
        if success {
            colorStart = "\x1b[32m" // Green
        } else {
            colorStart = "\x1b[31m" // Red
        }
        colorEnd = "\x1b[0m" // Reset
    }

    return fmt.Sprintf("%s%s %s%s", colorStart, symbol, status, colorEnd)
}

func main() {
    fmt.Println(renderStatus("Build succeeded", true))
    // TTY + Unicode + Color: "✓ Build succeeded" (in green)
    // TTY + ASCII + Color:   "+ Build succeeded" (in green)
    // Redirected:            "+ Build succeeded" (no color)
}
```
