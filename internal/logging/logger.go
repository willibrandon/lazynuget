// Package logging provides logging infrastructure for LazyNuGet.
package logging

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

// noOpLogger is a logger that does nothing (for stub implementation).
type noOpLogger struct{}

func (l *noOpLogger) Debug(format string, args ...interface{}) {}
func (l *noOpLogger) Info(format string, args ...interface{})  {}
func (l *noOpLogger) Warn(format string, args ...interface{})  {}
func (l *noOpLogger) Error(format string, args ...interface{}) {}

// New creates a new logger instance.
// For now, this returns a no-op logger stub.
func New(level, logPath string) Logger {
	return &noOpLogger{}
}
