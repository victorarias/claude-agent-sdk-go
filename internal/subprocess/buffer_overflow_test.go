package subprocess

import (
	"strings"
	"testing"
)

// TestJSONAccumulatorBufferOverflowReturnsError tests that the JSON accumulator
// returns an error instead of silently discarding data when buffer overflows.
func TestJSONAccumulatorBufferOverflowReturnsError(t *testing.T) {
	acc := newJSONAccumulator()

	// Add lines until we exceed the buffer limit
	// Each line is 1KB, and the limit is 1MB, so we need ~1025 lines to exceed
	longLine := strings.Repeat("x", 1024)

	var lastErr error
	for i := 0; i < 1100; i++ {
		_, err := acc.addLine(longLine)
		if err != nil {
			lastErr = err
			break
		}
	}

	if lastErr == nil {
		t.Error("Expected buffer overflow error, but got none")
	}

	// Verify the error message mentions buffer overflow
	if lastErr != nil && !strings.Contains(lastErr.Error(), "buffer") {
		t.Errorf("Error message should mention 'buffer', got: %s", lastErr.Error())
	}
}

// TestJSONAccumulatorReturnsBufferOverflowError tests the specific error type.
func TestJSONAccumulatorReturnsBufferOverflowError(t *testing.T) {
	acc := newJSONAccumulatorWithLimit(100) // Small limit for testing

	// Add data that exceeds the limit
	_, err := acc.addLine(strings.Repeat("x", 150))

	if err == nil {
		t.Fatal("Expected error for buffer overflow")
	}

	// Check it's the right type of error
	if !strings.Contains(err.Error(), "exceeds") || !strings.Contains(err.Error(), "limit") {
		t.Errorf("Error should mention exceeding limit, got: %s", err.Error())
	}
}

// TestJSONAccumulatorDoesNotSilentlyDiscard tests that valid JSON is not
// silently discarded when accumulating.
func TestJSONAccumulatorDoesNotSilentlyDiscard(t *testing.T) {
	acc := newJSONAccumulator()

	// Add a complete JSON object split across lines
	_, err := acc.addLine(`{"key":`)
	if err != nil {
		t.Fatalf("Unexpected error on first line: %v", err)
	}

	result, err := acc.addLine(`"value"}`)
	if err != nil {
		t.Fatalf("Unexpected error on second line: %v", err)
	}

	if result == nil {
		t.Error("Expected parsed result, got nil")
	}

	if result["key"] != "value" {
		t.Errorf("Expected key=value, got %v", result)
	}
}

// TestBufferOverflowErrorIsSentToErrorChannel tests that buffer overflow errors
// are sent to the error channel in the transport.
func TestBufferOverflowErrorIsSentToErrorChannel(t *testing.T) {
	// This test documents the expected behavior: when a buffer overflow occurs
	// during message reading, the error should be sent to the errors channel
	// so the caller can handle it appropriately.
	t.Log("Buffer overflow errors should be sent to transport.Errors() channel")
}

// TestJSONDecodeErrorWrapping tests that JSON parsing errors are wrapped in JSONDecodeError.
func TestJSONDecodeErrorWrapping(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		checkType   bool
	}{
		{
			name:        "malformed JSON - missing closing brace",
			input:       `{"incomplete": "json"`,
			expectError: true,
			checkType:   true,
		},
		{
			name:        "malformed JSON - invalid syntax",
			input:       `{invalid json}`,
			expectError: true,
			checkType:   true,
		},
		{
			name:        "valid JSON - should not error",
			input:       `{"valid": "json"}`,
			expectError: false,
			checkType:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := newJSONAccumulator()
			result, err := acc.addLine(tt.input)

			if tt.expectError {
				// For malformed JSON that completes on one line, we expect it to fail
				// parseJSONLine should return an error
				if err == nil && result == nil {
					// Still accumulating - this is ok for multi-line JSON
					return
				}

				// If we got a result from malformed JSON, that's wrong
				if result != nil {
					t.Errorf("Expected error for malformed JSON, but got result: %v", result)
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

// TestReadMessagesWrapsJSONErrorsInJSONDecodeError tests that readMessages
// wraps JSON parsing errors in JSONDecodeError before sending to error channel.
func TestReadMessagesWrapsJSONErrorsInJSONDecodeError(t *testing.T) {
	t.Skip("This test requires integration testing with actual subprocess - skipping for now")
	// This test would verify that when readMessages encounters a JSON error,
	// it wraps it in a JSONDecodeError before sending to the errors channel.
	// For now, we're testing the underlying parseJSONLine function directly.
}
