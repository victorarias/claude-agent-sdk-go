package types

import (
	"encoding/json"
	"testing"
)

// TestHookOutputPreToolUseSpecific tests hook-specific output for PreToolUse events.
func TestHookOutputPreToolUseSpecific(t *testing.T) {
	tests := []struct {
		name     string
		output   *HookOutput
		expected string
	}{
		{
			name: "PreToolUse with permission decision allow",
			output: &HookOutput{
				HookSpecific: map[string]any{
					"hookEventName":       "PreToolUse",
					"permissionDecision":  "allow",
				},
			},
			expected: `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow"}}`,
		},
		{
			name: "PreToolUse with permission decision deny and reason",
			output: &HookOutput{
				HookSpecific: map[string]any{
					"hookEventName":              "PreToolUse",
					"permissionDecision":         "deny",
					"permissionDecisionReason":   "Unsafe operation",
				},
			},
			expected: `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":"Unsafe operation"}}`,
		},
		{
			name: "PreToolUse with updated input",
			output: &HookOutput{
				HookSpecific: map[string]any{
					"hookEventName":       "PreToolUse",
					"permissionDecision":  "allow",
					"updatedInput": map[string]any{
						"modified_param": "new_value",
					},
				},
			},
			expected: `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"modified_param":"new_value"}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.output)
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

// TestHookOutputPostToolUseSpecific tests hook-specific output for PostToolUse events.
func TestHookOutputPostToolUseSpecific(t *testing.T) {
	tests := []struct {
		name     string
		output   *HookOutput
		expected string
	}{
		{
			name: "PostToolUse with additional context",
			output: &HookOutput{
				HookSpecific: map[string]any{
					"hookEventName":     "PostToolUse",
					"additionalContext": "Tool executed successfully",
				},
			},
			expected: `{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"Tool executed successfully"}}`,
		},
		{
			name: "PostToolUse without additional context",
			output: &HookOutput{
				HookSpecific: map[string]any{
					"hookEventName": "PostToolUse",
				},
			},
			expected: `{"hookSpecificOutput":{"hookEventName":"PostToolUse"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.output)
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

// TestHookOutputUserPromptSubmitSpecific tests hook-specific output for UserPromptSubmit events.
func TestHookOutputUserPromptSubmitSpecific(t *testing.T) {
	output := &HookOutput{
		HookSpecific: map[string]any{
			"hookEventName":     "UserPromptSubmit",
			"additionalContext": "Prompt validated",
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal HookOutput: %v", err)
	}

	expected := `{"hookSpecificOutput":{"hookEventName":"UserPromptSubmit","additionalContext":"Prompt validated"}}`

	var got, want map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatalf("failed to unmarshal expected: %v", err)
	}

	if !deepMapsEqual(got, want) {
		t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", string(data), expected)
	}
}

// TestHookOutputWithBothCommonAndSpecificFields tests combining common hook output
// fields with hook-specific output fields.
func TestHookOutputWithBothCommonAndSpecificFields(t *testing.T) {
	output := &HookOutput{
		Continue:       boolPtr(true),
		SuppressOutput: true,
		Decision:       "block",
		SystemMessage:  "Operation blocked",
		HookSpecific: map[string]any{
			"hookEventName":       "PreToolUse",
			"permissionDecision":  "deny",
			"permissionDecisionReason": "Security policy violation",
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal HookOutput: %v", err)
	}

	// Unmarshal to verify all fields are present
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Check common fields
	if result["continue"] != true {
		t.Errorf("continue: got %v, want true", result["continue"])
	}
	if result["suppressOutput"] != true {
		t.Errorf("suppressOutput: got %v, want true", result["suppressOutput"])
	}
	if result["decision"] != "block" {
		t.Errorf("decision: got %v, want 'block'", result["decision"])
	}
	if result["systemMessage"] != "Operation blocked" {
		t.Errorf("systemMessage: got %v, want 'Operation blocked'", result["systemMessage"])
	}

	// Check hook-specific output
	hookSpecific, ok := result["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("hookSpecificOutput: not found or not a map")
	}
	if hookSpecific["hookEventName"] != "PreToolUse" {
		t.Errorf("hookEventName: got %v, want 'PreToolUse'", hookSpecific["hookEventName"])
	}
	if hookSpecific["permissionDecision"] != "deny" {
		t.Errorf("permissionDecision: got %v, want 'deny'", hookSpecific["permissionDecision"])
	}
	if hookSpecific["permissionDecisionReason"] != "Security policy violation" {
		t.Errorf("permissionDecisionReason: got %v, want 'Security policy violation'", hookSpecific["permissionDecisionReason"])
	}
}

// TestHookOutputSpecificUnmarshaling tests unmarshaling hook-specific output from JSON.
func TestHookOutputSpecificUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, output *HookOutput)
	}{
		{
			name:  "PreToolUse hook-specific",
			input: `{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"allow","updatedInput":{"key":"value"}}}`,
			validate: func(t *testing.T, output *HookOutput) {
				if output.HookSpecific == nil {
					t.Fatal("HookSpecific is nil")
				}
				if output.HookSpecific["hookEventName"] != "PreToolUse" {
					t.Errorf("hookEventName: got %v, want 'PreToolUse'", output.HookSpecific["hookEventName"])
				}
				if output.HookSpecific["permissionDecision"] != "allow" {
					t.Errorf("permissionDecision: got %v, want 'allow'", output.HookSpecific["permissionDecision"])
				}
				updatedInput, ok := output.HookSpecific["updatedInput"].(map[string]any)
				if !ok {
					t.Fatal("updatedInput not a map")
				}
				if updatedInput["key"] != "value" {
					t.Errorf("updatedInput[key]: got %v, want 'value'", updatedInput["key"])
				}
			},
		},
		{
			name:  "PostToolUse hook-specific",
			input: `{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"Context message"}}`,
			validate: func(t *testing.T, output *HookOutput) {
				if output.HookSpecific == nil {
					t.Fatal("HookSpecific is nil")
				}
				if output.HookSpecific["hookEventName"] != "PostToolUse" {
					t.Errorf("hookEventName: got %v, want 'PostToolUse'", output.HookSpecific["hookEventName"])
				}
				if output.HookSpecific["additionalContext"] != "Context message" {
					t.Errorf("additionalContext: got %v, want 'Context message'", output.HookSpecific["additionalContext"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output HookOutput
			if err := json.Unmarshal([]byte(tt.input), &output); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			tt.validate(t, &output)
		})
	}
}

// TestNewPreToolUseOutput tests the helper function for creating PreToolUse hook outputs.
func TestNewPreToolUseOutput(t *testing.T) {
	tests := []struct {
		name              string
		decision          string
		reason            string
		updatedInput      map[string]any
		wantDecision      string
		wantReason        bool
		wantUpdatedInput  bool
	}{
		{
			name:         "allow without reason or updated input",
			decision:     "allow",
			wantDecision: "allow",
			wantReason:   false,
			wantUpdatedInput: false,
		},
		{
			name:         "deny with reason",
			decision:     "deny",
			reason:       "Security violation",
			wantDecision: "deny",
			wantReason:   true,
			wantUpdatedInput: false,
		},
		{
			name:         "allow with updated input",
			decision:     "allow",
			updatedInput: map[string]any{"param": "modified"},
			wantDecision: "allow",
			wantReason:   false,
			wantUpdatedInput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output *HookOutput
			if tt.updatedInput != nil {
				output = NewPreToolUseOutput(tt.decision, tt.reason, tt.updatedInput)
			} else {
				output = NewPreToolUseOutput(tt.decision, tt.reason, nil)
			}

			if output.HookSpecific == nil {
				t.Fatal("HookSpecific is nil")
			}

			if output.HookSpecific["hookEventName"] != "PreToolUse" {
				t.Errorf("hookEventName: got %v, want 'PreToolUse'", output.HookSpecific["hookEventName"])
			}

			if output.HookSpecific["permissionDecision"] != tt.wantDecision {
				t.Errorf("permissionDecision: got %v, want %v", output.HookSpecific["permissionDecision"], tt.wantDecision)
			}

			if tt.wantReason {
				if _, ok := output.HookSpecific["permissionDecisionReason"]; !ok {
					t.Error("expected permissionDecisionReason to be present")
				}
				if output.HookSpecific["permissionDecisionReason"] != tt.reason {
					t.Errorf("permissionDecisionReason: got %v, want %v", output.HookSpecific["permissionDecisionReason"], tt.reason)
				}
			}

			if tt.wantUpdatedInput {
				if _, ok := output.HookSpecific["updatedInput"]; !ok {
					t.Error("expected updatedInput to be present")
				}
			}
		})
	}
}

// TestNewPostToolUseOutput tests the helper function for creating PostToolUse hook outputs.
func TestNewPostToolUseOutput(t *testing.T) {
	tests := []struct {
		name               string
		additionalContext  string
		wantContext        bool
	}{
		{
			name:        "with additional context",
			additionalContext: "Tool completed successfully",
			wantContext: true,
		},
		{
			name:        "without additional context",
			additionalContext: "",
			wantContext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := NewPostToolUseOutput(tt.additionalContext)

			if output.HookSpecific == nil {
				t.Fatal("HookSpecific is nil")
			}

			if output.HookSpecific["hookEventName"] != "PostToolUse" {
				t.Errorf("hookEventName: got %v, want 'PostToolUse'", output.HookSpecific["hookEventName"])
			}

			if tt.wantContext {
				if _, ok := output.HookSpecific["additionalContext"]; !ok {
					t.Error("expected additionalContext to be present")
				}
				if output.HookSpecific["additionalContext"] != tt.additionalContext {
					t.Errorf("additionalContext: got %v, want %v", output.HookSpecific["additionalContext"], tt.additionalContext)
				}
			}
		})
	}
}

// Helper function for deep map comparison including nested maps
func deepMapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if !deepEqual(va, vb) {
			return false
		}
	}
	return true
}

func deepEqual(a, b any) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle map types
	mapA, okA := a.(map[string]any)
	mapB, okB := b.(map[string]any)
	if okA && okB {
		return deepMapsEqual(mapA, mapB)
	}
	if okA != okB {
		return false
	}

	// For other types, use direct comparison
	return a == b
}
