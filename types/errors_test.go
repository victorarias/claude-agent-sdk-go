// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"errors"
	"testing"
	"time"
)

// TestTimeoutError tests the TimeoutError type.
func TestTimeoutError(t *testing.T) {
	t.Run("Error message format", func(t *testing.T) {
		err := &TimeoutError{
			Operation: "reading response",
			Duration:  5 * time.Second,
		}

		expected := "timeout after 5s: reading response"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrTimeout", func(t *testing.T) {
		err := &TimeoutError{
			Operation: "connecting",
			Duration:  10 * time.Second,
		}

		if !errors.Is(err, ErrTimeout) {
			t.Error("Expected errors.Is(err, ErrTimeout) to be true")
		}
	})

	t.Run("errors.As works correctly", func(t *testing.T) {
		err := &TimeoutError{
			Operation: "writing data",
			Duration:  2 * time.Second,
		}

		var timeoutErr *TimeoutError
		if !errors.As(err, &timeoutErr) {
			t.Fatal("Expected errors.As to return true")
		}

		if timeoutErr.Operation != "writing data" {
			t.Errorf("Expected operation 'writing data', got %q", timeoutErr.Operation)
		}
		if timeoutErr.Duration != 2*time.Second {
			t.Errorf("Expected duration 2s, got %v", timeoutErr.Duration)
		}
	})

	t.Run("wrapped error still matches ErrTimeout", func(t *testing.T) {
		timeoutErr := &TimeoutError{
			Operation: "query",
			Duration:  30 * time.Second,
		}
		wrappedErr := &SDKError{
			Message: "operation failed",
			Cause:   timeoutErr,
		}

		if !errors.Is(wrappedErr, ErrTimeout) {
			t.Error("Expected wrapped error to match ErrTimeout")
		}
	})
}

// TestClosedError tests the ClosedError type.
func TestClosedError(t *testing.T) {
	t.Run("Error message format", func(t *testing.T) {
		err := &ClosedError{
			Resource: "transport connection",
		}

		expected := "resource closed: transport connection"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrClosed", func(t *testing.T) {
		err := &ClosedError{
			Resource: "CLI process",
		}

		if !errors.Is(err, ErrClosed) {
			t.Error("Expected errors.Is(err, ErrClosed) to be true")
		}
	})

	t.Run("errors.As works correctly", func(t *testing.T) {
		err := &ClosedError{
			Resource: "message channel",
		}

		var closedErr *ClosedError
		if !errors.As(err, &closedErr) {
			t.Fatal("Expected errors.As to return true")
		}

		if closedErr.Resource != "message channel" {
			t.Errorf("Expected resource 'message channel', got %q", closedErr.Resource)
		}
	})

	t.Run("wrapped error still matches ErrClosed", func(t *testing.T) {
		closedErr := &ClosedError{
			Resource: "session",
		}
		wrappedErr := &SDKError{
			Message: "operation on closed resource",
			Cause:   closedErr,
		}

		if !errors.Is(wrappedErr, ErrClosed) {
			t.Error("Expected wrapped error to match ErrClosed")
		}
	})
}

// TestSDKError tests the SDKError type.
func TestSDKError(t *testing.T) {
	t.Run("Error message without cause", func(t *testing.T) {
		err := &SDKError{
			Message: "operation failed",
		}

		expected := "sdk: operation failed"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Error message with cause", func(t *testing.T) {
		cause := errors.New("underlying error")
		err := &SDKError{
			Message: "operation failed",
			Cause:   cause,
		}

		expected := "sdk: operation failed: underlying error"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})
}

// TestCLINotFoundError tests the CLINotFoundError type.
func TestCLINotFoundError(t *testing.T) {
	t.Run("Error with explicit path", func(t *testing.T) {
		err := &CLINotFoundError{
			CLIPath: "/usr/local/bin/claude",
		}

		expected := "claude CLI not found at: /usr/local/bin/claude"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Error with searched paths", func(t *testing.T) {
		err := &CLINotFoundError{
			SearchedPaths: []string{"/usr/bin", "/usr/local/bin", "/opt/bin"},
		}

		expected := "claude CLI not found, searched: /usr/bin, /usr/local/bin, /opt/bin"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrCLINotFound", func(t *testing.T) {
		err := &CLINotFoundError{
			CLIPath: "/usr/bin/claude",
		}

		if !errors.Is(err, ErrCLINotFound) {
			t.Error("Expected errors.Is(err, ErrCLINotFound) to be true")
		}
	})
}

