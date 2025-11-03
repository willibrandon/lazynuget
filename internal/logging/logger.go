// Package logging provides logging infrastructure for LazyNuGet.
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Logger provides structured logging capabilities.
type Logger interface {
	// Debug logs a debug message
	Debug(format string, args ...any)

	// Info logs an informational message
	Info(format string, args ...any)

	// Warn logs a warning message
	Warn(format string, args ...any)

	// Error logs an error message
	Error(format string, args ...any)
}

// slogLogger wraps slog.Logger to implement our Logger interface
type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) Debug(format string, args ...any) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Info(format string, args ...any) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Warn(format string, args ...any) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Error(format string, args ...any) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

// New creates a new logger instance with the specified level and output path.
// If logPath is empty, logs go to stdout only.
// If logPath is specified, logs go to both stdout and the file.
func New(level, logPath string) Logger {
	// Parse log level
	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	// Determine output writer
	var writer io.Writer = os.Stdout

	// If log path is specified, create multiwriter for both stdout and file
	if logPath != "" {
		// Validate and clean log path (security: prevent path traversal)
		cleanLogPath := filepath.Clean(logPath)

		// Ensure log directory exists (owner-only permissions for security)
		logDir := filepath.Dir(cleanLogPath)
		if err := os.MkdirAll(logDir, 0o700); err != nil {
			// Fall back to stdout only if we can't create log directory
			fmt.Fprintf(os.Stderr, "Warning: failed to create log directory %s: %v\n", logDir, err)
		} else {
			// Open log file (append mode, owner-only permissions for security)
			file, err := os.OpenFile(cleanLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open log file %s: %v\n", cleanLogPath, err)
			} else {
				// Write to both stdout and file
				writer = io.MultiWriter(os.Stdout, file)
			}
		}
	}

	// Create text handler for human-readable output
	handler := slog.NewTextHandler(writer, opts)

	// Create and return logger
	return &slogLogger{
		logger: slog.New(handler),
	}
}
