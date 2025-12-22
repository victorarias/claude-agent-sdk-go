// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

// TestHookOutputAsyncJSONMarshaling tests that async hook output fields
// are marshaled with correct JSON keys to match CLI expectations.
func TestHookOutputAsyncJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		output   *HookOutput
		expected string
	}{
		{
			name: "async with timeout",
			output: &HookOutput{
				Async:        true,
				AsyncTimeout: intPtr(5000),
			},
			expected: `{"async":true,"asyncTimeout":5000}`,
		},
		{
			name: "async without timeout",
			output: &HookOutput{
				Async: true,
			},
			expected: `{"async":true}`,
		},
		{
			name: "sync hook output (no async fields)",
			output: &HookOutput{
				Continue:   boolPtr(true),
				StopReason: "test",
			},
			expected: `{"continue":true,"stopReason":"test"}`,
		},
		{
			name: "async with other fields",
			output: &HookOutput{
				Async:          true,
				AsyncTimeout:   intPtr(3000),
				SuppressOutput: true,
				Decision:       "block",
			},
			expected: `{"suppressOutput":true,"decision":"block","async":true,"asyncTimeout":3000}`,
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

			// Compare field by field
			if !mapsEqual(got, want) {
				t.Errorf("JSON mismatch:\ngot:  %s\nwant: %s", string(data), tt.expected)
			}
		})
	}
}

// TestHookOutputAsyncJSONUnmarshaling tests that we can unmarshal async hook
// output from JSON with the CLI's expected format.
func TestHookOutputAsyncJSONUnmarshaling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *HookOutput
	}{
		{
			name:  "async with timeout",
			input: `{"async":true,"asyncTimeout":5000}`,
			expected: &HookOutput{
				Async:        true,
				AsyncTimeout: intPtr(5000),
			},
		},
		{
			name:  "async without timeout",
			input: `{"async":true}`,
			expected: &HookOutput{
				Async: true,
			},
		},
		{
			name:  "sync output",
			input: `{"continue":true,"stopReason":"test"}`,
			expected: &HookOutput{
				Continue:   boolPtr(true),
				StopReason: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got HookOutput
			if err := json.Unmarshal([]byte(tt.input), &got); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			// Compare fields
			if got.Async != tt.expected.Async {
				t.Errorf("Async: got %v, want %v", got.Async, tt.expected.Async)
			}

			if !intPtrEqual(got.AsyncTimeout, tt.expected.AsyncTimeout) {
				t.Errorf("AsyncTimeout: got %v, want %v", got.AsyncTimeout, tt.expected.AsyncTimeout)
			}

			if !boolPtrEqual(got.Continue, tt.expected.Continue) {
				t.Errorf("Continue: got %v, want %v", got.Continue, tt.expected.Continue)
			}

			if got.StopReason != tt.expected.StopReason {
				t.Errorf("StopReason: got %q, want %q", got.StopReason, tt.expected.StopReason)
			}
		})
	}
}

// TestAsyncHookOutputDistinctFromSync verifies that we can distinguish
// between async and sync hook outputs based on the Async field.
func TestAsyncHookOutputDistinctFromSync(t *testing.T) {
	asyncOutput := &HookOutput{
		Async:        true,
		AsyncTimeout: intPtr(1000),
	}

	syncOutput := &HookOutput{
		Continue:   boolPtr(true),
		StopReason: "done",
	}

	if !asyncOutput.Async {
		t.Error("asyncOutput.Async should be true")
	}

	if syncOutput.Async {
		t.Error("syncOutput.Async should be false")
	}

	// Marshal and verify JSON keys
	asyncJSON, _ := json.Marshal(asyncOutput)
	var asyncMap map[string]any
	json.Unmarshal(asyncJSON, &asyncMap)

	if _, ok := asyncMap["async"]; !ok {
		t.Error("async output should have 'async' key in JSON")
	}

	if _, ok := asyncMap["asyncTimeout"]; !ok {
		t.Error("async output should have 'asyncTimeout' key in JSON")
	}

	syncJSON, _ := json.Marshal(syncOutput)
	var syncMap map[string]any
	json.Unmarshal(syncJSON, &syncMap)

	if _, ok := syncMap["async"]; ok {
		t.Error("sync output should not have 'async' key in JSON (omitempty)")
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || v != bv {
			return false
		}
	}
	return true
}
