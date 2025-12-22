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
