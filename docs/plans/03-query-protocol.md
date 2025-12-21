# Plan 03: Query/Control Protocol (Complete Feature Parity)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the Query layer that handles bidirectional control protocol on top of Transport, including MCP server integration and complete message handling.

**Architecture:** Use channels for control request/response routing. Use goroutines for concurrent message handling. Maintain pending requests map with request IDs. Support SDK-hosted MCP servers via in-process tools.

**Tech Stack:** Go 1.21+, sync, encoding/json

**Reference:** `.reference/claude-agent-sdk-python/src/claude_code_sdk/_internal/query.py`

---

## Task 1: Message Parser

**Files:**
- Create: `parser.go`
- Create: `parser_test.go`

**Step 1: Write failing test**

Create `parser_test.go`:

```go
package sdk

import (
	"testing"
)

func TestParseMessage_System(t *testing.T) {
	raw := map[string]any{
		"type":    "system",
		"subtype": "init",
		"data": map[string]any{
			"version":    "2.0.0",
			"session_id": "test_123",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	sys, ok := msg.(*SystemMessage)
	if !ok {
		t.Fatalf("expected *SystemMessage, got %T", msg)
	}

	if sys.Subtype != "init" {
		t.Errorf("got subtype %q, want init", sys.Subtype)
	}

	if sys.SessionID != "test_123" {
		t.Errorf("got session_id %q, want test_123", sys.SessionID)
	}
}

func TestParseMessage_Assistant(t *testing.T) {
	raw := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "Hello!"},
			},
			"model": "claude-sonnet-4-5",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	asst, ok := msg.(*AssistantMessage)
	if !ok {
		t.Fatalf("expected *AssistantMessage, got %T", msg)
	}

	if asst.Model != "claude-sonnet-4-5" {
		t.Errorf("got model %q, want claude-sonnet-4-5", asst.Model)
	}

	if asst.Text() != "Hello!" {
		t.Errorf("got text %q, want Hello!", asst.Text())
	}
}

func TestParseMessage_Result(t *testing.T) {
	raw := map[string]any{
		"type":          "result",
		"subtype":       "success",
		"duration_ms":   float64(1000),
		"session_id":    "test_123",
		"total_cost_usd": float64(0.001),
		"num_turns":     float64(5),
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	result, ok := msg.(*ResultMessage)
	if !ok {
		t.Fatalf("expected *ResultMessage, got %T", msg)
	}

	if result.Subtype != "success" {
		t.Errorf("got subtype %q, want success", result.Subtype)
	}

	if result.TotalCostUSD == nil || *result.TotalCostUSD != 0.001 {
		t.Errorf("got cost %v, want 0.001", result.TotalCostUSD)
	}
}

func TestParseMessage_User(t *testing.T) {
	raw := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": "Hello Claude!",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	user, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	if user.Text() != "Hello Claude!" {
		t.Errorf("got text %q, want Hello Claude!", user.Text())
	}
}

func TestParseMessage_StreamEvent(t *testing.T) {
	raw := map[string]any{
		"type":       "stream_event",
		"uuid":       "event_123",
		"session_id": "sess_456",
		"event": map[string]any{
			"type":  "content_block_delta",
			"index": float64(0),
			"delta": map[string]any{
				"type": "text_delta",
				"text": "Hello",
			},
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	event, ok := msg.(*StreamEvent)
	if !ok {
		t.Fatalf("expected *StreamEvent, got %T", msg)
	}

	if event.UUID != "event_123" {
		t.Errorf("got uuid %q, want event_123", event.UUID)
	}
	if event.SessionID != "sess_456" {
		t.Errorf("got session_id %q, want sess_456", event.SessionID)
	}
	if event.EventType != "content_block_delta" {
		t.Errorf("got event_type %q, want content_block_delta", event.EventType)
	}
	if event.Index == nil || *event.Index != 0 {
		t.Error("expected index 0")
	}
}

func TestParseMessage_WithParentToolUseID(t *testing.T) {
	raw := map[string]any{
		"type":               "user",
		"uuid":               "msg_123",
		"parent_tool_use_id": "tool_456",
		"message": map[string]any{
			"role":    "user",
			"content": "Subagent response",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	user, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	if user.UUID != "msg_123" {
		t.Errorf("got uuid %q, want msg_123", user.UUID)
	}
	if user.ParentToolUseID == nil || *user.ParentToolUseID != "tool_456" {
		t.Error("expected parent_tool_use_id tool_456")
	}
}

func TestParseContentBlock(t *testing.T) {
	tests := []struct {
		name string
		raw  map[string]any
		want string // content type
	}{
		{
			name: "text",
			raw:  map[string]any{"type": "text", "text": "hello"},
			want: "text",
		},
		{
			name: "thinking",
			raw:  map[string]any{"type": "thinking", "thinking": "hmm"},
			want: "thinking",
		},
		{
			name: "tool_use",
			raw: map[string]any{
				"type":  "tool_use",
				"id":    "tool_1",
				"name":  "Bash",
				"input": map[string]any{"command": "ls"},
			},
			want: "tool_use",
		},
		{
			name: "tool_result",
			raw: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool_1",
				"content":     "file1.txt\nfile2.txt",
			},
			want: "tool_result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := parseContentBlock(tt.raw)
			if err != nil {
				t.Fatalf("parseContentBlock failed: %v", err)
			}
			if block.Type() != tt.want {
				t.Errorf("got type %q, want %q", block.Type(), tt.want)
			}
		})
	}
}
```

**Step 2: Write implementation**

Create `parser.go`:

