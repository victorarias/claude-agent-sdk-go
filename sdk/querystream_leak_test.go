package sdk

import (
	"context"
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

	// Create a transport that will send MORE messages than the buffer can hold
	// QueryStream uses a buffered channel of size 100, so we send 150 messages
	transport := NewMockTransport()
	go func() {
		// Send MANY messages to exceed buffer capacity
		for i := 0; i < 150; i++ {
			transport.SendMessage(map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello!"},
					},
				},
			})
			time.Sleep(1 * time.Millisecond)
		}
		// Never send result - keep transport "alive"
	}()

	// Use a context with timeout to eventually clean up
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Call QueryStream but ABANDON the channels without consuming
	// This simulates a caller that starts the stream but doesn't read from it
	msgChan, errChan := QueryStream(ctx, "Hello", types.WithTransport(transport))

	// Explicitly abandon the channels - don't consume them
	_ = msgChan
	_ = errChan

	// Give the goroutine time to fill the buffer and block
	time.Sleep(300 * time.Millisecond)

	// Force GC again
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// Count goroutines after
	afterCount := runtime.NumGoroutine()

	// The goroutine should have leaked because:
	// 1. The QueryStream goroutine is trying to send to msgChan
	// 2. The msgChan buffer (size 100) is full
	// 3. Nobody is receiving from msgChan
	// 4. The goroutine is blocked forever on the send
	// 5. There's no mechanism to cancel the goroutine when caller abandons

	// With proper cleanup, no goroutines should leak
	// We expect exactly 0 leaked goroutines (the QueryStream goroutine should exit)
	leaked := afterCount - beforeCount

	if leaked > 0 {
		t.Errorf("Goroutine leak detected: count went from %d to %d (diff: %d). "+
			"QueryStream goroutine should exit when caller abandons channels, "+
			"but it's blocked trying to send to an abandoned msgChan.",
			beforeCount, afterCount, leaked)
		t.Logf("This indicates the goroutine is stuck at client.go:402 trying to send to msgChan")
	}

	t.Logf("Goroutines: before=%d, after=%d, leaked=%d",
		beforeCount, afterCount, leaked)
}

// TestQueryStream_WithContextCancellation tests that QueryStream properly
// cleans up goroutines when the context is cancelled.
func TestQueryStream_WithContextCancellation(t *testing.T) {
	// Force GC to clean up any existing goroutines
	runtime.GC()
	time.Sleep(10 * time.Millisecond)

	// Count goroutines before
	beforeCount := runtime.NumGoroutine()

	// Create a transport that will keep sending messages
	transport := NewMockTransport()
	go func() {
		// Send messages continuously
		for i := 0; i < 100; i++ {
			transport.SendMessage(map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello!"},
					},
				},
			})
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

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
