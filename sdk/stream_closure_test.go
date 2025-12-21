package sdk

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// mockTransportForStreamClosure is a mock transport for testing stream closure coordination.
type mockTransportForStreamClosure struct {
	messages   chan map[string]any
	errors     chan error
	writes     []string
	writeMu    sync.Mutex
	endInputCalled atomic.Bool
	endInputTime   time.Time
	endInputMu     sync.Mutex
}

func newMockTransportForStreamClosure() *mockTransportForStreamClosure {
	return &mockTransportForStreamClosure{
		messages: make(chan map[string]any, 100),
		errors:   make(chan error, 1),
		writes:   make([]string, 0),
	}
}

func (t *mockTransportForStreamClosure) Messages() <-chan map[string]any {
	return t.messages
}

func (t *mockTransportForStreamClosure) Errors() <-chan error {
	return t.errors
}

func (t *mockTransportForStreamClosure) Write(data string) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	t.writes = append(t.writes, data)
	return nil
}

func (t *mockTransportForStreamClosure) EndInput() error {
	t.endInputMu.Lock()
	defer t.endInputMu.Unlock()
	t.endInputCalled.Store(true)
	t.endInputTime = time.Now()
	return nil
}

func (t *mockTransportForStreamClosure) Close() error {
	close(t.messages)
	return nil
}

func (t *mockTransportForStreamClosure) IsReady() bool {
	return true
}

// TestStreamInputWaitsForResultWhenHooksActive tests that StreamInput waits for
// the first result message before calling EndInput when hooks are active.
func TestStreamInputWaitsForResultWhenHooksActive(t *testing.T) {
	transport := newMockTransportForStreamClosure()
	query := NewQuery(transport, true)

	// Register a hook to make hooks active
	query.hookCallbacks["test_hook"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		return nil, nil
	}

	// Start the query
	ctx := testContext(t)
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Create input channel with a single message
	input := make(chan map[string]any, 1)
	input <- map[string]any{"type": "user", "message": "test"}
	close(input)

	// Track when StreamInput completes
	var streamInputDone atomic.Bool
	var streamInputCompleteTime time.Time
	var streamInputMu sync.Mutex

	go func() {
		// StreamInput should wait for result before returning
		query.StreamInputWithWait(input)
		streamInputMu.Lock()
		streamInputCompleteTime = time.Now()
		streamInputMu.Unlock()
		streamInputDone.Store(true)
	}()

	// Give StreamInput time to process the input message
	time.Sleep(50 * time.Millisecond)

	// StreamInput should NOT have completed yet (waiting for result)
	if streamInputDone.Load() {
		t.Fatal("StreamInput completed before result was received - should wait for result when hooks are active")
	}

	// Now send a result message
	resultTime := time.Now()
	transport.messages <- map[string]any{
		"type":       "result",
		"session_id": "test-session",
	}

	// Wait for StreamInput to complete
	deadline := time.Now().Add(2 * time.Second)
	for !streamInputDone.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !streamInputDone.Load() {
		t.Fatal("StreamInput did not complete after result was received")
	}

	// Verify StreamInput completed AFTER result was received
	streamInputMu.Lock()
	completedAfterResult := streamInputCompleteTime.After(resultTime) || streamInputCompleteTime.Equal(resultTime)
	streamInputMu.Unlock()

	if !completedAfterResult {
		t.Error("StreamInput should complete after result is received")
	}
}

// TestStreamInputDoesNotWaitWhenNoHooksOrMCP tests that StreamInput returns
// immediately when there are no hooks or MCP servers active.
func TestStreamInputDoesNotWaitWhenNoHooksOrMCP(t *testing.T) {
	transport := newMockTransportForStreamClosure()
	query := NewQuery(transport, true)

	// No hooks or MCP servers registered

	ctx := testContext(t)
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Create input channel
	input := make(chan map[string]any, 1)
	input <- map[string]any{"type": "user", "message": "test"}
	close(input)

	// StreamInput should complete quickly when no hooks/MCP
	done := make(chan struct{})
	go func() {
		query.StreamInputWithWait(input)
		close(done)
	}()

	select {
	case <-done:
		// Good - completed without waiting for result
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StreamInput should complete quickly when no hooks or MCP servers are active")
	}
}

// TestStreamInputWaitsForResultWhenMCPActive tests that StreamInput waits for
// the first result message before calling EndInput when MCP servers are active.
func TestStreamInputWaitsForResultWhenMCPActive(t *testing.T) {
	transport := newMockTransportForStreamClosure()
	query := NewQuery(transport, true)

	// Register an MCP server to make MCP active
	query.RegisterMCPServer(&types.MCPServer{
		Name:    "test-server",
		Version: "1.0.0",
	})

	ctx := testContext(t)
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
		t.Fatal("StreamInput completed before result was received - should wait for result when MCP servers are active")
	}

	// Send result
	transport.messages <- map[string]any{
		"type":       "result",
		"session_id": "test-session",
	}

	// Wait for completion
	deadline := time.Now().Add(2 * time.Second)
	for !streamInputDone.Load() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if !streamInputDone.Load() {
		t.Fatal("StreamInput did not complete after result was received")
	}
}

// TestStreamInputTimeoutWhenNoResult tests that StreamInput times out
// if no result is received within the timeout period.
func TestStreamInputTimeoutWhenNoResult(t *testing.T) {
	transport := newMockTransportForStreamClosure()
	query := NewQuery(transport, true)

	// Set a short timeout for testing
	query.SetStreamCloseTimeout(100 * time.Millisecond)

	// Register a hook to make hooks active
	query.hookCallbacks["test_hook"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		return nil, nil
	}

	ctx := testContext(t)
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Create input channel
	input := make(chan map[string]any, 1)
	input <- map[string]any{"type": "user", "message": "test"}
	close(input)

	// StreamInput should timeout and complete
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
			t.Errorf("StreamInput completed too quickly: %v (expected ~100ms timeout)", elapsed)
		}
		if elapsed > 500*time.Millisecond {
			t.Errorf("StreamInput took too long: %v (expected ~100ms timeout)", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("StreamInput did not timeout")
	}
}

// TestFirstResultEventSignaling tests that the first result event is properly signaled.
func TestFirstResultEventSignaling(t *testing.T) {
	transport := newMockTransportForStreamClosure()
	query := NewQuery(transport, true)

	ctx := testContext(t)
	if err := query.Start(ctx); err != nil {
		t.Fatalf("Failed to start query: %v", err)
	}
	defer query.Close()

	// Initially no result received
	if query.ResultReceived() {
		t.Fatal("ResultReceived should be false initially")
	}

	// Wait for first result with timeout
	waitChan := query.WaitForFirstResult()

	select {
	case <-waitChan:
		t.Fatal("WaitForFirstResult should not return before result is received")
	case <-time.After(50 * time.Millisecond):
		// Good - still waiting
	}

	// Send result message
	transport.messages <- map[string]any{
		"type":       "result",
		"session_id": "test-session",
	}

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
