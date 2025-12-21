# Plan 04: Client API

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the high-level Client that wraps Transport and Query for easy use.

**Architecture:** Client manages lifecycle. Provides simple Query() for one-shot and streaming modes. Uses functional options for configuration.

**Tech Stack:** Go 1.21+, context

---

## Task 1: Client Structure

**Files:**
- Create: `client.go`
- Create: `client_test.go`

**Step 1: Write failing test**

Create `client_test.go`:

```go
package sdk

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient(
		WithModel("claude-opus-4"),
		WithMaxTurns(10),
	)

	if client.options.Model != "claude-opus-4" {
		t.Errorf("got model %q, want %q", client.options.Model, "claude-opus-4")
	}
	if client.options.MaxTurns != 10 {
		t.Errorf("got maxTurns %d, want %d", client.options.MaxTurns, 10)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestNewClient -v
```

Expected: FAIL - Client not defined

**Step 3: Write implementation**

Create `client.go`:

```go
package sdk

import (
	"context"
	"sync"
)

// Client is the high-level interface for Claude Agent SDK.
type Client struct {
	options   *Options
	transport Transport
	query     *Query

	connected bool
	mu        sync.Mutex
}

// NewClient creates a new SDK client.
func NewClient(opts ...Option) *Client {
	options := DefaultOptions()
	ApplyOptions(options, opts...)

	return &Client{
		options: options,
	}
}

// Options returns the client's options.
func (c *Client) Options() *Options {
	return c.options
}

// IsConnected returns true if the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}
```

**Step 4: Run tests**

```bash
go test -run TestNewClient -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add Client structure"
```

---

## Task 2: Connect Method

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestClient_Connect(t *testing.T) {
	// Use mock transport
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	ctx := context.Background()
	err := client.Connect(ctx, "")
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("client should be connected")
	}

	client.Close()
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestClient_Connect -v
```

Expected: FAIL - Connect not defined

**Step 3: Write implementation**

Add to `client.go`:

```go
// Connect establishes a connection to Claude.
// If prompt is empty, streaming mode is used.
func (c *Client) Connect(ctx context.Context, prompt string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Create transport if not provided (for testing)
	if c.transport == nil {
		c.transport = NewSubprocessTransport(prompt, c.options)
	}

	// Connect transport
	if err := c.transport.Connect(ctx); err != nil {
		return err
	}

	// Create query
	streaming := prompt == ""
	c.query = NewQuery(c.transport, streaming)

	// Set callbacks
	// (canUseTool would be set via options in future)

	// Start query
	if err := c.query.Start(ctx); err != nil {
		c.transport.Close()
		return err
	}

	// Initialize if streaming
	if streaming {
		if _, err := c.query.Initialize(nil); err != nil {
			c.transport.Close()
			return err
		}
	}

	c.connected = true
	return nil
}

// Disconnect closes the connection.
func (c *Client) Disconnect() error {
	return c.Close()
}

// Close closes the client and releases resources.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false

	if c.query != nil {
		c.query.Close()
		c.query = nil
	}

	if c.transport != nil {
		c.transport.Close()
		c.transport = nil
	}

	return nil
}
```

**Step 4: Run tests**

```bash
go test -run TestClient_Connect -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add Connect and Close methods"
```

---

## Task 3: Query Function (One-Shot)

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestQuery_OneShot(t *testing.T) {
	// Create mock transport that returns messages
	transport := NewMockTransport()
	go func() {
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
				"model": "claude-test",
			},
		}
		transport.messages <- map[string]any{
			"type":           "result",
			"subtype":        "success",
			"duration_ms":    100,
			"duration_api_ms": 80,
			"is_error":       false,
			"num_turns":      1,
			"session_id":     "test_123",
		}
		close(transport.messages)
	}()

	ctx := context.Background()
	messages, err := Query(ctx, "Hello", WithTransport(transport))

	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_OneShot -v
```

Expected: FAIL - Query function not defined

**Step 3: Write implementation**

Add to `client.go`:

```go
// Query performs a one-shot query and returns all messages.
func Query(ctx context.Context, prompt string, opts ...Option) ([]Message, error) {
	options := DefaultOptions()
	ApplyOptions(options, opts...)

	// Extract transport if provided
	var transport Transport
	if options.customTransport != nil {
		transport = options.customTransport
	} else {
		transport = NewSubprocessTransport(prompt, options)
	}

	// Connect
	if err := transport.Connect(ctx); err != nil {
		return nil, err
	}
	defer transport.Close()

	// Collect messages
	var messages []Message

	for msg := range transport.Messages() {
		parsed, err := ParseMessage(msg)
		if err != nil {
			// Skip unparseable messages
			continue
		}
		messages = append(messages, parsed)

		// Stop on result
		if _, ok := parsed.(*ResultMessage); ok {
			break
		}
	}

	return messages, nil
}

// Add to options.go
type Options struct {
	// ... existing fields ...

	// Internal: custom transport for testing
	customTransport Transport
}

// WithTransport sets a custom transport (for testing).
func WithTransport(t Transport) Option {
	return func(o *Options) {
		o.customTransport = t
	}
}
```

Also add the message parser:

Add to `parser.go`:

```go
package sdk

import (
	"encoding/json"
	"fmt"
)

// ParseMessage parses a raw JSON message into a typed Message.
func ParseMessage(data map[string]any) (Message, error) {
	msgType, ok := data["type"].(string)
	if !ok {
		return nil, &ParseError{Message: "missing type field"}
	}

	switch msgType {
	case "user":
		return parseUserMessage(data)
	case "assistant":
		return parseAssistantMessage(data)
	case "system":
		return parseSystemMessage(data)
	case "result":
		return parseResultMessage(data)
	case "stream_event":
		return parseStreamEvent(data)
	default:
		return nil, &ParseError{Message: fmt.Sprintf("unknown message type: %s", msgType)}
	}
}

func parseUserMessage(data map[string]any) (*UserMessage, error) {
	msg := &UserMessage{}

	if msgData, ok := data["message"].(map[string]any); ok {
		msg.Content = msgData["content"]
	}
	if uuid, ok := data["uuid"].(string); ok {
		msg.UUID = &uuid
	}
	if parentID, ok := data["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}

	return msg, nil
}

func parseAssistantMessage(data map[string]any) (*AssistantMessage, error) {
	msg := &AssistantMessage{}

	msgData, ok := data["message"].(map[string]any)
	if !ok {
		return nil, &ParseError{Message: "missing message field"}
	}

	msg.Model, _ = msgData["model"].(string)

	if content, ok := msgData["content"].([]any); ok {
		for _, block := range content {
			if blockMap, ok := block.(map[string]any); ok {
				parsed, err := ParseContentBlock(blockMap)
				if err == nil {
					msg.Content = append(msg.Content, parsed)
				}
			}
		}
	}

	if parentID, ok := data["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}
	if errStr, ok := msgData["error"].(string); ok {
		msg.Error = &errStr
	}

	return msg, nil
}

func parseSystemMessage(data map[string]any) (*SystemMessage, error) {
	msg := &SystemMessage{
		Data: data,
	}
	msg.Subtype, _ = data["subtype"].(string)
	return msg, nil
}

func parseResultMessage(data map[string]any) (*ResultMessage, error) {
	msg := &ResultMessage{}

	msg.Subtype, _ = data["subtype"].(string)
	msg.SessionID, _ = data["session_id"].(string)

	if v, ok := data["duration_ms"].(float64); ok {
		msg.DurationMS = int(v)
	}
	if v, ok := data["duration_api_ms"].(float64); ok {
		msg.DurationAPI = int(v)
	}
	msg.IsError, _ = data["is_error"].(bool)
	if v, ok := data["num_turns"].(float64); ok {
		msg.NumTurns = int(v)
	}
	if v, ok := data["total_cost_usd"].(float64); ok {
		msg.TotalCostUSD = &v
	}
	if v, ok := data["usage"].(map[string]any); ok {
		msg.Usage = v
	}
	if v, ok := data["result"].(string); ok {
		msg.Result = &v
	}
	msg.StructuredOutput = data["structured_output"]

	return msg, nil
}

func parseStreamEvent(data map[string]any) (*StreamEvent, error) {
	msg := &StreamEvent{}

	msg.UUID, _ = data["uuid"].(string)
	msg.SessionID, _ = data["session_id"].(string)
	msg.Event, _ = data["event"].(map[string]any)

	if parentID, ok := data["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}

	return msg, nil
}
```

**Step 4: Run tests**

```bash
go test -run TestQuery_OneShot -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add client.go client_test.go parser.go options.go
git commit -m "feat: add Query function and message parser"
```

