// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"context"
	"encoding/json"
	"runtime"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestQueryStream_NoGoroutineLeak verifies that QueryStream does NOT leak goroutines
// when the caller abandons the returned channels without consuming them.
//
// The goroutine should exit cleanly even when:
// 1. The caller doesn't consume from msgChan/errChan
// 2. The transport continues to produce messages
// 3. The msgChan buffer fills up
//
// This requires proper context-based cancellation to ensure cleanup.
func TestQueryStream_NoGoroutineLeak(t *testing.T) {
	// Force GC to clean up any existing goroutines
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// Count goroutines before
	beforeCount := runtime.NumGoroutine()

	// Create a transport and pre-fill it with messages
	// MockTransport has a buffer of 100, so we fill it partially
	transport := NewMockTransport()

	// Pre-fill with messages (less than buffer size to avoid blocking)
	for i := 0; i < 50; i++ {
		transport.SendMessage(map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		})
	}

	// Use a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Respond to initialize request so QueryStream can start.
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			written := transport.Written()
			if len(written) == 0 {
				continue
			}
			var req map[string]any
			if err := json.Unmarshal([]byte(written[len(written)-1]), &req); err != nil {
				continue
			}
			if req["type"] != "control_request" {
				continue
			}
			reqID, _ := req["request_id"].(string)
			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "leak_test"},
				},
			})
			return
		}
	}()

	// Call QueryStream but ABANDON the channels without consuming
	// This simulates a caller that starts the stream but doesn't read from it
	msgChan, errChan := QueryStream(ctx, "Hello", types.WithTransport(transport))

	// Explicitly abandon the channels - don't consume them
	_ = msgChan
	_ = errChan

	// Give the goroutine time to process and potentially block
	time.Sleep(100 * time.Millisecond)

	// Now cancel the context to signal abandonment
	cancel()

	// Give time for the goroutine to exit after cancellation
	time.Sleep(100 * time.Millisecond)

	// Force GC again
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// Count goroutines after
	afterCount := runtime.NumGoroutine()

	// With proper cleanup, no goroutines should leak
	// We allow small margin for test infrastructure
	leaked := afterCount - beforeCount

	if leaked > 1 {
		t.Errorf("Goroutine leak detected: count went from %d to %d (diff: %d). "+
			"QueryStream goroutine should exit when caller cancels context.",
			beforeCount, afterCount, leaked)
	}

	t.Logf("Goroutines: before=%d, after=%d, leaked=%d",
		beforeCount, afterCount, leaked)
}

// TestQueryStream_WithContextCancellation tests that QueryStream properly
// cleans up goroutines when the context is canceled.
func TestQueryStream_WithContextCancellation(t *testing.T) {
	// Force GC to clean up any existing goroutines
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// Count goroutines before
	beforeCount := runtime.NumGoroutine()

	// Create a transport and pre-fill with messages
	transport := NewMockTransport()

	// Pre-fill with some messages
	for i := 0; i < 50; i++ {
		transport.SendMessage(map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		})
	}

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Respond to initialize request so QueryStream can start.
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			written := transport.Written()
			if len(written) == 0 {
				continue
			}
			var req map[string]any
			if err := json.Unmarshal([]byte(written[len(written)-1]), &req); err != nil {
				continue
			}
			if req["type"] != "control_request" {
				continue
			}
			reqID, _ := req["request_id"].(string)
			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "leak_test"},
				},
			})
			return
		}
	}()

	// Call QueryStream
	msgChan, errChan := QueryStream(ctx, "Hello", types.WithTransport(transport))

	// Abandon the channels
	_ = msgChan
	_ = errChan

	// Give the goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Cancel the context - this should stop the goroutine
	cancel()

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Force GC
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// Count goroutines after
	afterCount := runtime.NumGoroutine()

	// With proper context cancellation, the goroutine should exit
	// We allow a small margin (1-2 goroutines) for test infrastructure
	leaked := afterCount - beforeCount

	if leaked > 2 {
		t.Errorf("Expected minimal goroutine leak with context cancellation, but count went from %d to %d (diff: %d)",
			beforeCount, afterCount, leaked)
	}

	t.Logf("With context cancellation: before=%d, after=%d, leaked=%d",
		beforeCount, afterCount, leaked)
}
