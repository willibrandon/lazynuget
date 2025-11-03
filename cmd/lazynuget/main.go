package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/willibrandon/lazynuget/internal/bootstrap"
)

// Version information (injected at build time via ldflags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Exit codes
const (
	ExitSuccess     = 0
	ExitUserError   = 1
	ExitSystemError = 2
)

func main() {
	// Layer 1 panic recovery: Ultimate safety net
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "FATAL PANIC: %v\nStack Trace:\n%s\n", r, debug.Stack())
			os.Exit(ExitSystemError)
		}
	}()

	// Create application instance
	app, err := bootstrap.NewApp(version, commit, date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create application: %v\n", err)
		os.Exit(ExitUserError)
	}

	// Parse command-line flags
	flags, exitEarly, err := app.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(ExitUserError)
	}

	// Exit early for --version or --help
	if exitEarly {
		os.Exit(ExitSuccess)
	}

	// Initialize application
	if err := app.Bootstrap(); err != nil {
		fmt.Fprintf(os.Stderr, "Startup failed: %v\n", err)
		os.Exit(ExitUserError)
	}

	// For now, just exit successfully after bootstrap
	// US2 will add lifecycle management and signal handling
	// US4 will add GUI/TUI support
	_ = flags // Flags will be used in US3 and US4

	os.Exit(ExitSuccess)
}