---

## Task 4: QueryStream Function

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestQueryStream(t *testing.T) {
	transport := NewMockTransport()
	go func() {
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
				"model": "claude-test",
			},
		}
		transport.messages <- map[string]any{
			"type":           "result",
			"subtype":        "success",
			"duration_ms":    100,
			"duration_api_ms": 80,
			"is_error":       false,
			"num_turns":      1,
			"session_id":     "test_123",
		}
		close(transport.messages)
	}()

	ctx := context.Background()
	msgChan, errChan := QueryStream(ctx, "Hello", WithTransport(transport))

	var messages []Message
	for msg := range msgChan {
		messages = append(messages, msg)
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	default:
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQueryStream -v
```

Expected: FAIL - QueryStream not defined

**Step 3: Write implementation**

Add to `client.go`:

```go
// QueryStream performs a query and streams messages back.
func QueryStream(ctx context.Context, prompt string, opts ...Option) (<-chan Message, <-chan error) {
	msgChan := make(chan Message, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(msgChan)
		defer close(errChan)

		options := DefaultOptions()
		ApplyOptions(options, opts...)

		var transport Transport
		if options.customTransport != nil {
			transport = options.customTransport
		} else {
			transport = NewSubprocessTransport(prompt, options)
		}

		if err := transport.Connect(ctx); err != nil {
			errChan <- err
			return
		}
		defer transport.Close()

		for msg := range transport.Messages() {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			default:
			}

			parsed, err := ParseMessage(msg)
			if err != nil {
				continue
			}

			select {
			case msgChan <- parsed:
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}

			if _, ok := parsed.(*ResultMessage); ok {
				return
			}
		}
	}()

	return msgChan, errChan
}
```

**Step 4: Run tests**

```bash
go test -run TestQueryStream -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add QueryStream function"
```

---

## Task 5: Streaming Client Methods

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestClient_Query(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	go func() {
		// Wait for query to be written
		time.Sleep(10 * time.Millisecond)

		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
				"model": "claude-test",
			},
		}
		transport.messages <- map[string]any{
			"type":           "result",
			"subtype":        "success",
			"duration_ms":    100,
			"duration_api_ms": 80,
			"is_error":       false,
			"num_turns":      1,
			"session_id":     "test_123",
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx, ""); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Send query
	err := client.SendQuery("Hello")
	if err != nil {
		t.Errorf("SendQuery failed: %v", err)
	}

	// Receive response
	msg, err := client.ReceiveMessage()
	if err != nil {
		t.Errorf("ReceiveMessage failed: %v", err)
	}

	if _, ok := msg.(*AssistantMessage); !ok {
		t.Errorf("expected AssistantMessage, got %T", msg)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestClient_Query -v
```

Expected: FAIL - SendQuery not defined

**Step 3: Write implementation**

Add to `client.go`:

```go
import "time"

// SendQuery sends a query in streaming mode.
func (c *Client) SendQuery(prompt string, sessionID ...string) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	sid := "default"
	if len(sessionID) > 0 {
		sid = sessionID[0]
	}

	return q.SendUserMessage(prompt, sid)
}

// ReceiveMessage receives the next message.
func (c *Client) ReceiveMessage() (Message, error) {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return nil, &ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	select {
	case msg, ok := <-q.Messages():
		if !ok {
			return nil, &ConnectionError{Message: "channel closed"}
		}
		return ParseMessage(msg)
	case err := <-q.Errors():
		return nil, err
	}
}

// ReceiveAll receives all messages until result.
func (c *Client) ReceiveAll() ([]Message, error) {
	var messages []Message
	for {
		msg, err := c.ReceiveMessage()
		if err != nil {
			return messages, err
		}
		messages = append(messages, msg)
		if _, ok := msg.(*ResultMessage); ok {
			return messages, nil
		}
	}
}

// Interrupt sends an interrupt signal.
func (c *Client) Interrupt() error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.Interrupt()
}

// SetPermissionMode changes the permission mode.
func (c *Client) SetPermissionMode(mode PermissionMode) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.SetPermissionMode(mode)
}

// SetModel changes the AI model.
func (c *Client) SetModel(model string) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.SetModel(model)
}

// RewindFiles rewinds tracked files to a specific user message.
func (c *Client) RewindFiles(userMessageID string) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &ConnectionError{Message: "not connected"}
	}
	q := c.query
	c.mu.Unlock()

	return q.RewindFiles(userMessageID)
}

// ServerInfo returns the initialization info.
func (c *Client) ServerInfo() map[string]any {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.query != nil {
		return c.query.InitResult()
	}
	return nil
}
```

