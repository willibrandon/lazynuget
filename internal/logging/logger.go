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
	Debug(format string, args ...interface{})

	// Info logs an informational message
	Info(format string, args ...interface{})

	// Warn logs a warning message
	Warn(format string, args ...interface{})

	// Error logs an error message
	Error(format string, args ...interface{})
}

// slogLogger wraps slog.Logger to implement our Logger interface
type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) Debug(format string, args ...interface{}) {
	l.logger.Debug(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Info(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Warn(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *slogLogger) Error(format string, args ...interface{}) {
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
		// Ensure log directory exists
		logDir := filepath.Dir(logPath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			// Fall back to stdout only if we can't create log directory
			fmt.Fprintf(os.Stderr, "Warning: failed to create log directory %s: %v\n", logDir, err)
		} else {
			// Open log file (append mode)
			file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to open log file %s: %v\n", logPath, err)
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
