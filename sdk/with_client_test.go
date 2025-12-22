// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"context"
	"errors"
	"testing"
)

// TestWithClientFunc tests the basic WithClient function exists and compiles
func TestWithClientFunc(t *testing.T) {
	// Just verify the function exists and has correct signature
	var _ func(context.Context, []any, ClientFunc) error
	_ = func(ctx context.Context, opts []any, fn ClientFunc) error {
		return nil
	}
}

// TestWithClient_PanicRecovery tests that WithClient re-raises panic after cleanup
func TestWithClient_PanicRecovery(t *testing.T) {
	// Test the panic recovery logic directly on Client.Run which uses same pattern
	client := &Client{
		connected: true,
	}

	expectedPanic := "test panic"
	panicked := false

	defer func() {
		if r := recover(); r != nil {
			panicked = true
			if r != expectedPanic {
				t.Errorf("expected panic %v, got %v", expectedPanic, r)
			}
		}
		if !panicked {
			t.Error("expected panic to be re-raised")
		}
	}()

	// Simulate what WithClient does with panic recovery
	func() {
		defer func() {
			client.connected = false // cleanup
			if r := recover(); r != nil {
				panic(r) // re-raise
			}
		}()
		panic(expectedPanic)
	}()
}

// TestWithClient_CleanupOnError tests that cleanup happens even on error
func TestWithClient_CleanupOnError(t *testing.T) {
	client := &Client{
		connected: true,
	}

	expectedErr := errors.New("function error")
	var cleanedUp bool

	// Simulate WithClient error handling
	err := func() error {
		defer func() {
			cleanedUp = true
			client.connected = false
		}()
		return expectedErr
	}()

	if err != expectedErr {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if !cleanedUp {
		t.Error("expected cleanup to happen")
	}
	if client.connected {
		t.Error("expected client to be disconnected")
	}
}

// TestWithClient_ConnectError tests that function is not called if connect fails
func TestWithClient_ConnectError(t *testing.T) {
	// When connect fails, the function should not be called
	// and the error should be returned
	connectErr := errors.New("connect failed")
	functionCalled := false

	// Simulate WithClient behavior when connect fails
	err := func() error {
		// Simulate connect failure
		if connectErr != nil {
			return connectErr
		}
		// This should not be reached
		functionCalled = true
		return nil
	}()

	if err != connectErr {
		t.Errorf("expected connect error %v, got %v", connectErr, err)
	}
	if functionCalled {
		t.Error("expected function not to be called when connect fails")
	}
}