```go
package sdk

import (
	"fmt"
)

// ParseMessage parses a raw message map into a typed Message.
func ParseMessage(raw map[string]any) (Message, error) {
	msgType, _ := raw["type"].(string)

	switch msgType {
	case "system":
		return parseSystemMessage(raw)
	case "assistant":
		return parseAssistantMessage(raw)
	case "user":
		return parseUserMessage(raw)
	case "result":
		return parseResultMessage(raw)
	case "stream_event":
		return parseStreamEvent(raw)
	default:
		return nil, &ParseError{
			Message: fmt.Sprintf("unknown message type: %s", msgType),
		}
	}
}

// parseStreamEvent parses a StreamEvent for partial message updates.
func parseStreamEvent(raw map[string]any) (*StreamEvent, error) {
	event := &StreamEvent{
		UUID:      getString(raw, "uuid"),
		SessionID: getString(raw, "session_id"),
	}

	// Extract parent_tool_use_id if present
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		event.ParentToolUseID = &parentID
	}

	// Parse the nested event data
	if eventData, ok := raw["event"].(map[string]any); ok {
		event.Event = eventData
		event.EventType = getString(eventData, "type")

		// Extract index if present (for content_block events)
		if idx, ok := eventData["index"].(float64); ok {
			idxInt := int(idx)
			event.Index = &idxInt
		}

		// Extract delta if present
		if delta, ok := eventData["delta"].(map[string]any); ok {
			event.Delta = delta
		}
	}

	return event, nil
}

func parseSystemMessage(raw map[string]any) (*SystemMessage, error) {
	msg := &SystemMessage{
		Subtype: getString(raw, "subtype"),
	}

	if data, ok := raw["data"].(map[string]any); ok {
		msg.SessionID = getString(data, "session_id")
		msg.Version = getString(data, "version")
		msg.Data = data
	}

	return msg, nil
}

func parseAssistantMessage(raw map[string]any) (*AssistantMessage, error) {
	msg := &AssistantMessage{}

	// Extract parent_tool_use_id for subagent messages
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}

	// Extract error field for API error messages
	if errType, ok := raw["error"].(string); ok {
		err := AssistantMessageError(errType)
		msg.Error = &err
	}

	if msgData, ok := raw["message"].(map[string]any); ok {
		msg.Model = getString(msgData, "model")
		msg.StopReason = getString(msgData, "stop_reason")

		// Parse content blocks
		if content, ok := msgData["content"].([]any); ok {
			for _, item := range content {
				if blockRaw, ok := item.(map[string]any); ok {
					block, err := parseContentBlock(blockRaw)
					if err != nil {
						continue // Skip invalid blocks
					}
					msg.Content = append(msg.Content, block)
				}
			}
		}
	}

	return msg, nil
}

func parseUserMessage(raw map[string]any) (*UserMessage, error) {
	msg := &UserMessage{
		UUID: getString(raw, "uuid"),
	}

	// Extract parent_tool_use_id for subagent messages
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}

	if msgData, ok := raw["message"].(map[string]any); ok {
		msg.Role = getString(msgData, "role")

		// Content can be string or array
		switch c := msgData["content"].(type) {
		case string:
			msg.Content = []ContentBlock{&TextBlock{TextContent: c}}
		case []any:
			for _, item := range c {
				if blockRaw, ok := item.(map[string]any); ok {
					block, err := parseContentBlock(blockRaw)
					if err != nil {
						continue
					}
					msg.Content = append(msg.Content, block)
				}
			}
		}
	}

	return msg, nil
}

func parseResultMessage(raw map[string]any) (*ResultMessage, error) {
	msg := &ResultMessage{
		Subtype:   getString(raw, "subtype"),
		SessionID: getString(raw, "session_id"),
		IsError:   getBool(raw, "is_error"),
	}

	if dur, ok := raw["duration_ms"].(float64); ok {
		msg.DurationMS = int(dur)
	}
	if durAPI, ok := raw["duration_api_ms"].(float64); ok {
		msg.DurationAPIMS = int(durAPI)
	}
	if turns, ok := raw["num_turns"].(float64); ok {
		msg.NumTurns = int(turns)
	}
	if cost, ok := raw["total_cost_usd"].(float64); ok {
		msg.TotalCostUSD = &cost
	}

	return msg, nil
}

func parseContentBlock(raw map[string]any) (ContentBlock, error) {
	blockType, _ := raw["type"].(string)

	switch blockType {
	case "text":
		return &TextBlock{
			TextContent: getString(raw, "text"),
		}, nil

	case "thinking":
		return &ThinkingBlock{
			ThinkingContent: getString(raw, "thinking"),
		}, nil

	case "tool_use":
		input, _ := raw["input"].(map[string]any)
		return &ToolUseBlock{
			ID:        getString(raw, "id"),
			Name:      getString(raw, "name"),
			ToolInput: input,
		}, nil

	case "tool_result":
		var content string
		switch c := raw["content"].(type) {
		case string:
			content = c
		}
		return &ToolResultBlock{
			ToolUseID:     getString(raw, "tool_use_id"),
			ResultContent: content,
			IsError:       getBool(raw, "is_error"),
		}, nil

	default:
		return nil, fmt.Errorf("unknown content block type: %s", blockType)
	}
}

// Helper functions
func getString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func getBool(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}

func getFloat(m map[string]any, key string) float64 {
	v, _ := m[key].(float64)
	return v
}
```

**Step 3: Run tests**

```bash
go test -run "TestParseMessage|TestParseContentBlock" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add parser.go parser_test.go
git commit -m "feat: add message parser"
```

---

## Task 2: Mock Transport for Testing

**Files:**
- Create: `mock_transport_test.go`

**Step 1: Create mock transport**

Create `mock_transport_test.go`:

```go
package sdk

import (
	"sync"
)

// MockTransport is a test double for Transport.
type MockTransport struct {
	messages chan map[string]any
	errors   chan error
	written  []string
	writeMu  sync.Mutex
	closed   bool
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		messages: make(chan map[string]any, 100),
		errors:   make(chan error, 1),
		written:  make([]string, 0),
	}
}

func (m *MockTransport) Messages() <-chan map[string]any {
	return m.messages
}

func (m *MockTransport) Errors() <-chan error {
	return m.errors
}

func (m *MockTransport) Write(data string) error {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	m.written = append(m.written, data)
	return nil
}

func (m *MockTransport) WriteJSON(obj any) error {
	return nil // Simplified for tests
}

func (m *MockTransport) Close() error {
	if !m.closed {
		m.closed = true
		close(m.messages)
	}
	return nil
}

func (m *MockTransport) IsReady() bool {
	return !m.closed
}

// Test helpers
func (m *MockTransport) Written() []string {
	m.writeMu.Lock()
	defer m.writeMu.Unlock()
	return append([]string{}, m.written...)
}

func (m *MockTransport) SendMessage(msg map[string]any) {
	m.messages <- msg
}
```

