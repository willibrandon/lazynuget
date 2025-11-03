package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkHotReloadLatency tests FR-045: Hot-reload latency <3s
// Measures time from file change to reload completion
func BenchmarkHotReloadLatency(b *testing.B) {
	// Create temp config file
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	initialContent := `
version: "1.0"
theme: dark
logLevel: info
`
	err := os.WriteFile(configPath, []byte(initialContent), 0o644)
	if err != nil {
		b.Fatalf("Failed to create config: %v", err)
	}

	// Create loader
	loader := NewLoader()

	// Create watcher with callback
	reloadComplete := make(chan time.Time, 1)

	watcher, err := NewConfigWatcher(WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    LoadOptions{},
		DebounceDelay:  100 * time.Millisecond,

		OnReload: func(_ *Config) {
			reloadComplete <- time.Now()
		},

		OnError: func(err error) {
			b.Logf("Reload error: %v", err)
		},
	}, loader)
	if err != nil {
		b.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching
	ctx := context.Background()
	_, _, err = watcher.Watch(ctx)
	if err != nil {
		b.Fatalf("Failed to start watcher: %v", err)
	}

	// Allow watcher to initialize
	time.Sleep(200 * time.Millisecond)

	// Measure reload latency
	latencies := make([]time.Duration, 0, b.N)

	for i := 0; b.Loop(); i++ {
		// Modify config file
		modifiedContent := `
version: "1.0"
theme: light
logLevel: debug
maxConcurrentOps: ` + string(rune('0'+i%10)) + `
`
		changeStart := time.Now()

		err := os.WriteFile(configPath, []byte(modifiedContent), 0o644)
		if err != nil {
			b.Fatalf("Failed to write config: %v", err)
		}

		// Wait for reload to complete (with timeout)
		select {
		case reloadTime := <-reloadComplete:
			latency := reloadTime.Sub(changeStart)
			latencies = append(latencies, latency)

		case <-time.After(5 * time.Second):
			b.Fatalf("Reload timeout after 5s")
		}

		// Small delay between iterations
		time.Sleep(100 * time.Millisecond)
	}

	b.StopTimer()

	// Calculate average latency
	var totalLatency time.Duration
	for _, lat := range latencies {
		totalLatency += lat
	}
	avgLatency := totalLatency / time.Duration(len(latencies))

	b.Logf("Average hot-reload latency: %v", avgLatency)
	b.Logf("Latencies: %v", latencies)

	// Verify performance: <3s per FR-045
	if avgLatency > 3*time.Second {
		b.Errorf("Hot-reload latency %v exceeds 3s target (FR-045)", avgLatency)
	}

	// Also check that all individual latencies are <3s
	for i, lat := range latencies {
		if lat > 3*time.Second {
			b.Errorf("Reload #%d latency %v exceeds 3s target", i, lat)
		}
	}
}

// TestHotReloadLatencyManual is a manual test for hot-reload latency
// Run with: go test -v -run TestHotReloadLatencyManual
func TestHotReloadLatencyManual(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping manual hot-reload latency test in short mode")
	}

	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	initialContent := `
version: "1.0"
theme: dark
logLevel: info
`
	err := os.WriteFile(configPath, []byte(initialContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Create loader
	loader := NewLoader()

	// Create watcher with callback
	reloadComplete := make(chan time.Time, 1)

	watcher, err := NewConfigWatcher(WatchOptions{
		ConfigFilePath: configPath,
		LoadOptions:    LoadOptions{},
		DebounceDelay:  100 * time.Millisecond,

		OnReload: func(newConfig *Config) {
			t.Logf("Config reloaded: theme=%s", newConfig.Theme)
			reloadComplete <- time.Now()
		},

		OnError: func(err error) {
			t.Logf("Reload error: %v", err)
		},
	}, loader)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Start watching
	ctx := context.Background()
	_, _, err = watcher.Watch(ctx)
	if err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Allow watcher to initialize
	time.Sleep(200 * time.Millisecond)

	// Perform 5 reload tests
	latencies := make([]time.Duration, 0, 5)

	for i := range 5 {
		t.Logf("Test iteration %d", i+1)

		// Modify config file
		modifiedContent := `
version: "1.0"
theme: light
logLevel: debug
maxConcurrentOps: ` + string(rune('0'+i%10)) + `
`
		changeStart := time.Now()

		err := os.WriteFile(configPath, []byte(modifiedContent), 0o644)
		if err != nil {
			t.Fatalf("Failed to write config: %v", err)
		}

		// Wait for reload to complete
		select {
		case reloadTime := <-reloadComplete:
			latency := reloadTime.Sub(changeStart)
			latencies = append(latencies, latency)
			t.Logf("Reload latency: %v", latency)

		case <-time.After(5 * time.Second):
			t.Fatalf("Reload timeout after 5s")
		}

		// Delay between iterations
		time.Sleep(200 * time.Millisecond)
	}

	// Calculate average
	var totalLatency time.Duration
	for _, lat := range latencies {
		totalLatency += lat
	}
	avgLatency := totalLatency / time.Duration(len(latencies))

	t.Logf("Average latency: %v", avgLatency)
	t.Logf("All latencies: %v", latencies)

	// Verify <3s requirement
	if avgLatency > 3*time.Second {
		t.Errorf("Average latency %v exceeds 3s target (FR-045)", avgLatency)
	}

	for i, lat := range latencies {
		if lat > 3*time.Second {
			t.Errorf("Iteration %d latency %v exceeds 3s target", i+1, lat)
		}
	}
}
