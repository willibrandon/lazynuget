package config

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigChangeType represents the type of configuration change that occurred.
// See: contracts/watcher.md
type ConfigChangeType string

const (
	// ConfigUpdated indicates the config file was modified
	ConfigUpdated ConfigChangeType = "updated"
	// ConfigDeleted indicates the config file was deleted
	ConfigDeleted ConfigChangeType = "deleted"
	// ConfigCreated indicates the config file was created
	ConfigCreated ConfigChangeType = "created"
)

// ConfigChangeEvent represents a configuration file change event.
// See: contracts/watcher.md
type ConfigChangeEvent struct {
	Type      ConfigChangeType
	FilePath  string
	Timestamp time.Time
	NewConfig *Config // nil if deleted or validation failed
	Error     error   // non-nil if reload failed
}

// ConfigWatcher watches a configuration file for changes and triggers reloads.
// See: contracts/watcher.md
type ConfigWatcher interface {
	// Watch starts watching the config file for changes.
	// Returns a channel for config change events and an error channel.
	Watch(ctx context.Context) (<-chan ConfigChangeEvent, <-chan error, error)

	// Stop stops the watcher and releases resources.
	Stop() error
}

// WatchOptions configures the config file watcher behavior.
// See: contracts/watcher.md
type WatchOptions struct {
	// ConfigFilePath is the path to the config file to watch
	ConfigFilePath string

	// LoadOptions are the options used to reload config
	LoadOptions LoadOptions

	// DebounceDelay is the delay before processing file change events
	// Default: 100ms per FR-044
	DebounceDelay time.Duration

	// OnReload is called when config is successfully reloaded
	OnReload func(*Config)

	// OnError is called when config reload fails
	OnError func(error)

	// OnFileDeleted is called when the config file is deleted
	OnFileDeleted func()
}

// configWatcher implements ConfigWatcher using fsnotify.
type configWatcher struct {
	opts       WatchOptions
	loader     ConfigLoader
	watcher    *fsnotify.Watcher
	stopCh     chan struct{}
	stoppedCh  chan struct{}
	mu         sync.Mutex
	lastConfig *Config
}

// NewConfigWatcher creates a new config file watcher.
func NewConfigWatcher(opts WatchOptions, loader ConfigLoader) (ConfigWatcher, error) {
	// Set default debounce delay if not specified
	if opts.DebounceDelay == 0 {
		opts.DebounceDelay = 100 * time.Millisecond // Per FR-044
	}

	// Validate file path
	if opts.ConfigFilePath == "" {
		return nil, fmt.Errorf("config file path is required")
	}

	// Get absolute path
	absPath, err := filepath.Abs(opts.ConfigFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}
	opts.ConfigFilePath = absPath

	// Create fsnotify watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Add file to watch
	if err := fsWatcher.Add(absPath); err != nil {
		fsWatcher.Close()
		return nil, fmt.Errorf("failed to watch config file: %w", err)
	}

	return &configWatcher{
		opts:      opts,
		loader:    loader,
		watcher:   fsWatcher,
		stopCh:    make(chan struct{}),
		stoppedCh: make(chan struct{}),
	}, nil
}

// Watch implements ConfigWatcher.Watch() (T100)
func (cw *configWatcher) Watch(ctx context.Context) (<-chan ConfigChangeEvent, <-chan error, error) {
	eventCh := make(chan ConfigChangeEvent, 10)
	errCh := make(chan error, 10)

	go cw.watchLoop(ctx, eventCh, errCh)

	return eventCh, errCh, nil
}

// watchLoop is the main event processing loop
func (cw *configWatcher) watchLoop(ctx context.Context, eventCh chan<- ConfigChangeEvent, errCh chan<- error) {
	defer close(cw.stoppedCh)
	defer close(eventCh)
	defer close(errCh)

	// Debounce timer (T102)
	var debounceTimer *time.Timer
	var pendingEvent fsnotify.Event

	for {
		select {
		case <-ctx.Done():
			return
		case <-cw.stopCh:
			return
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Debounce: wait for DebounceDelay after last event (T102)
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			pendingEvent = event

			debounceTimer = time.AfterFunc(cw.opts.DebounceDelay, func() {
				cw.handleFileEvent(ctx, pendingEvent, eventCh, errCh)
			})

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			errCh <- fmt.Errorf("file watcher error: %w", err)
		}
	}
}

// handleFileEvent processes a debounced file system event (T101)
func (cw *configWatcher) handleFileEvent(ctx context.Context, event fsnotify.Event, eventCh chan<- ConfigChangeEvent, errCh chan<- error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	changeEvent := ConfigChangeEvent{
		FilePath:  event.Name,
		Timestamp: time.Now(),
	}

	// Determine change type (T101)
	if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
		changeEvent.Type = ConfigDeleted
		changeEvent.Error = fmt.Errorf("config file deleted or renamed")

		// Trigger OnFileDeleted callback (T104)
		if cw.opts.OnFileDeleted != nil {
			go cw.opts.OnFileDeleted()
		}

		eventCh <- changeEvent
		return
	}

	if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
		if event.Has(fsnotify.Create) {
			changeEvent.Type = ConfigCreated
		} else {
			changeEvent.Type = ConfigUpdated
		}

		// Attempt to reload config (T103: reload validation)
		newConfig, err := cw.loader.Load(ctx, cw.opts.LoadOptions)
		if err != nil {
			// Reload failed - keep previous config (Per FR-047)
			changeEvent.Error = fmt.Errorf("config reload failed: %w", err)

			// Trigger OnError callback (T104)
			if cw.opts.OnError != nil {
				go cw.opts.OnError(changeEvent.Error)
			}

			eventCh <- changeEvent
			return
		}

		// Reload succeeded
		changeEvent.NewConfig = newConfig
		cw.lastConfig = newConfig

		// Trigger OnReload callback (T104)
		if cw.opts.OnReload != nil {
			go cw.opts.OnReload(newConfig)
		}

		eventCh <- changeEvent
	}
}

// Stop implements ConfigWatcher.Stop() (T105)
func (cw *configWatcher) Stop() error {
	close(cw.stopCh)
	<-cw.stoppedCh

	if cw.watcher != nil {
		return cw.watcher.Close()
	}

	return nil
}