**Step 2: Commit**

```bash
git add mock_transport_test.go
git commit -m "test: add mock transport for testing"
```

---

## Task 3: Query Structure with Enhanced Message Handling

**Files:**
- Create: `query.go`
- Create: `query_test.go`

**Step 1: Write failing test**

Create `query_test.go`:

```go
package sdk

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestNewQuery(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	if query == nil {
		t.Fatal("NewQuery returned nil")
	}

	if query.transport != transport {
		t.Error("transport not set correctly")
	}

	if !query.streaming {
		t.Error("streaming flag not set")
	}
}

func TestQuery_Start(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	err := query.Start(ctx)
	if err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// Should be able to receive messages
	transport.messages <- map[string]any{"type": "system", "subtype": "init"}

	select {
	case msg := <-query.Messages():
		if msg.Type() != "system" {
			t.Errorf("got type %v, want system", msg.Type())
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message")
	}

	query.Close()
}

func TestQuery_RawMessages(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.messages <- map[string]any{"type": "test", "custom": true}

	select {
	case msg := <-query.RawMessages():
		if msg["type"] != "test" {
			t.Errorf("got type %v, want test", msg["type"])
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for raw message")
	}
}

func TestQuery_ResultReceived(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send a result message
	transport.messages <- map[string]any{
		"type":       "result",
		"subtype":    "success",
		"session_id": "test_123",
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !query.ResultReceived() {
		t.Error("expected result to be received")
	}
}
```

**Step 2: Write implementation**

Create `query.go`:

```go
package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Query handles the bidirectional control protocol.
type Query struct {
	transport Transport
	streaming bool

	// Control protocol state
	pendingRequests map[string]chan map[string]any
	pendingMu       sync.Mutex
	requestCounter  atomic.Uint64

	// Hook callbacks
	hookCallbacks  map[string]HookCallback
	nextCallbackID atomic.Uint64
	hookMu         sync.RWMutex

	// Permission callback
	canUseTool CanUseToolCallback

	// Message channels
	messages    chan Message          // Parsed messages
	rawMessages chan map[string]any   // Raw messages for custom handling
	errors      chan error

	// Result tracking
	resultReceived atomic.Bool
	lastResult     *ResultMessage
	resultMu       sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool

	// Initialization result
	initResult map[string]any
	initMu     sync.RWMutex
}

// NewQuery creates a new Query.
func NewQuery(transport Transport, streaming bool) *Query {
	return &Query{
		transport:       transport,
		streaming:       streaming,
		pendingRequests: make(map[string]chan map[string]any),
		hookCallbacks:   make(map[string]HookCallback),
		messages:        make(chan Message, 100),
		rawMessages:     make(chan map[string]any, 100),
		errors:          make(chan error, 1),
	}
}

// Start begins processing messages from the transport.
func (q *Query) Start(ctx context.Context) error {
	q.ctx, q.cancel = context.WithCancel(ctx)

	// Start message router
	q.wg.Add(1)
	go q.routeMessages()

	return nil
}

// Messages returns the channel of parsed SDK messages.
func (q *Query) Messages() <-chan Message {
	return q.messages
}

// RawMessages returns the channel of raw messages for custom handling.
func (q *Query) RawMessages() <-chan map[string]any {
	return q.rawMessages
}

// Errors returns the channel of errors.
func (q *Query) Errors() <-chan error {
	return q.errors
}

// ResultReceived returns true if a result message has been received.
func (q *Query) ResultReceived() bool {
	return q.resultReceived.Load()
}

// LastResult returns the last result message received.
func (q *Query) LastResult() *ResultMessage {
	q.resultMu.RLock()
	defer q.resultMu.RUnlock()
	return q.lastResult
}

// Close stops the query.
func (q *Query) Close() error {
	if q.closed.Swap(true) {
		return nil // Already closed
	}

	if q.cancel != nil {
		q.cancel()
	}

	q.wg.Wait()

	// Close channels after goroutines finish
	close(q.messages)
	close(q.rawMessages)

	return nil
}

// routeMessages reads from transport and routes control vs SDK messages.
func (q *Query) routeMessages() {
	defer q.wg.Done()

	for {
		select {
		case <-q.ctx.Done():
			return
		case raw, ok := <-q.transport.Messages():
			if !ok {
				return
			}

			msgType, _ := raw["type"].(string)

			switch msgType {
			case "control_response":
				q.handleControlResponse(raw)
			case "control_request":
				go q.handleControlRequest(raw)
			case "control_cancel_request":
				q.handleCancelRequest(raw)
			default:
				// Parse and route regular messages
				q.handleSDKMessage(raw)
			}
		}
	}
}

// handleSDKMessage parses and routes an SDK message.
func (q *Query) handleSDKMessage(raw map[string]any) {
	// Send to raw channel for custom handling
	select {
	case q.rawMessages <- raw:
	default:
		// Drop if full
	}

	// Parse message
	msg, err := ParseMessage(raw)
	if err != nil {
		// Still send to channel for unknown message types
		select {
		case q.errors <- err:
		default:
		}
		return
	}

	// Track result messages
	if result, ok := msg.(*ResultMessage); ok {
		q.resultMu.Lock()
		q.lastResult = result
		q.resultMu.Unlock()
		q.resultReceived.Store(true)
	}

	// Send parsed message
	select {
	case q.messages <- msg:
	case <-q.ctx.Done():
	}
}

// handleCancelRequest handles a request cancellation.
func (q *Query) handleCancelRequest(raw map[string]any) {
	requestID, _ := raw["request_id"].(string)
	if requestID == "" {
		return
	}

	q.pendingMu.Lock()
	if respChan, exists := q.pendingRequests[requestID]; exists {
		// Send cancellation signal
		select {
		case respChan <- map[string]any{"cancelled": true}:
		default:
		}
	}
	q.pendingMu.Unlock()
}
```

