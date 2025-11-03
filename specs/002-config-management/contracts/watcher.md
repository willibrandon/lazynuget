# ConfigWatcher Contract

Configuration file watcher for hot-reload functionality.

## Primary Interface

```go
package contracts

import (
	"context"
	"time"
)

// ConfigWatcher monitors the configuration file for changes and triggers reloads.
// Hot-reload is opt-in (disabled by default) for safety.
//
// Implementation: internal/config/watcher.go (via fsnotify)
// See: FR-043 through FR-049
type ConfigWatcher interface {
	// Watch starts monitoring the config file for changes.
	// Returns immediately after starting the background watcher goroutine.
	//
	// When a file change is detected:
	//   1. Debounce for 100ms (handle editors saving temp files)
	//   2. Reload config using ConfigLoader
	//   3. Validate new config
	//   4. If valid: Call reload callback with new config
	//   5. If invalid: Log error, keep existing config, call error callback
	//
	// Handles these file events:
	//   - Write/Modify: Reload config
	//   - Delete: Fall back to defaults, notify via callback
	//   - Rename: Treat as delete (file no longer at watched path)
	//
	// Performance: Reload latency <3 seconds from file modification to callback (FR-045)
	//
	// Thread safety: Callbacks executed in watcher goroutine, must be thread-safe or use channels
	//
	// Returns:
	//   - error: If watcher cannot be started (file not found, permissions, etc.)
	Watch(ctx context.Context, opts WatchOptions) error

	// Stop stops the file watcher and releases resources.
	// Blocks until watcher goroutine exits.
	// Safe to call multiple times (no-op after first call).
	Stop() error
}
```

## Supporting Types

```go
// WatchOptions configures the file watcher behavior.
type WatchOptions struct {
	// ConfigFilePath is the absolute path to the config file to watch.
	// Must be a regular file (not directory).
	ConfigFilePath string

	// Loader is used to reload config when file changes.
	// Must be the same loader used for initial load.
	Loader ConfigLoader

	// OnReload callback is invoked when config successfully reloads.
	// Receives the new validated config.
	// Must be thread-safe or use channels to communicate with main thread.
	// Cannot be nil.
	OnReload func(newConfig *Config)

	// OnError callback is invoked when config reload fails validation.
	// Receives validation errors.
	// Application continues with previous valid config.
	// Cannot be nil.
	OnError func(errors []ValidationError)

	// OnFileDeleted callback is invoked when config file is deleted.
	// Application typically falls back to defaults.
	// Cannot be nil.
	OnFileDeleted func()

	// DebounceInterval is the wait time after last file event before triggering reload.
	// Default: 100ms (handles editors saving temp files + atomic writes)
	// Range: 10ms - 5s
	DebounceInterval time.Duration

	// Logger for logging watcher events and errors.
	Logger Logger
}

// ConfigChangeNotifier is an alternative callback-based interface for hot-reload.
// Use when callbacks are more convenient than channels.
type ConfigChangeNotifier interface {
	// NotifyChange signals that configuration has changed.
	// Used by ConfigWatcher to notify the application of config updates.
	NotifyChange(ctx context.Context, event ConfigChangeEvent)
}

// ConfigChangeEvent describes a configuration change event.
type ConfigChangeEvent struct {
	// Type of change
	Type ConfigChangeType

	// NewConfig is the new configuration (if Type is ConfigReloaded)
	NewConfig *Config

	// Errors from validation (if Type is ConfigReloadFailed)
	Errors []ValidationError

	// Timestamp of the change event
	Timestamp time.Time

	// SourcePath is the path to the file that changed
	SourcePath string
}

// ConfigChangeType describes the type of configuration change.
type ConfigChangeType int

const (
	ConfigReloaded      ConfigChangeType = iota // Config successfully reloaded
	ConfigReloadFailed                          // Config reload failed validation
	ConfigFileDeleted                           // Config file was deleted
)
```
