package config

import (
	"testing"
)

func TestMergeConfigs(t *testing.T) {
	// Create base config (defaults)
	base := GetDefaultConfig()
	baseLogLevel := base.LogLevel         // Should be "info"
	baseMaxOps := base.MaxConcurrentOps   // Should be 4

	// Create override config (simulating parsed file)
	override := &Config{
		LogLevel:         "debug",
		MaxConcurrentOps: 8,
		Theme:            "dark",
	}

	// Merge
	merged := mergeConfigs(base, override)

	// Verify overrides were applied
	if merged.LogLevel != "debug" {
		t.Errorf("Expected merged LogLevel=debug, got %s (base was %s)", merged.LogLevel, baseLogLevel)
	}

	if merged.MaxConcurrentOps != 8 {
		t.Errorf("Expected merged MaxConcurrentOps=8, got %d (base was %d)", merged.MaxConcurrentOps, baseMaxOps)
	}

	if merged.Theme != "dark" {
		t.Errorf("Expected merged Theme=dark, got %s", merged.Theme)
	}

	// Verify non-overridden values kept defaults
	if merged.CacheSize != base.CacheSize {
		t.Errorf("Expected CacheSize to remain default %d, got %d", base.CacheSize, merged.CacheSize)
	}
}