**Step 3: Run tests**

```bash
go test -run "TestNewQuery|TestQuery_Start|TestQuery_RawMessages|TestQuery_ResultReceived" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add Query structure with enhanced message handling"
```

---

## Task 4: Control Request/Response

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
func TestQuery_SendControlRequest(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate response in background
	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.messages <- map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"status": "ok"},
				},
			}
		}
	}()

	response, err := query.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 5*time.Second)

	if err != nil {
		t.Errorf("sendControlRequest failed: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("unexpected response: %v", response)
	}
}

func TestQuery_SendControlRequest_Timeout(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Don't send a response - should timeout
	_, err := query.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 100*time.Millisecond)

	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestQuery_SendControlRequest_Error(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.messages <- map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "error",
					"request_id": reqID,
					"error":      "something went wrong",
				},
			}
		}
	}()

	_, err := query.sendControlRequest(map[string]any{
		"subtype": "test",
	}, 5*time.Second)

	if err == nil {
		t.Error("expected error response")
	}
}
```

**Step 2: Write implementation**

Add to `query.go`:

```go
// sendControlRequest sends a control request and waits for response.
func (q *Query) sendControlRequest(request map[string]any, timeout time.Duration) (map[string]any, error) {
	if !q.streaming {
		return nil, fmt.Errorf("control requests require streaming mode")
	}

	// Generate request ID
	id := q.requestCounter.Add(1)
	requestID := fmt.Sprintf("req_%d", id)

	// Create response channel
	respChan := make(chan map[string]any, 1)
	q.pendingMu.Lock()
	q.pendingRequests[requestID] = respChan
	q.pendingMu.Unlock()

	defer func() {
		q.pendingMu.Lock()
		delete(q.pendingRequests, requestID)
		q.pendingMu.Unlock()
	}()

	// Build and send request
	controlReq := map[string]any{
		"type":       "control_request",
		"request_id": requestID,
		"request":    request,
	}

	data, err := json.Marshal(controlReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := q.transport.Write(string(data)); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}

	// Wait for response
	select {
	case resp := <-respChan:
		if errMsg, ok := resp["error"].(string); ok {
			return nil, fmt.Errorf("control request error: %s", errMsg)
		}
		if resp["cancelled"] == true {
			return nil, fmt.Errorf("control request cancelled")
		}
		return resp, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("control request timeout: %v", request["subtype"])
	case <-q.ctx.Done():
		return nil, q.ctx.Err()
	}
}

// handleControlResponse routes a control response to the waiting request.
func (q *Query) handleControlResponse(msg map[string]any) {
	response, ok := msg["response"].(map[string]any)
	if !ok {
		return
	}

	requestID, _ := response["request_id"].(string)
	if requestID == "" {
		return
	}

	q.pendingMu.Lock()
	respChan, exists := q.pendingRequests[requestID]
	q.pendingMu.Unlock()

	if !exists {
		return
	}

	// Check for error
	if response["subtype"] == "error" {
		errMsg, _ := response["error"].(string)
		select {
		case respChan <- map[string]any{"error": errMsg}:
		default:
		}
		return
	}

	// Send success response
	respData, _ := response["response"].(map[string]any)
	if respData == nil {
		respData = make(map[string]any)
	}

	select {
	case respChan <- respData:
	default:
	}
}
```

**Step 3: Run tests**

```bash
go test -run TestQuery_SendControlRequest -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add control request/response handling"
```

---

## Task 5: Initialize Method with Hook Registration

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
func TestQuery_Initialize(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate init response
	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.messages <- map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response": map[string]any{
						"commands": []any{"/help", "/clear"},
					},
				},
			}
		}
	}()

	result, err := query.Initialize(nil)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}

	if result == nil {
		t.Error("expected initialization result")
	}
}

func TestQuery_Initialize_WithHooks(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			// Verify hooks are in request
			request := req["request"].(map[string]any)
			if request["hooks"] == nil {
				t.Error("expected hooks in request")
			}

			transport.messages <- map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			}
		}
	}()

	hooks := map[HookEvent][]HookMatcher{
		HookPreToolUse: {
			{
				Matcher: map[string]any{"tool_name": "Bash"},
				Hooks: []HookCallback{
					func(input any, toolUseID *string) (*HookOutput, error) {
						return &HookOutput{Continue: true}, nil
					},
				},
			},
		},
	}

	_, err := query.Initialize(hooks)
	if err != nil {
		t.Errorf("Initialize with hooks failed: %v", err)
	}
}
```

**Step 2: Write implementation**

Add to `query.go`:

```go
// Initialize sends the initialization request to the CLI.
func (q *Query) Initialize(hooks map[HookEvent][]HookMatcher) (map[string]any, error) {
	if !q.streaming {
		return nil, nil
	}

	// Build hooks configuration
	hooksConfig := q.buildHooksConfig(hooks)

	request := map[string]any{
		"subtype": "initialize",
	}
	if len(hooksConfig) > 0 {
		request["hooks"] = hooksConfig
	}

	result, err := q.sendControlRequest(request, 60*time.Second)
	if err != nil {
		return nil, err
	}

	q.initMu.Lock()
	q.initResult = result
	q.initMu.Unlock()

	return result, nil
}

// buildHooksConfig builds the hooks configuration for initialization.
func (q *Query) buildHooksConfig(hooks map[HookEvent][]HookMatcher) map[string]any {
	if hooks == nil {
		return nil
	}

	config := make(map[string]any)

	for event, matchers := range hooks {
		if len(matchers) == 0 {
			continue
		}

		var matcherConfigs []map[string]any
		for _, matcher := range matchers {
			callbackIDs := make([]string, len(matcher.Hooks))
			for i, callback := range matcher.Hooks {
				callbackID := fmt.Sprintf("hook_%d", q.nextCallbackID.Add(1))
				q.hookMu.Lock()
				q.hookCallbacks[callbackID] = callback
				q.hookMu.Unlock()
				callbackIDs[i] = callbackID
			}

			matcherConfig := map[string]any{
				"matcher":         matcher.Matcher,
				"hookCallbackIds": callbackIDs,
			}
			if matcher.Timeout > 0 {
				matcherConfig["timeout"] = matcher.Timeout
			}
			matcherConfigs = append(matcherConfigs, matcherConfig)
		}
		config[string(event)] = matcherConfigs
	}

	return config
}

// InitResult returns the initialization result.
func (q *Query) InitResult() map[string]any {
	q.initMu.RLock()
	defer q.initMu.RUnlock()
	return q.initResult
}
```

