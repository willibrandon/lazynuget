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

	// Check for subcommands first (before app initialization)
	// This allows utility commands to run without full bootstrap
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "encrypt-value":
			// Run encrypt-value subcommand
			exitCode := runEncryptValue(os.Args[2:])
			os.Exit(exitCode)
		}
	}

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

	// Initialize application with flags
	if err := app.Bootstrap(flags); err != nil {
		fmt.Fprintf(os.Stderr, "Startup failed: %v\n", err)
		os.Exit(ExitUserError)
	}

	// Run application and wait for shutdown signal
	// Even in non-interactive mode, we set up signal handlers so SIGINT/SIGTERM work correctly
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(ExitSystemError)
	}

	os.Exit(ExitSuccess)
}
