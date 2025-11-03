package integration

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/willibrandon/lazynuget/internal/config"
)

// T089: Test that hot-reload is disabled by default
// See: FR-043
func TestHotReloadDisabledByDefault(t *testing.T) {
	defaults := config.GetDefaultConfig()

	// Verify hotReload is false by default
	if defaults.HotReload {
		t.Error("Expected hotReload to be false by default, got true")
	}
}

// T090: Test that config change is detected within 3 seconds when hot-reload enabled
// See: FR-045
func TestHotReloadDetectsChanges(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	initialConfig := `logLevel: info
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	loader := config.NewConfigLoader()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.LogLevel != "info" || !cfg.HotReload {
		t.Errorf("Initial config incorrect: logLevel=%s, hotReload=%v", cfg.LogLevel, cfg.HotReload)
	}

	// Start watcher
	watcher, err := config.NewConfigWatcher(config.WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    opts,
	}, loader)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	eventCh, errCh, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	// Modify config file
	time.Sleep(200 * time.Millisecond) // Let watcher initialize
	updatedConfig := `logLevel: debug
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0o644); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Wait for change event (should arrive within 3 seconds)
	select {
	case event := <-eventCh:
		if event.Type != config.ConfigUpdated {
			t.Errorf("Expected ConfigUpdated, got %s", event.Type)
		}
		if event.NewConfig == nil {
			t.Error("Expected NewConfig to be set")
		} else if event.NewConfig.LogLevel != "debug" {
			t.Errorf("Expected reloaded logLevel=debug, got %s", event.NewConfig.LogLevel)
		}
	case err := <-errCh:
		t.Fatalf("Unexpected error: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("Config change not detected within 3 seconds")
	}
}

// T091: Test that invalid reload keeps previous config
// See: FR-047
func TestInvalidReloadKeepsPreviousConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	validConfig := `logLevel: debug
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(validConfig), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewConfigLoader()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	cfg, err := loader.Load(ctx, opts)
	if err != nil {
		t.Fatalf("Initial load failed: %v", err)
	}

	watcher, err := config.NewConfigWatcher(config.WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    opts,
	}, loader)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	eventCh, errCh, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	// Write invalid config
	time.Sleep(200 * time.Millisecond)
	invalidConfig := `invalid yaml {{`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0o644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Should get error event, but watcher keeps previous config
	select {
	case event := <-eventCh:
		if event.Error == nil {
			t.Error("Expected error for invalid config")
		}
		// NewConfig should be nil on error
		if event.NewConfig != nil {
			t.Error("Expected NewConfig to be nil on error")
		}
	case err := <-errCh:
		t.Logf("Got error (expected): %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("Expected error event for invalid config")
	}

	// Previous config (cfg) should still be valid
	if cfg.LogLevel != "debug" {
		t.Errorf("Previous config was corrupted: logLevel=%s", cfg.LogLevel)
	}
}

// T092: Test hot-reload success notification
// See: FR-048
func TestHotReloadSuccessNotification(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	initialConfig := `logLevel: info
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewConfigLoader()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	var reloadMu sync.Mutex
	reloadCalled := false
	watcher, err := config.NewConfigWatcher(config.WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    opts,
		OnReload: func(_ *config.Config) {
			reloadMu.Lock()
			reloadCalled = true
			reloadMu.Unlock()
		},
	}, loader)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	eventCh, _, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	// Update config
	time.Sleep(200 * time.Millisecond)
	updatedConfig := `logLevel: debug
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(updatedConfig), 0o644); err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// Wait for success event
	select {
	case event := <-eventCh:
		if event.Error != nil {
			t.Errorf("Expected success, got error: %v", event.Error)
		}
		if event.NewConfig == nil {
			t.Error("Expected NewConfig on success")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for reload")
	}

	// Give callback time to execute
	time.Sleep(100 * time.Millisecond)
	reloadMu.Lock()
	called := reloadCalled
	reloadMu.Unlock()
	if !called {
		t.Error("OnReload callback was not called")
	}
}

// T093: Test hot-reload failure notification
// See: FR-048
func TestHotReloadFailureNotification(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	validConfig := `logLevel: info
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(validConfig), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewConfigLoader()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	var errorMu sync.Mutex
	errorCalled := false
	watcher, err := config.NewConfigWatcher(config.WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    opts,
		OnError: func(_ error) {
			errorMu.Lock()
			errorCalled = true
			errorMu.Unlock()
		},
	}, loader)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	eventCh, _, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	// Write invalid config
	time.Sleep(200 * time.Millisecond)
	invalidConfig := `{{{ invalid`
	if err := os.WriteFile(configPath, []byte(invalidConfig), 0o644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// Wait for error event
	select {
	case event := <-eventCh:
		if event.Error == nil {
			t.Error("Expected error event")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for error")
	}

	// Give callback time to execute
	time.Sleep(100 * time.Millisecond)
	errorMu.Lock()
	called := errorCalled
	errorMu.Unlock()
	if !called {
		t.Error("OnError callback was not called")
	}
}

// T094: Test config file deletion triggers callback
func TestConfigFileDeletion(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	initialConfig := `logLevel: info
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewConfigLoader()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	var deletedMu sync.Mutex
	deletedCalled := false
	watcher, err := config.NewConfigWatcher(config.WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    opts,
		OnFileDeleted: func() {
			deletedMu.Lock()
			deletedCalled = true
			deletedMu.Unlock()
		},
	}, loader)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	eventCh, _, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	// Delete file
	time.Sleep(200 * time.Millisecond)
	if err := os.Remove(configPath); err != nil {
		t.Fatalf("Failed to delete config: %v", err)
	}

	// Wait for deletion event
	select {
	case event := <-eventCh:
		if event.Type != config.ConfigDeleted {
			t.Errorf("Expected ConfigDeleted, got %s", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for deletion event")
	}

	// Give callback time to execute
	time.Sleep(100 * time.Millisecond)
	deletedMu.Lock()
	called := deletedCalled
	deletedMu.Unlock()
	if !called {
		t.Error("OnFileDeleted callback was not called")
	}
}

// T095: Test rapid successive writes are debounced properly
func TestRapidWritesDebounced(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	initialConfig := `logLevel: info
hotReload: true
`
	if err := os.WriteFile(configPath, []byte(initialConfig), 0o644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewConfigLoader()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := config.LoadOptions{
		ConfigFilePath: configPath,
		EnvVarPrefix:   "LAZYNUGET_",
	}

	watcher, err := config.NewConfigWatcher(config.WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    opts,
		DebounceDelay:  100 * time.Millisecond, // 100ms debounce
	}, loader)
	if err != nil {
		t.Fatalf("NewConfigWatcher() failed: %v", err)
	}
	defer watcher.Stop()

	eventCh, _, err := watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	time.Sleep(200 * time.Millisecond) // Let watcher initialize

	// Write multiple times rapidly (5 writes in 50ms)
	for range 5 {
		config := `logLevel: debug
hotReload: true
`
		if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Should only get ONE event due to debouncing
	eventCount := 0
	timeout := time.After(500 * time.Millisecond)

drainEvents:
	for {
		select {
		case <-eventCh:
			eventCount++
		case <-timeout:
			break drainEvents
		}
	}

	// Should get exactly 1 event (debounced)
	if eventCount != 1 {
		t.Errorf("Expected 1 debounced event, got %d", eventCount)
	}
}