**Step 3: Run tests**

```bash
go test -run TestQuery_Initialize -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add Initialize method with hook registration"
```

---

## Task 6: Interrupt and Control Methods

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write tests**

Add to `query_test.go`:

```go
func TestQuery_Interrupt(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.messages <- map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			}
		}
	}()

	err := query.Interrupt()
	if err != nil {
		t.Errorf("Interrupt failed: %v", err)
	}
}

func TestQuery_SetPermissionMode(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.messages <- map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			}
		}
	}()

	err := query.SetPermissionMode(PermissionBypass)
	if err != nil {
		t.Errorf("SetPermissionMode failed: %v", err)
	}
}

func TestQuery_SetModel(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			// Verify model is in request
			request := req["request"].(map[string]any)
			if request["model"] != "claude-opus-4" {
				t.Errorf("expected model claude-opus-4, got %v", request["model"])
			}

			transport.messages <- map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			}
		}
	}()

	err := query.SetModel("claude-opus-4")
	if err != nil {
		t.Errorf("SetModel failed: %v", err)
	}
}
```

**Step 2: Write implementation**

Add to `query.go`:

```go
// Interrupt sends an interrupt signal to the CLI.
func (q *Query) Interrupt() error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 30*time.Second)
	return err
}

// SetPermissionMode changes the permission mode.
func (q *Query) SetPermissionMode(mode PermissionMode) error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype": "set_permission_mode",
		"mode":    string(mode),
	}, 30*time.Second)
	return err
}

// SetModel changes the AI model.
func (q *Query) SetModel(model string) error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype": "set_model",
		"model":   model,
	}, 30*time.Second)
	return err
}

// RewindFiles rewinds tracked files to a specific user message.
func (q *Query) RewindFiles(userMessageID string) error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype":         "rewind_files",
		"user_message_id": userMessageID,
	}, 30*time.Second)
	return err
}
```

**Step 3: Run tests**

```bash
go test -run "TestQuery_Interrupt|TestQuery_SetPermissionMode|TestQuery_SetModel" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add Interrupt, SetPermissionMode, SetModel, RewindFiles"
```

---

## Task 7: Hook Callback Handling

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
func TestQuery_HandleHookCallback(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Register a hook callback
	callbackCalled := false
	query.hookMu.Lock()
	query.hookCallbacks["hook_1"] = func(input any, toolUseID *string) (*HookOutput, error) {
		callbackCalled = true
		return &HookOutput{Continue: true}, nil
	}
	query.hookMu.Unlock()

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send hook callback request
	transport.messages <- map[string]any{
		"type":       "control_request",
		"request_id": "req_hook_1",
		"request": map[string]any{
			"subtype":     "hook_callback",
			"callback_id": "hook_1",
			"input": map[string]any{
				"tool_name": "Bash",
			},
		},
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !callbackCalled {
		t.Error("hook callback was not called")
	}

	// Verify response was sent
	written := transport.Written()
	if len(written) == 0 {
		t.Error("no response was written")
	}
}

func TestQuery_HandleHookCallback_Error(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	query.hookMu.Lock()
	query.hookCallbacks["hook_err"] = func(input any, toolUseID *string) (*HookOutput, error) {
		return nil, fmt.Errorf("hook error")
	}
	query.hookMu.Unlock()

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.messages <- map[string]any{
		"type":       "control_request",
		"request_id": "req_hook_err",
		"request": map[string]any{
			"subtype":     "hook_callback",
			"callback_id": "hook_err",
			"input":       map[string]any{},
		},
	}

	time.Sleep(100 * time.Millisecond)

	// Verify error response was sent
	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no response was written")
	}

	var resp map[string]any
	json.Unmarshal([]byte(written[0]), &resp)
	response := resp["response"].(map[string]any)
	if response["subtype"] != "error" {
		t.Error("expected error response")
	}
}
```

**Step 2: Write implementation**

Add to `query.go`:

```go
// handleControlRequest handles incoming control requests from CLI.
func (q *Query) handleControlRequest(msg map[string]any) {
	requestID, _ := msg["request_id"].(string)
	request, _ := msg["request"].(map[string]any)
	if request == nil {
		return
	}

	subtype, _ := request["subtype"].(string)

	var responseData map[string]any
	var err error

	switch subtype {
	case "can_use_tool":
		responseData, err = q.handleCanUseTool(request)
	case "hook_callback":
		responseData, err = q.handleHookCallback(request)
	default:
		err = fmt.Errorf("unsupported control request: %s", subtype)
	}

	// Send response
	q.sendControlResponse(requestID, responseData, err)
}

// handleHookCallback invokes a registered hook callback.
func (q *Query) handleHookCallback(request map[string]any) (map[string]any, error) {
	callbackID, _ := request["callback_id"].(string)
	input := request["input"]

	var toolUseID *string
	if id, ok := request["tool_use_id"].(string); ok {
		toolUseID = &id
	}

	q.hookMu.RLock()
	callback, exists := q.hookCallbacks[callbackID]
	q.hookMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("hook callback not found: %s", callbackID)
	}

	output, err := callback(input, toolUseID)
	if err != nil {
		return nil, err
	}

	// Convert HookOutput to response
	return q.hookOutputToResponse(output), nil
}

// hookOutputToResponse converts a HookOutput to a response map.
func (q *Query) hookOutputToResponse(output *HookOutput) map[string]any {
	if output == nil {
		return make(map[string]any)
	}

	result := make(map[string]any)

	if output.Continue {
		result["continue"] = true
	}
	if output.SuppressOutput {
		result["suppressOutput"] = true
	}
	if output.StopReason != "" {
		result["stopReason"] = output.StopReason
	}
	if output.Decision != "" {
		result["decision"] = output.Decision
	}
	if output.SystemMessage != "" {
		result["systemMessage"] = output.SystemMessage
	}
	if output.Reason != "" {
		result["reason"] = output.Reason
	}
	if output.HookSpecific != nil {
		result["hookSpecificOutput"] = output.HookSpecific
	}

	return result
}