**Step 4: Run tests**

```bash
go test -run TestClient_Query -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add streaming client methods"
```

---

## Task 6: Context Manager Pattern

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestClient_WithContext(t *testing.T) {
	transport := NewMockTransport()

	ctx := context.Background()
	err := WithClient(ctx, func(client *Client) error {
		client.transport = transport
		return nil
	})

	if err != nil {
		t.Errorf("WithClient failed: %v", err)
	}

	// Transport should be closed
	if transport.connected {
		t.Error("transport should be closed after WithClient")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestClient_WithContext -v
```

Expected: FAIL - WithClient not defined

**Step 3: Write implementation**

Add to `client.go`:

```go
// ClientFunc is a function that uses a client.
type ClientFunc func(*Client) error

// WithClient creates a client, runs the function, and ensures cleanup.
func WithClient(ctx context.Context, fn ClientFunc, opts ...Option) error {
	client := NewClient(opts...)

	// The function can set up the transport and connect
	if err := fn(client); err != nil {
		client.Close()
		return err
	}

	defer client.Close()
	return nil
}

// Run connects and runs a function with the client.
func (c *Client) Run(ctx context.Context, fn func() error) error {
	if err := c.Connect(ctx, ""); err != nil {
		return err
	}
	defer c.Close()
	return fn()
}
```

**Step 4: Run tests**

```bash
go test -run TestClient_WithContext -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add context manager pattern"
```

---

## Task 7: Helper Methods

**Files:**
- Modify: `client.go`
- Modify: `types.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestAssistantMessage_Text(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "Hello "},
			&TextBlock{Text: "World"},
		},
	}

	text := msg.Text()
	if text != "Hello World" {
		t.Errorf("got %q, want %q", text, "Hello World")
	}
}

func TestAssistantMessage_ToolCalls(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "Let me help"},
			&ToolUseBlock{ID: "tool_1", Name: "Bash", Input: map[string]any{"command": "ls"}},
			&ToolUseBlock{ID: "tool_2", Name: "Read", Input: map[string]any{"path": "/tmp"}},
		},
	}

	tools := msg.ToolCalls()
	if len(tools) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(tools))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run "TestAssistantMessage_Text|TestAssistantMessage_ToolCalls" -v
```

Expected: FAIL - Text method not defined

**Step 3: Write implementation**

Add to `types.go`:

```go
import "strings"

// Text returns all text content concatenated.
func (m *AssistantMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if textBlock, ok := block.(*TextBlock); ok {
			parts = append(parts, textBlock.Text)
		}
	}
	return strings.Join(parts, "")
}

// ToolCalls returns all tool use blocks.
func (m *AssistantMessage) ToolCalls() []*ToolUseBlock {
	var tools []*ToolUseBlock
	for _, block := range m.Content {
		if toolBlock, ok := block.(*ToolUseBlock); ok {
			tools = append(tools, toolBlock)
		}
	}
	return tools
}

// Thinking returns the thinking content if present.
func (m *AssistantMessage) Thinking() string {
	for _, block := range m.Content {
		if thinkingBlock, ok := block.(*ThinkingBlock); ok {
			return thinkingBlock.Thinking
		}
	}
	return ""
}

// IsSuccess returns true if the result is successful.
func (m *ResultMessage) IsSuccess() bool {
	return !m.IsError && m.Subtype == "success"
}

// Cost returns the cost in USD.
func (m *ResultMessage) Cost() float64 {
	if m.TotalCostUSD != nil {
		return *m.TotalCostUSD
	}
	return 0
}
```

**Step 4: Run tests**

```bash
go test -run "TestAssistantMessage_Text|TestAssistantMessage_ToolCalls" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add client.go types.go client_test.go
git commit -m "feat: add helper methods for messages"
```

---

## Summary

After completing Plan 04, you have:

- [x] Client structure with options
- [x] Connect and Close methods
- [x] Query function (one-shot)
- [x] QueryStream function
- [x] Streaming client methods (SendQuery, ReceiveMessage, etc.)
- [x] Context manager pattern (WithClient, Run)
- [x] Helper methods for messages

**Next:** Plan 05 - Integration & Examples
