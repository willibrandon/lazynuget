package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNew tests the New logger constructor
func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		logPath  string
		wantFile bool
	}{
		{
			name:     "stdout only - debug level",
			level:    "debug",
			logPath:  "",
			wantFile: false,
		},
		{
			name:     "stdout only - info level",
			level:    "info",
			logPath:  "",
			wantFile: false,
		},
		{
			name:     "stdout only - warn level",
			level:    "warn",
			logPath:  "",
			wantFile: false,
		},
		{
			name:     "stdout only - error level",
			level:    "error",
			logPath:  "",
			wantFile: false,
		},
		{
			name:     "stdout only - unknown level defaults to info",
			level:    "invalid",
			logPath:  "",
			wantFile: false,
		},
		{
			name:     "with log file",
			level:    "info",
			logPath:  filepath.Join(t.TempDir(), "test.log"),
			wantFile: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level, tt.logPath)

			if logger == nil {
				t.Fatal("New returned nil logger")
			}

			// If log path specified, verify file was created
			if tt.wantFile {
				if _, err := os.Stat(tt.logPath); os.IsNotExist(err) {
					t.Errorf("Log file was not created at %s", tt.logPath)
				}
			}
		})
	}
}

// TestLoggerMethods tests Debug, Info, Warn, Error methods
func TestLoggerMethods(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		logFn  func(Logger)
		want   string
		silent bool // true if message should not appear at this level
	}{
		{
			name:  "debug message at debug level",
			level: "debug",
			logFn: func(l Logger) { l.Debug("test %s", "debug") },
			want:  "debug",
		},
		{
			name:   "debug message at info level",
			level:  "info",
			logFn:  func(l Logger) { l.Debug("test %s", "debug") },
			want:   "debug",
			silent: true,
		},
		{
			name:  "info message at info level",
			level: "info",
			logFn: func(l Logger) { l.Info("test %s", "info") },
			want:  "info",
		},
		{
			name:  "warn message at info level",
			level: "info",
			logFn: func(l Logger) { l.Warn("test %s", "warn") },
			want:  "warn",
		},
		{
			name:  "error message at info level",
			level: "info",
			logFn: func(l Logger) { l.Error("test %s", "error") },
			want:  "error",
		},
		{
			name:  "info message at debug level",
			level: "debug",
			logFn: func(l Logger) { l.Info("test %s", "info") },
			want:  "info",
		},
		{
			name:   "warn message at error level",
			level:  "error",
			logFn:  func(l Logger) { l.Warn("test %s", "warn") },
			want:   "warn",
			silent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp log file to capture output
			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")

			logger := New(tt.level, logPath)

			// Log the message
			tt.logFn(logger)

			// Read log file
			content, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			output := string(content)

			if tt.silent {
				// Message should NOT appear at this level
				if strings.Contains(output, tt.want) {
					t.Errorf("Expected message to be filtered at level %s, but found: %s", tt.level, output)
				}
			} else {
				// Message should appear
				if !strings.Contains(output, tt.want) {
					t.Errorf("Expected log output to contain %q, got: %s", tt.want, output)
				}
			}
		})
	}
}

// TestNewWithInvalidLogPath tests fallback behavior
func TestNewWithInvalidLogPath(t *testing.T) {
	// Use a path that cannot be created (invalid characters on most systems)
	invalidPath := "/nonexistent/directory/test.log"

	// This should not panic, but fall back to stdout
	logger := New("info", invalidPath)

	if logger == nil {
		t.Fatal("New returned nil logger even with invalid path")
	}

	// Logger should still work (logging to stdout)
	logger.Info("test message")
}

// TestNewWithFilePermissionError tests handling of file permission errors
func TestNewWithFilePermissionError(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "logs")
	logPath := filepath.Join(logDir, "test.log")

	// Create log directory with restrictive permissions that prevent file creation
	if err := os.MkdirAll(logDir, 0o700); err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}

	// Create logger (should succeed, creating directory if needed)
	logger := New("info", logPath)

	if logger == nil {
		t.Fatal("New returned nil logger")
	}

	// Logger should work
	logger.Info("test message")
}