// handleCanUseTool handles tool permission requests.
func (q *Query) handleCanUseTool(request map[string]any) (map[string]any, error) {
	if q.canUseTool == nil {
		// Default: allow all
		return map[string]any{"behavior": "allow"}, nil
	}

	toolName, _ := request["tool_name"].(string)
	input, _ := request["input"].(map[string]any)
	suggestions, _ := request["permission_suggestions"].([]any)

	ctx := &ToolPermissionContext{
		Suggestions: q.parsePermissionSuggestions(suggestions),
	}

	result, err := q.canUseTool(toolName, input, ctx)
	if err != nil {
		return nil, err
	}

	return q.permissionResultToResponse(result)
}

// parsePermissionSuggestions converts raw suggestions to PermissionUpdate slice.
func (q *Query) parsePermissionSuggestions(raw []any) []PermissionUpdate {
	if raw == nil {
		return nil
	}

	var updates []PermissionUpdate
	for _, item := range raw {
		if m, ok := item.(map[string]any); ok {
			update := PermissionUpdate{
				Type:        getString(m, "type"),
				ToolName:    getString(m, "tool_name"),
				Destination: getString(m, "destination"),
			}
			updates = append(updates, update)
		}
	}
	return updates
}

// permissionResultToResponse converts a permission result to a response map.
func (q *Query) permissionResultToResponse(result any) (map[string]any, error) {
	switch r := result.(type) {
	case *PermissionResultAllow:
		resp := map[string]any{
			"behavior": "allow",
		}
		if r.UpdatedInput != nil {
			resp["updatedInput"] = r.UpdatedInput
		}
		if len(r.PermissionUpdates) > 0 {
			resp["permissionUpdates"] = r.PermissionUpdates
		}
		return resp, nil

	case *PermissionResultDeny:
		resp := map[string]any{
			"behavior": "deny",
			"message":  r.Message,
		}
		if r.Interrupt {
			resp["interrupt"] = true
		}
		return resp, nil

	default:
		return nil, fmt.Errorf("invalid permission result type: %T", result)
	}
}

// sendControlResponse sends a control response.
func (q *Query) sendControlResponse(requestID string, data map[string]any, err error) {
	response := map[string]any{
		"type": "control_response",
	}

	if err != nil {
		response["response"] = map[string]any{
			"subtype":    "error",
			"request_id": requestID,
			"error":      err.Error(),
		}
	} else {
		response["response"] = map[string]any{
			"subtype":    "success",
			"request_id": requestID,
			"response":   data,
		}
	}

	responseData, _ := json.Marshal(response)
	q.transport.Write(string(responseData))
}
```

**Step 3: Run tests**

```bash
go test -run TestQuery_HandleHookCallback -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add hook callback and permission handling"
```

---

## Task 8: Stream Input and Message Sending

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
func TestQuery_SendUserMessage(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	err := query.SendUserMessage("Hello Claude!", "session_123")
	if err != nil {
		t.Errorf("SendUserMessage failed: %v", err)
	}

	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no message written")
	}

	var msg map[string]any
	json.Unmarshal([]byte(written[0]), &msg)

	if msg["type"] != "user" {
		t.Errorf("expected type user, got %v", msg["type"])
	}
}

func TestQuery_StreamInput(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Create input channel
	input := make(chan map[string]any, 3)
	input <- map[string]any{"type": "user", "message": map[string]any{"content": "hello"}}
	input <- map[string]any{"type": "user", "message": map[string]any{"content": "world"}}
	close(input)

	err := query.StreamInput(input)
	if err != nil {
		t.Errorf("StreamInput failed: %v", err)
	}

	written := transport.Written()
	if len(written) != 2 {
		t.Errorf("expected 2 messages written, got %d", len(written))
	}
}

func TestQuery_SetCanUseTool(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	called := false
	query.SetCanUseTool(func(toolName string, input map[string]any, ctx *ToolPermissionContext) (any, error) {
		called = true
		return &PermissionResultAllow{Behavior: "allow"}, nil
	})

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate can_use_tool request
	transport.messages <- map[string]any{
		"type":       "control_request",
		"request_id": "req_perm_1",
		"request": map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "Bash",
			"input":     map[string]any{"command": "ls"},
		},
	}

	time.Sleep(100 * time.Millisecond)

	if !called {
		t.Error("canUseTool callback was not called")
	}
}
```

**Step 2: Write implementation**

Add to `query.go`:

```go
// StreamInput streams messages from a channel to the CLI.
func (q *Query) StreamInput(input <-chan map[string]any) error {
	for msg := range input {
		select {
		case <-q.ctx.Done():
			return q.ctx.Err()
		default:
		}

		if err := q.SendMessage(msg); err != nil {
			return err
		}
	}

	return nil
}

// SendMessage sends a single message to the CLI.
func (q *Query) SendMessage(msg map[string]any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return q.transport.Write(string(data))
}

// SendUserMessage sends a user message.
func (q *Query) SendUserMessage(content string, sessionID string) error {
	msg := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": content,
		},
		"parent_tool_use_id": nil,
		"session_id":         sessionID,
	}
	return q.SendMessage(msg)
}

// SetCanUseTool sets the tool permission callback.
func (q *Query) SetCanUseTool(callback CanUseToolCallback) {
	q.canUseTool = callback
}
```

**Step 3: Run tests**

```bash
go test -run "TestQuery_SendUserMessage|TestQuery_StreamInput|TestQuery_SetCanUseTool" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add message sending and SetCanUseTool"
```

---

## Task 9: MCP Server Support Types

**Files:**
- Create: `mcp.go`
- Create: `mcp_test.go`

**Step 1: Write failing test**

Create `mcp_test.go`:

