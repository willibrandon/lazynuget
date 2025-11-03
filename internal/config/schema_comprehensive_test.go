package config

import (
	"testing"
)

// TestIsHotReloadable tests the IsHotReloadable method thoroughly
func TestIsHotReloadable(t *testing.T) {
	schema := GetConfigSchema()

	tests := []struct {
		name           string
		settingPath    string
		wantReloadable bool
	}{
		// Hot-reloadable settings
		{name: "theme is hot reloadable", settingPath: "theme", wantReloadable: true},
		{name: "logLevel is hot reloadable", settingPath: "logLevel", wantReloadable: true},
		{name: "compactMode is hot reloadable", settingPath: "compactMode", wantReloadable: true},
		{name: "showHints is hot reloadable", settingPath: "showHints", wantReloadable: true},
		{name: "showLineNumbers is hot reloadable", settingPath: "showLineNumbers", wantReloadable: true},
		{name: "colorScheme.border is hot reloadable", settingPath: "colorScheme.border", wantReloadable: true},
		{name: "colorScheme.text is hot reloadable", settingPath: "colorScheme.text", wantReloadable: true},

		// Not hot-reloadable settings
		{name: "dotnetPath is not hot reloadable", settingPath: "dotnetPath", wantReloadable: false},
		{name: "version is not hot reloadable", settingPath: "version", wantReloadable: false},
		{name: "logDir is not hot reloadable", settingPath: "logDir", wantReloadable: false},

		// Non-existent settings (should default to false)
		{name: "unknown setting is not hot reloadable", settingPath: "unknownSetting", wantReloadable: false},
		{name: "empty string is not hot reloadable", settingPath: "", wantReloadable: false},
		{name: "invalid nested path is not hot reloadable", settingPath: "foo.bar.baz", wantReloadable: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := schema.IsHotReloadable(tt.settingPath)
			if got != tt.wantReloadable {
				t.Errorf("IsHotReloadable(%q) = %v, want %v", tt.settingPath, got, tt.wantReloadable)
			}
		})
	}
}

// TestGetConfigSchema tests that schema is properly initialized
func TestGetConfigSchema(t *testing.T) {
	schema := GetConfigSchema()

	if schema == nil {
		t.Fatal("GetConfigSchema() returned nil")
	}

	if schema.Settings == nil {
		t.Fatal("Schema.Settings is nil")
	}

	if len(schema.Settings) == 0 {
		t.Error("Schema should have settings defined")
	}

	// Verify some expected settings exist
	expectedSettings := []string{
		"logLevel",
		"theme",
		"maxConcurrentOps",
		"compactMode",
		"showHints",
		"dotnetPath",
	}

	for _, settingPath := range expectedSettings {
		if _, exists := schema.Settings[settingPath]; !exists {
			t.Errorf("Expected setting %q not found in schema", settingPath)
		}
	}
}

// TestSchemaSettingProperties tests that settings have proper metadata
func TestSchemaSettingProperties(t *testing.T) {
	schema := GetConfigSchema()

	// Test logLevel setting
	if setting, exists := schema.Settings["logLevel"]; exists {
		if setting.Path != "logLevel" {
			t.Errorf("logLevel path = %v, want logLevel", setting.Path)
		}
		if !setting.HotReloadable {
			t.Error("logLevel should be hot reloadable")
		}
		if len(setting.Constraints) == 0 {
			t.Error("logLevel should have constraints")
		}
	} else {
		t.Error("logLevel setting not found")
	}

	// Test dotnetPath setting
	if setting, exists := schema.Settings["dotnetPath"]; exists {
		if setting.HotReloadable {
			t.Error("dotnetPath should not be hot reloadable")
		}
	}

	// Test theme setting
	if setting, exists := schema.Settings["theme"]; exists {
		if !setting.HotReloadable {
			t.Error("theme should be hot reloadable")
		}
		// Should have enum constraint
		hasEnumConstraint := false
		for _, constraint := range setting.Constraints {
			if constraint.Type == "enum" {
				hasEnumConstraint = true
				break
			}
		}
		if !hasEnumConstraint {
			t.Error("theme should have enum constraint")
		}
	}
}

// TestSchemaConstraints tests that constraints are properly defined
func TestSchemaConstraints(t *testing.T) {
	schema := GetConfigSchema()

	tests := []struct {
		name            string
		settingPath     string
		constraintType  string
		wantConstraints bool
	}{
		{name: "logLevel has enum constraint", settingPath: "logLevel", wantConstraints: true, constraintType: "enum"},
		{name: "theme has enum constraint", settingPath: "theme", wantConstraints: true, constraintType: "enum"},
		{name: "colorScheme.border has hexcolor constraint", settingPath: "colorScheme.border", wantConstraints: true, constraintType: "hexcolor"},
		{name: "maxConcurrentOps has range constraint", settingPath: "maxConcurrentOps", wantConstraints: true, constraintType: "range"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting, exists := schema.Settings[tt.settingPath]
			if !exists {
				t.Fatalf("Setting %q not found", tt.settingPath)
			}

			if tt.wantConstraints {
				if len(setting.Constraints) == 0 {
					t.Errorf("Setting %q should have constraints", tt.settingPath)
				}

				// Check for specific constraint type
				found := false
				for _, c := range setting.Constraints {
					if c.Type == tt.constraintType {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Setting %q should have %s constraint", tt.settingPath, tt.constraintType)
				}
			}
		})
	}
}
