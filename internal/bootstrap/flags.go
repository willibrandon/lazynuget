package bootstrap

import (
	"flag"
	"fmt"
	"os"
)

// Flags holds parsed command-line flags.
type Flags struct {
	ConfigPath     string
	LogLevel       string
	ShowVersion    bool
	ShowHelp       bool
	NonInteractive bool
}

// ParseFlags parses command-line arguments and returns the flags.
// It returns true if the application should exit early (--version or --help).
func (app *App) ParseFlags(args []string) (*Flags, bool, error) {
	fs := flag.NewFlagSet("lazynuget", flag.ContinueOnError)
	fs.Usage = func() { /* Custom usage handled by ShowHelp */ }

	flags := &Flags{}

	fs.BoolVar(&flags.ShowVersion, "version", false, "Show version information")
	fs.BoolVar(&flags.ShowHelp, "help", false, "Show this help message")
	fs.StringVar(&flags.ConfigPath, "config", "", "Path to configuration file")
	fs.StringVar(&flags.LogLevel, "log-level", "info", "Set log level (debug|info|warn|error)")
	fs.BoolVar(&flags.NonInteractive, "non-interactive", false, "Run in non-interactive mode (no TUI)")

	if err := fs.Parse(args); err != nil {
		return nil, false, err
	}

	// Handle --version flag
	if flags.ShowVersion {
		ShowVersion(app.version)
		return flags, true, nil
	}

	// Handle --help flag
	if flags.ShowHelp {
		ShowHelp()
		return flags, true, nil
	}

	return flags, false, nil
}

// ShowHelp displays usage information for all available flags.
func ShowHelp() {
	fmt.Println("LazyNuGet - Terminal UI for NuGet package management")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lazynuget [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --version           Show version information and exit")
	fmt.Println("  --help              Show this help message and exit")
	fmt.Println("  --config PATH       Path to configuration file")
	fmt.Println("  --log-level LEVEL   Set log level (debug|info|warn|error)")
	fmt.Println("  --non-interactive   Run in non-interactive mode (no TUI)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  lazynuget                               # Start interactive TUI")
	fmt.Println("  lazynuget --version                     # Show version")
	fmt.Println("  lazynuget --config ~/.config/custom.yml # Use custom config")
	fmt.Println("  lazynuget --log-level debug             # Enable debug logging")
	fmt.Println()
}

// init customizes the default flag error output
func init() {
	flag.CommandLine.SetOutput(os.Stderr)
}