```go
package sdk

import (
	"context"
	"testing"
)

func TestMCPToolHandler(t *testing.T) {
	handler := &MCPToolHandler{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"input": map[string]any{"type": "string"},
			},
		},
		Handler: func(ctx context.Context, input map[string]any) (any, error) {
			return map[string]any{"result": "success"}, nil
		},
	}

	if handler.Name != "test_tool" {
		t.Errorf("expected name test_tool, got %s", handler.Name)
	}

	result, err := handler.Handler(context.Background(), map[string]any{"input": "test"})
	if err != nil {
		t.Errorf("handler failed: %v", err)
	}

	resultMap := result.(map[string]any)
	if resultMap["result"] != "success" {
		t.Errorf("expected success, got %v", resultMap["result"])
	}
}

func TestMCPServerBuilder(t *testing.T) {
	server := NewMCPServerBuilder("test-server").
		WithTool("greet", "Greets a user", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}, func(ctx context.Context, input map[string]any) (any, error) {
			name := input["name"].(string)
			return map[string]any{"greeting": "Hello, " + name + "!"}, nil
		}).
		Build()

	if server.Name != "test-server" {
		t.Errorf("expected name test-server, got %s", server.Name)
	}

	if len(server.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(server.Tools))
	}

	if server.Tools[0].Name != "greet" {
		t.Errorf("expected tool name greet, got %s", server.Tools[0].Name)
	}
}
```

**Step 2: Write implementation**

Create `mcp.go`:

```go
package sdk

import (
	"context"
	"fmt"
	"sync"
)

// MCPToolHandler defines a tool that can be called via MCP.
type MCPToolHandler struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     func(ctx context.Context, input map[string]any) (any, error)
}

// MCPServer represents an SDK-hosted MCP server.
type MCPServer struct {
	Name  string
	Tools []*MCPToolHandler

	// Internal state
	toolsByName map[string]*MCPToolHandler
	mu          sync.RWMutex
}

// NewMCPServer creates a new MCP server.
func NewMCPServer(name string, tools []*MCPToolHandler) *MCPServer {
	server := &MCPServer{
		Name:        name,
		Tools:       tools,
		toolsByName: make(map[string]*MCPToolHandler),
	}

	for _, tool := range tools {
		server.toolsByName[tool.Name] = tool
	}

	return server
}

// GetTool returns a tool by name.
func (s *MCPServer) GetTool(name string) (*MCPToolHandler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tool, ok := s.toolsByName[name]
	return tool, ok
}

// CallTool calls a tool by name with the given input.
func (s *MCPServer) CallTool(ctx context.Context, name string, input map[string]any) (any, error) {
	tool, ok := s.GetTool(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool.Handler(ctx, input)
}

// ToConfig returns the MCP server configuration for the CLI.
func (s *MCPServer) ToConfig() map[string]any {
	tools := make([]map[string]any, len(s.Tools))
	for i, tool := range s.Tools {
		tools[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return map[string]any{
		"name":  s.Name,
		"tools": tools,
	}
}

// MCPServerBuilder provides a fluent API for building MCP servers.
type MCPServerBuilder struct {
	name  string
	tools []*MCPToolHandler
}

// NewMCPServerBuilder creates a new MCP server builder.
func NewMCPServerBuilder(name string) *MCPServerBuilder {
	return &MCPServerBuilder{
		name:  name,
		tools: make([]*MCPToolHandler, 0),
	}
}

// WithTool adds a tool to the server.
func (b *MCPServerBuilder) WithTool(
	name string,
	description string,
	inputSchema map[string]any,
	handler func(ctx context.Context, input map[string]any) (any, error),
) *MCPServerBuilder {
	b.tools = append(b.tools, &MCPToolHandler{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Handler:     handler,
	})
	return b
}

// Build creates the MCP server.
func (b *MCPServerBuilder) Build() *MCPServer {
	return NewMCPServer(b.name, b.tools)
}
```

**Step 3: Run tests**

```bash
go test -run "TestMCPToolHandler|TestMCPServerBuilder" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add mcp.go mcp_test.go
git commit -m "feat: add MCP server support types"
```

---

## Task 10: Query MCP Tool Call Handling

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
func TestQuery_MCPToolCall(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Register an MCP server
	server := NewMCPServerBuilder("test-server").
		WithTool("echo", "Echoes input", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{"type": "string"},
			},
		}, func(ctx context.Context, input map[string]any) (any, error) {
			return map[string]any{"echo": input["message"]}, nil
		}).
		Build()

	query.RegisterMCPServer(server)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate MCP tool call request
	transport.messages <- map[string]any{
		"type":       "control_request",
		"request_id": "req_mcp_1",
		"request": map[string]any{
			"subtype":     "mcp_tool_call",
			"server_name": "test-server",
			"tool_name":   "echo",
			"input":       map[string]any{"message": "hello"},
		},
	}

	time.Sleep(100 * time.Millisecond)

	// Verify response was sent
	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no response written")
	}

	var resp map[string]any
	json.Unmarshal([]byte(written[0]), &resp)
	response := resp["response"].(map[string]any)
	if response["subtype"] != "success" {
		t.Errorf("expected success, got %v", response["subtype"])
	}
}
```

**Step 2: Write implementation**

Add to `query.go`:

```go
// MCP server registry
var (
	mcpServers   = make(map[string]*MCPServer)
	mcpServersMu sync.RWMutex
)

// RegisterMCPServer registers an MCP server with the query.
func (q *Query) RegisterMCPServer(server *MCPServer) {
	mcpServersMu.Lock()
	defer mcpServersMu.Unlock()
	mcpServers[server.Name] = server
}

// UnregisterMCPServer removes an MCP server.
func (q *Query) UnregisterMCPServer(name string) {
	mcpServersMu.Lock()
	defer mcpServersMu.Unlock()
	delete(mcpServers, name)
}

// Add MCP tool call handling to handleControlRequest
// Update the switch in handleControlRequest:
/*
	case "mcp_tool_call":
		responseData, err = q.handleMCPToolCall(request)
*/

