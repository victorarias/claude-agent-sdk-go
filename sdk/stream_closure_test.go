package sdk

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestStreamInputWaitsForResultWhenHooksActive tests that StreamInputWithWait waits for
// the first result message before returning when hooks are active.
func TestStreamInputWaitsForResultWhenHooksActive(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Register a hook to make hooks active
	query.hookCallbacks["test_hook"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		return nil, nil
	}

	// Start the query
	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Create input channel with a single message
	input := make(chan map[string]any, 1)
	input <- map[string]any{"type": "user", "message": "test"}
	close(input)

	// Track when StreamInputWithWait completes
	var streamInputDone atomic.Bool

	go func() {
		// StreamInputWithWait should wait for result before returning
		query.StreamInputWithWait(input)
		streamInputDone.Store(true)
	}()

	// Give StreamInputWithWait time to process the input message
	time.Sleep(50 * time.Millisecond)

	// StreamInputWithWait should NOT have completed yet (waiting for result)
	if streamInputDone.Load() {
		t.Fatal("StreamInputWithWait completed before result was received - should wait for result when hooks are active")
	}

	// Now send a result message
	transport.SendMessage(map[string]any{
		"type":       "result",
		"session_id": "test-session",
	})

	// Wait for StreamInputWithWait to complete
	deadline := time.Now().Add(2 * time.Second)
	for !streamInputDone.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !streamInputDone.Load() {
		t.Fatal("StreamInputWithWait did not complete after result was received")
	}
}

// TestStreamInputDoesNotWaitWhenNoHooksOrMCP tests that StreamInputWithWait returns
// immediately when there are no hooks or MCP servers active.
func TestStreamInputDoesNotWaitWhenNoHooksOrMCP(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// No hooks or MCP servers registered

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Create input channel
	input := make(chan map[string]any, 1)
	input <- map[string]any{"type": "user", "message": "test"}
	close(input)

	// StreamInputWithWait should complete quickly when no hooks/MCP
	done := make(chan struct{})
	go func() {
		query.StreamInputWithWait(input)
		close(done)
	}()

	select {
	case <-done:
		// Good - completed without waiting for result
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StreamInputWithWait should complete quickly when no hooks or MCP servers are active")
	}
}

// TestStreamInputWaitsForResultWhenMCPActive tests that StreamInputWithWait waits for
// the first result message before returning when MCP servers are active.
func TestStreamInputWaitsForResultWhenMCPActive(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Register an MCP server to make MCP active
	query.RegisterMCPServer(&types.MCPServer{
		Name:    "test-server",
		Version: "1.0.0",
	})

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Create input channel
	input := make(chan map[string]any, 1)
	input <- map[string]any{"type": "user", "message": "test"}
	close(input)

	// Track completion
	var streamInputDone atomic.Bool

	go func() {
		query.StreamInputWithWait(input)
		streamInputDone.Store(true)
	}()

	// Give time to process
	time.Sleep(50 * time.Millisecond)

	// Should NOT have completed (waiting for result)
	if streamInputDone.Load() {
		t.Fatal("StreamInputWithWait completed before result was received - should wait for result when MCP servers are active")
	}

	// Send result
	transport.SendMessage(map[string]any{
		"type":       "result",
		"session_id": "test-session",
	})

	// Wait for completion
	deadline := time.Now().Add(2 * time.Second)
	for !streamInputDone.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !streamInputDone.Load() {
		t.Fatal("StreamInputWithWait did not complete after result was received")
	}
}

// TestStreamInputTimeoutWhenNoResult tests that StreamInputWithWait times out
// if no result is received within the timeout period.
func TestStreamInputTimeoutWhenNoResult(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Set a short timeout for testing
	query.SetStreamCloseTimeout(100 * time.Millisecond)

	// Register a hook to make hooks active
	query.hookCallbacks["test_hook"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		return nil, nil
	}

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Create input channel
	input := make(chan map[string]any, 1)
	input <- map[string]any{"type": "user", "message": "test"}
	close(input)

	// StreamInputWithWait should timeout and complete
	done := make(chan struct{})
	start := time.Now()
	go func() {
		query.StreamInputWithWait(input)
		close(done)
	}()

	select {
	case <-done:
		elapsed := time.Since(start)
		// Should have taken approximately the timeout duration
		if elapsed < 80*time.Millisecond {
			t.Errorf("StreamInputWithWait completed too quickly: %v (expected ~100ms timeout)", elapsed)
		}
		if elapsed > 500*time.Millisecond {
			t.Errorf("StreamInputWithWait took too long: %v (expected ~100ms timeout)", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamInputWithWait did not timeout")
	}
}

// TestWaitForFirstResult tests that the first result event is properly signaled.
func TestWaitForFirstResult(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Initially no result received
	if query.ResultReceived() {
		t.Fatal("ResultReceived should be false initially")
	}

	// Get the wait channel
	waitChan := query.WaitForFirstResult()

	// Should not be ready yet
	select {
	case <-waitChan:
		t.Fatal("WaitForFirstResult should not return before result is received")
	case <-time.After(50 * time.Millisecond):
		// Good - still waiting
	}

	// Send result message
	transport.SendMessage(map[string]any{
		"type":       "result",
		"session_id": "test-session",
	})

	// Wait for result to be processed
	select {
	case <-waitChan:
		// Good - wait completed
	case <-time.After(1 * time.Second):
		t.Fatal("WaitForFirstResult did not complete after result was received")
	}

	// Now ResultReceived should be true
	if !query.ResultReceived() {
		t.Fatal("ResultReceived should be true after result is received")
	}
}

// TestHasActiveHooksOrMCP tests the helper method for detecting active hooks/MCP.
func TestHasActiveHooksOrMCP(t *testing.T) {
	transport := NewMockTransport()

	// Test with no hooks or MCP
	query := NewQuery(transport, true)
	if query.HasActiveHooksOrMCP() {
		t.Error("HasActiveHooksOrMCP should return false when no hooks or MCP")
	}

	// Test with hooks
	query.hookCallbacks["test"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		return nil, nil
	}
	if !query.HasActiveHooksOrMCP() {
		t.Error("HasActiveHooksOrMCP should return true when hooks are registered")
	}

	// Test with MCP only
	query2 := NewQuery(transport, true)
	query2.RegisterMCPServer(&types.MCPServer{Name: "test", Version: "1.0.0"})
	if !query2.HasActiveHooksOrMCP() {
		t.Error("HasActiveHooksOrMCP should return true when MCP servers are registered")
	}
}
