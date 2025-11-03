package config

import (
	"context"
	"testing"
)

// TestValidateKeybindingConflicts tests keybinding conflict detection
func TestValidateKeybindingConflicts(t *testing.T) {
	tests := []struct {
		keybindings map[string]KeyBinding
		name        string
		wantErrors  int
	}{
		{
			name: "no conflicts",
			keybindings: map[string]KeyBinding{
				"quit":    {Action: "quit", Key: "q", Context: "global"},
				"help":    {Action: "help", Key: "?", Context: "global"},
				"refresh": {Action: "refresh", Key: "r", Context: "global"},
			},
			wantErrors: 0,
		},
		{
			name: "conflict in same context",
			keybindings: map[string]KeyBinding{
				"quit":   {Action: "quit", Key: "q", Context: "global"},
				"custom": {Action: "custom", Key: "q", Context: "global"}, // Conflict!
			},
			wantErrors: 1,
		},
		{
			name: "no conflict in different contexts",
			keybindings: map[string]KeyBinding{
				"quit":       {Action: "quit", Key: "q", Context: "global"},
				"quick_edit": {Action: "quick_edit", Key: "q", Context: "editor"},
			},
			wantErrors: 0,
		},
		{
			name: "multiple conflicts",
			keybindings: map[string]KeyBinding{
				"action1": {Action: "action1", Key: "a", Context: "global"},
				"action2": {Action: "action2", Key: "a", Context: "global"}, // Conflict!
				"action3": {Action: "action3", Key: "b", Context: "global"},
				"action4": {Action: "action4", Key: "b", Context: "global"}, // Conflict!
			},
			wantErrors: 2,
		},
		{
			name: "triple assignment to same key",
			keybindings: map[string]KeyBinding{
				"first":  {Action: "first", Key: "x", Context: "test"},
				"second": {Action: "second", Key: "x", Context: "test"}, // Conflict!
				"third":  {Action: "third", Key: "x", Context: "test"},  // Conflict!
			},
			wantErrors: 2, // Two conflicts (second and third conflict with first)
		},
		{
			name:        "empty keybindings",
			keybindings: map[string]KeyBinding{},
			wantErrors:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Keybindings: tt.keybindings,
			}

			schema := GetConfigSchema()
			v := newValidator(schema)
			errors := v.validateKeybindingConflicts(cfg)

			if len(errors) != tt.wantErrors {
				t.Errorf("validateKeybindingConflicts() got %d errors, want %d", len(errors), tt.wantErrors)
				for i, err := range errors {
					t.Logf("Error %d: %s", i, err.Error())
				}
			}

			// Verify error details for conflicts
			if tt.wantErrors > 0 {
				for _, err := range errors {
					if err.Severity != "warning" {
						t.Errorf("Keybinding conflict should be a warning, got: %s", err.Severity)
					}
					if err.DefaultUsed != "conflicting binding ignored" {
						t.Errorf("Expected DefaultUsed='conflicting binding ignored', got: %v", err.DefaultUsed)
					}
				}
			}
		})
	}
}

// TestValidateKeybindingConflictsIntegration tests full validation with keybinding conflicts
func TestValidateKeybindingConflictsIntegration(t *testing.T) {
	loader := NewLoader()

	cfg := &Config{
		LogLevel: "info",
		Keybindings: map[string]KeyBinding{
			"quit":   {Action: "quit", Key: "q", Context: "global"},
			"custom": {Action: "custom", Key: "q", Context: "global"}, // Conflict!
		},
	}

	ctx := context.Background()
	errors, err := loader.Validate(ctx, cfg)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	// Should have at least one keybinding conflict warning
	hasKeybindingError := false
	for _, verr := range errors {
		if verr.Key == "keybindings.custom" || verr.Key == "keybindings.quit" {
			hasKeybindingError = true
			if verr.Severity != "warning" {
				t.Errorf("Keybinding conflict should be warning, got: %s", verr.Severity)
			}
		}
	}

	if !hasKeybindingError {
		t.Error("Expected keybinding conflict error in validation")
	}
}