// handleMCPToolCall handles MCP tool call requests (simplified protocol).
func (q *Query) handleMCPToolCall(request map[string]any) (map[string]any, error) {
	serverName, _ := request["server_name"].(string)
	toolName, _ := request["tool_name"].(string)
	input, _ := request["input"].(map[string]any)

	mcpServersMu.RLock()
	server, exists := mcpServers[serverName]
	mcpServersMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP server not found: %s", serverName)
	}

	result, err := server.CallTool(q.ctx, toolName, input)
	if err != nil {
		return nil, err
	}

	return map[string]any{"result": result}, nil
}

// handleMCPMessage handles full MCP JSONRPC protocol messages.
// This is the complete MCP protocol bridge that handles:
// - initialize: Returns server capabilities
// - notifications/initialized: Acknowledgement (no response)
// - tools/list: Returns available tools
// - tools/call: Invokes a tool
func (q *Query) handleMCPMessage(request map[string]any) (map[string]any, error) {
	serverName, _ := request["server_name"].(string)
	message, _ := request["message"].(map[string]any)

	if message == nil {
		return nil, fmt.Errorf("missing message in mcp_message request")
	}

	mcpServersMu.RLock()
	server, exists := mcpServers[serverName]
	mcpServersMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP server not found: %s", serverName)
	}

	method, _ := message["method"].(string)
	id := message["id"] // Can be string or number
	params, _ := message["params"].(map[string]any)

	switch method {
	case "initialize":
		return q.handleMCPInitialize(server, id, params)
	case "notifications/initialized":
		// Notification - no response needed
		return nil, nil
	case "tools/list":
		return q.handleMCPToolsList(server, id)
	case "tools/call":
		return q.handleMCPToolsCall(server, id, params)
	default:
		return q.buildMCPError(id, -32601, fmt.Sprintf("Method not found: %s", method)), nil
	}
}

// handleMCPInitialize handles the initialize method.
func (q *Query) handleMCPInitialize(server *MCPServer, id any, params map[string]any) (map[string]any, error) {
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    server.Name,
				"version": "1.0.0",
			},
		},
	}, nil
}

// handleMCPToolsList handles the tools/list method.
func (q *Query) handleMCPToolsList(server *MCPServer, id any) (map[string]any, error) {
	tools := make([]map[string]any, len(server.Tools))
	for i, tool := range server.Tools {
		tools[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		}
	}

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"tools": tools,
		},
	}, nil
}

// handleMCPToolsCall handles the tools/call method.
func (q *Query) handleMCPToolsCall(server *MCPServer, id any, params map[string]any) (map[string]any, error) {
	toolName, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]any)

	tool, ok := server.GetTool(toolName)
	if !ok {
		return q.buildMCPError(id, -32602, fmt.Sprintf("Tool not found: %s", toolName)), nil
	}

	result, err := tool.Handler(q.ctx, arguments)
	if err != nil {
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"result": map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": err.Error()},
				},
				"isError": true,
			},
		}, nil
	}

	// Convert result to MCP content format
	content := q.resultToMCPContent(result)

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]any{
			"content": content,
		},
	}, nil
}

// resultToMCPContent converts a tool result to MCP content format.
func (q *Query) resultToMCPContent(result any) []map[string]any {
	switch r := result.(type) {
	case string:
		return []map[string]any{{"type": "text", "text": r}}
	case map[string]any:
		// Check if already in MCP format
		if _, hasContent := r["content"]; hasContent {
			if content, ok := r["content"].([]map[string]any); ok {
				return content
			}
		}
		// Convert to JSON text
		data, _ := json.Marshal(r)
		return []map[string]any{{"type": "text", "text": string(data)}}
	default:
		data, _ := json.Marshal(result)
		return []map[string]any{{"type": "text", "text": string(data)}}
	}
}

// buildMCPError creates a JSONRPC error response.
func (q *Query) buildMCPError(id any, code int, message string) map[string]any {
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
}
```

Update `handleControlRequest` to include MCP handlers:

```go
func (q *Query) handleControlRequest(msg map[string]any) {
	requestID, _ := msg["request_id"].(string)
	request, _ := msg["request"].(map[string]any)
	if request == nil {
		return
	}

	subtype, _ := request["subtype"].(string)

	var responseData map[string]any
	var err error

	switch subtype {
	case "can_use_tool":
		responseData, err = q.handleCanUseTool(request)
	case "hook_callback":
		responseData, err = q.handleHookCallback(request)
	case "mcp_tool_call":
		responseData, err = q.handleMCPToolCall(request)
	case "mcp_message":
		// Full MCP JSONRPC protocol bridge
		responseData, err = q.handleMCPMessage(request)
		// For notifications, responseData is nil - don't send response
		if responseData == nil && err == nil {
			return
		}
	default:
		err = fmt.Errorf("unsupported control request: %s", subtype)
	}

	q.sendControlResponse(requestID, responseData, err)
}
```

**Step 3: Run tests**

```bash
go test -run TestQuery_MCPToolCall -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add MCP tool call handling"
```

---

## Summary

After completing Plan 03, you have:

- [x] Message parser with content block parsing
- [x] **StreamEvent parsing** for partial message updates (include_partial_messages)
- [x] **parent_tool_use_id and uuid fields** for subagent message tracking
- [x] Mock transport for testing
- [x] Query structure with enhanced message handling
- [x] Control request/response handling
- [x] Initialize method with hook registration
- [x] Interrupt, SetPermissionMode, SetModel, RewindFiles methods
- [x] Hook callback handling with typed outputs
- [x] Stream input and message sending
- [x] MCP server support types and builder
- [x] MCP tool call handling (simplified protocol)
- [x] **Full MCP JSONRPC bridge** (initialize, notifications/initialized, tools/list, tools/call)

**Key Features:**
- Bidirectional control protocol with request/response matching
- Result message tracking for conversation completion
- Raw message channel for custom handling
- Full hook callback support with typed outputs
- **Complete MCP JSONRPC 2.0 protocol implementation**
- SDK-hosted MCP server integration
- Permission handling with suggestions
- **StreamEvent support for real-time UI updates**
- **Subagent message tracking via parent_tool_use_id**

**Next:** Plan 04 - Client API
