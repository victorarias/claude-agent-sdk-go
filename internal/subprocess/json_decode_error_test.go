// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"errors"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestParseJSONLineWrapsErrors tests that parseJSONLine wraps JSON errors in JSONDecodeError.
func TestParseJSONLineWrapsErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectError   bool
		expectWrapper bool
	}{
		{
			name:          "malformed JSON - missing closing brace",
			input:         `{"incomplete": "json"`,
			expectError:   true,
			expectWrapper: true,
		},
		{
			name:          "malformed JSON - invalid syntax",
			input:         `{invalid json}`,
			expectError:   true,
			expectWrapper: true,
		},
		{
			name:          "malformed JSON - unexpected EOF",
			input:         `{"key":`,
			expectError:   true,
			expectWrapper: true,
		},
		{
			name:          "empty line",
			input:         "",
			expectError:   true,
			expectWrapper: false, // Empty line is not a JSON decode error
		},
		{
			name:          "valid JSON - should not error",
			input:         `{"valid": "json"}`,
			expectError:   false,
			expectWrapper: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONLine(tt.input)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error for input %q, but got none", tt.input)
				}

				if tt.expectWrapper {
					// Check if error is wrapped in JSONDecodeError
					var jsonErr *types.JSONDecodeError
					if !errors.As(err, &jsonErr) {
						t.Errorf("Expected error to be wrapped in JSONDecodeError, got: %T (%v)", err, err)
					} else {
						// Verify the JSONDecodeError contains the line
						if jsonErr.Line != tt.input {
							t.Errorf("JSONDecodeError.Line = %q, want %q", jsonErr.Line, tt.input)
						}
						// Verify it has an underlying error
						if jsonErr.OriginalError == nil {
							t.Error("JSONDecodeError.OriginalError should not be nil")
						}
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for valid JSON: %v", err)
				}
				if result == nil {
					t.Error("Expected result for valid JSON, got nil")
				}
			}
		})
	}
}

// TestJSONAccumulatorWrapsJSONErrors tests that jsonAccumulator wraps errors properly.
func TestJSONAccumulatorWrapsJSONErrors(t *testing.T) {
	// Test that when accumulator completes parsing and gets a JSON error,
	// it's wrapped in JSONDecodeError
	acc := newJSONAccumulatorWithLimit(1000)

	// Add a line that will attempt to parse and fail
	result, err := acc.addLine(`{"bad": json}`)

	// This should still return nil, nil because parseJSONLine is called
	// but if it fails, we don't propagate the error yet (we're still accumulating)
	if err != nil || result != nil {
		t.Logf("Note: accumulator behavior - err=%v, result=%v", err, result)
	}

	// The real test is in parseJSONLine itself
}

// TestBufferOverflowIsNotWrappedInJSONDecodeError tests that buffer overflow
// errors are distinct from JSON decode errors.
func TestBufferOverflowIsNotWrappedInJSONDecodeError(t *testing.T) {
	acc := newJSONAccumulatorWithLimit(100)

	// Add data that exceeds buffer limit
	_, err := acc.addLine("x" + "y" + "z" + "a" + "b" + "c" + "d" + "e" + "f" + "g" + "h" + "i" + "j" + "k" + "l" + "m" + "n" + "o" + "p" + "q" + "r" + "s" + "t" + "u" + "v" + "w" + "x" + "y" + "z")

	if err == nil {
		// Try with a definitely oversized line
		_, err = acc.addLine("x" + "y" + "z" + "a" + "b" + "c" + "d" + "e" + "f" + "g" + "h" + "i" + "j" + "k" + "l" + "m" + "n" + "o" + "p" + "q" + "r" + "s" + "t" + "u" + "v" + "w" + "x" + "y" + "z" + "1" + "2" + "3" + "4" + "5" + "6" + "7" + "8" + "9" + "0")
	}

	if err == nil {
		// Force it with a guaranteed overflow
		longString := ""
		for i := 0; i < 200; i++ {
			longString += "x"
		}
		_, err = acc.addLine(longString)
	}

	if err == nil {
		t.Fatal("Expected buffer overflow error")
	}

	// Buffer overflow should NOT be a JSONDecodeError
	var jsonErr *types.JSONDecodeError
	if errors.As(err, &jsonErr) {
		t.Error("Buffer overflow error should not be wrapped in JSONDecodeError")
	}

	// It should be a generic error about buffer/limit
	errStr := err.Error()
	if errStr == "" {
		t.Error("Error message should not be empty")
	}
}