// TestConnectionError tests the ConnectionError type.
func TestConnectionError(t *testing.T) {
	t.Run("Error message without cause", func(t *testing.T) {
		err := &ConnectionError{
			Message: "failed to connect",
		}

		expected := "connection error: failed to connect"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Error message with cause", func(t *testing.T) {
		cause := errors.New("network unreachable")
		err := &ConnectionError{
			Message: "failed to connect",
			Cause:   cause,
		}

		expected := "connection error: failed to connect: network unreachable"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrConnection", func(t *testing.T) {
		err := &ConnectionError{
			Message: "timeout",
		}

		if !errors.Is(err, ErrConnection) {
			t.Error("Expected errors.Is(err, ErrConnection) to be true")
		}
	})

	t.Run("Unwrap returns cause", func(t *testing.T) {
		cause := errors.New("network unreachable")
		err := &ConnectionError{
			Message: "failed to connect",
			Cause:   cause,
		}

		if err.Unwrap() != cause {
			t.Error("Expected Unwrap to return the cause")
		}
	})
}

// TestProcessError tests the ProcessError type.
func TestProcessError(t *testing.T) {
	t.Run("Error message format", func(t *testing.T) {
		err := &ProcessError{
			ExitCode: 1,
			Stderr:   "command not found",
		}

		expected := "process exited with code 1: command not found"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrProcess", func(t *testing.T) {
		err := &ProcessError{
			ExitCode: 127,
			Stderr:   "file not found",
		}

		if !errors.Is(err, ErrProcess) {
			t.Error("Expected errors.Is(err, ErrProcess) to be true")
		}
	})
}

// TestJSONDecodeError tests the JSONDecodeError type.
func TestJSONDecodeError(t *testing.T) {
	t.Run("Error message format", func(t *testing.T) {
		originalErr := errors.New("unexpected end of JSON input")
		err := &JSONDecodeError{
			Line:          `{"incomplete": `,
			OriginalError: originalErr,
		}

		expected := `JSON decode error on line "{\"incomplete\": ": unexpected end of JSON input`
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrParse", func(t *testing.T) {
		err := &JSONDecodeError{
			Line:          "invalid json",
			OriginalError: errors.New("parse error"),
		}

		if !errors.Is(err, ErrParse) {
			t.Error("Expected errors.Is(err, ErrParse) to be true")
		}
	})

	t.Run("Unwrap returns original error", func(t *testing.T) {
		originalErr := errors.New("syntax error")
		err := &JSONDecodeError{
			Line:          "bad json",
			OriginalError: originalErr,
		}

		if err.Unwrap() != originalErr {
			t.Error("Expected Unwrap to return the original error")
		}
	})
}

// TestMessageParseError tests the MessageParseError type.
func TestMessageParseError(t *testing.T) {
	t.Run("Error message format", func(t *testing.T) {
		err := &MessageParseError{
			Message: "missing required field 'type'",
			Data:    map[string]any{"content": "test"},
		}

		expected := "message parse error: missing required field 'type'"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrParse", func(t *testing.T) {
		err := &MessageParseError{
			Message: "invalid message structure",
		}

		if !errors.Is(err, ErrParse) {
			t.Error("Expected errors.Is(err, ErrParse) to be true")
		}
	})
}

// TestCLIVersionError tests the CLIVersionError type.
func TestCLIVersionError(t *testing.T) {
	t.Run("Error message format", func(t *testing.T) {
		err := &CLIVersionError{
			InstalledVersion: "1.0.0",
			MinimumVersion:   "2.0.0",
		}

		expected := "CLI version 1.0.0 is below minimum required version 2.0.0"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.Is returns true for ErrCLIVersion", func(t *testing.T) {
		err := &CLIVersionError{
			InstalledVersion: "0.5.0",
			MinimumVersion:   "1.0.0",
		}

		if !errors.Is(err, ErrCLIVersion) {
			t.Error("Expected errors.Is(err, ErrCLIVersion) to be true")
		}
	})
}