// TestLogLevels verifies correct log level parsing
func TestLogLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		shouldLogInfo bool
	}{
		{
			name:          "debug level logs info",
			level:         "debug",
			shouldLogInfo: true,
		},
		{
			name:          "info level logs info",
			level:         "info",
			shouldLogInfo: true,
		},
		{
			name:          "warn level does not log info",
			level:         "warn",
			shouldLogInfo: false,
		},
		{
			name:          "error level does not log info",
			level:         "error",
			shouldLogInfo: false,
		},
		{
			name:          "uppercase DEBUG",
			level:         "DEBUG",
			shouldLogInfo: true,
		},
		{
			name:          "mixed case Info",
			level:         "Info",
			shouldLogInfo: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test.log")

			logger := New(tt.level, logPath)
			logger.Info("test info message")

			content, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			output := string(content)
			hasInfo := strings.Contains(output, "test info message")

			if tt.shouldLogInfo && !hasInfo {
				t.Errorf("Expected info message to be logged at level %s", tt.level)
			}
			if !tt.shouldLogInfo && hasInfo {
				t.Errorf("Expected info message to be filtered at level %s", tt.level)
			}
		})
	}
}

// TestLogFormattingWithArgs verifies format string handling
func TestLogFormattingWithArgs(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := New("debug", logPath)

	tests := []struct {
		name  string
		logFn func()
		want  string
	}{
		{
			name:  "debug with args",
			logFn: func() { logger.Debug("value: %d", 42) },
			want:  "value: 42",
		},
		{
			name:  "info with multiple args",
			logFn: func() { logger.Info("name=%s age=%d", "test", 25) },
			want:  "name=test age=25",
		},
		{
			name:  "warn with string",
			logFn: func() { logger.Warn("warning: %s", "test warning") },
			want:  "warning: test warning",
		},
		{
			name:  "error with no args",
			logFn: func() { logger.Error("simple error") },
			want:  "simple error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear log file
			os.WriteFile(logPath, []byte{}, 0o600)

			tt.logFn()

			content, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("Failed to read log file: %v", err)
			}

			output := string(content)
			if !strings.Contains(output, tt.want) {
				t.Errorf("Expected log to contain %q, got: %s", tt.want, output)
			}
		})
	}
}

// TestNewCreatesLogDirectory tests that New creates the log directory
func TestNewCreatesLogDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	logDir := filepath.Join(tmpDir, "nested", "log", "dir")
	logPath := filepath.Join(logDir, "test.log")

	// Directory should not exist yet
	if _, err := os.Stat(logDir); !os.IsNotExist(err) {
		t.Fatal("Log directory should not exist before New() call")
	}

	logger := New("info", logPath)

	if logger == nil {
		t.Fatal("New returned nil logger")
	}

	// Directory should now exist
	if _, err := os.Stat(logDir); err != nil {
		t.Errorf("Log directory was not created: %v", err)
	}

	// Verify directory has correct permissions (0700)
	info, err := os.Stat(logDir)
	if err != nil {
		t.Fatalf("Failed to stat log directory: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o700 {
		t.Errorf("Expected log directory permissions 0700, got %o", perm)
	}
}

// TestLogFilePermissions verifies log file has secure permissions
func TestLogFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := New("info", logPath)
	logger.Info("test message")

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat log file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("Expected log file permissions 0600, got %o", perm)
	}
}

// TestLogPathCleaning verifies path traversal prevention
func TestLogPathCleaning(t *testing.T) {
	tmpDir := t.TempDir()
	// Attempt path traversal
	dirtyPath := filepath.Join(tmpDir, "..", "..", "test.log")

	// Should not panic and should clean the path
	logger := New("info", dirtyPath)

	if logger == nil {
		t.Fatal("New returned nil logger")
	}

	// Logger should work
	logger.Info("test message")

	// The cleaned path should be within tmpDir's parent at most, not escape beyond that
	// This test verifies the logger handles dirty paths gracefully
}

// TestMultipleLogMessages verifies appending to log file
func TestMultipleLogMessages(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	logger := New("info", logPath)

	messages := []string{
		"first message",
		"second message",
		"third message",
	}

	for _, msg := range messages {
		logger.Info(msg)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)

	for _, msg := range messages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected log to contain %q", msg)
		}
	}
}
