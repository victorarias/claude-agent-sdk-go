// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

// TestNewStopOutput tests the helper function for creating Stop hook outputs.
func TestNewStopOutput(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "basic stop output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewStopOutput()

			if output.HookSpecific == nil {
				t.Fatal("HookSpecific is nil")
			}

			if output.HookSpecific["hookEventName"] != "Stop" {
				t.Errorf("hookEventName: got %v, want 'Stop'", output.HookSpecific["hookEventName"])
			}
		})
	}
}

// TestNewStopOutputJSON tests that NewStopOutput produces correct JSON serialization.
func TestNewStopOutputJSON(t *testing.T) {
	output := NewStopOutput()

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal HookOutput: %v", err)
	}

	expected := `{"hookSpecificOutput":{"hookEventName":"Stop"}}`

	// Parse both to compare as maps to avoid field ordering issues
	var got, want map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("failed to unmarshal expected: %v", err)
	}

	// Compare nested structures
	if !deepMapsEqual(got, want) {
		t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", string(data), expected)
	}
}

// TestNewSubagentStopOutput tests the helper function for creating SubagentStop hook outputs.
func TestNewSubagentStopOutput(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "basic subagent stop output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewSubagentStopOutput()

			if output.HookSpecific == nil {
				t.Fatal("HookSpecific is nil")
			}

			if output.HookSpecific["hookEventName"] != "SubagentStop" {
				t.Errorf("hookEventName: got %v, want 'SubagentStop'", output.HookSpecific["hookEventName"])
			}
		})
	}
}

// TestNewSubagentStopOutputJSON tests that NewSubagentStopOutput produces correct JSON serialization.
func TestNewSubagentStopOutputJSON(t *testing.T) {
	output := NewSubagentStopOutput()

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal HookOutput: %v", err)
	}

	expected := `{"hookSpecificOutput":{"hookEventName":"SubagentStop"}}`

	// Parse both to compare as maps to avoid field ordering issues
	var got, want map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("failed to unmarshal expected: %v", err)
	}

	// Compare nested structures
	if !deepMapsEqual(got, want) {
		t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", string(data), expected)
	}
}

// TestNewPreCompactOutput tests the helper function for creating PreCompact hook outputs.
func TestNewPreCompactOutput(t *testing.T) {
	tests := []struct {
		name               string
		customInstructions string
		wantInstructions   bool
	}{
		{
			name:               "with custom instructions",
			customInstructions: "Please focus on performance optimization",
			wantInstructions:   true,
		},
		{
			name:               "without custom instructions",
			customInstructions: "",
			wantInstructions:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewPreCompactOutput(tt.customInstructions)

			if output.HookSpecific == nil {
				t.Fatal("HookSpecific is nil")
			}

			if output.HookSpecific["hookEventName"] != "PreCompact" {
				t.Errorf("hookEventName: got %v, want 'PreCompact'", output.HookSpecific["hookEventName"])
			}

			if tt.wantInstructions {
				if _, ok := output.HookSpecific["customInstructions"]; !ok {
					t.Error("expected customInstructions to be present")
				}
				if output.HookSpecific["customInstructions"] != tt.customInstructions {
					t.Errorf("customInstructions: got %v, want %v", output.HookSpecific["customInstructions"], tt.customInstructions)
				}
			} else {
				if _, ok := output.HookSpecific["customInstructions"]; ok {
					t.Error("expected customInstructions to be absent")
				}
			}
		})
	}
}

// TestNewPreCompactOutputJSON tests that NewPreCompactOutput produces correct JSON serialization.
func TestNewPreCompactOutputJSON(t *testing.T) {
	tests := []struct {
		name               string
		customInstructions string
		expected           string
	}{
		{
			name:               "with custom instructions",
			customInstructions: "Focus on performance",
			expected:           `{"hookSpecificOutput":{"hookEventName":"PreCompact","customInstructions":"Focus on performance"}}`,
		},
		{
			name:               "without custom instructions",
			customInstructions: "",
			expected:           `{"hookSpecificOutput":{"hookEventName":"PreCompact"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewPreCompactOutput(tt.customInstructions)

			data, err := json.Marshal(output)
			if err != nil {
				t.Fatalf("failed to marshal HookOutput: %v", err)
			}

			// Parse both to compare as maps to avoid field ordering issues
			var got, want map[string]any
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.expected), &want); err != nil {
				t.Fatalf("failed to unmarshal expected: %v", err)
			}

			// Compare nested structures
			if !deepMapsEqual(got, want) {
				t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", string(data), tt.expected)
			}
		})
	}
}

// Helper functions (deepMapsEqual and deepEqual are already defined in hooks_specific_test.go)
