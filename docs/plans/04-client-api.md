# Plan 04: Client API (Complete Feature Parity)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the high-level Client that wraps Transport and Query for easy use, with complete feature parity including hooks, MCP servers, and session management.

**Architecture:** Client manages lifecycle. Provides simple Query() for one-shot and streaming modes. Uses functional options for configuration. Supports hook registration, MCP server hosting, and session resume/continue.

**Tech Stack:** Go 1.21+, context

**Reference:** `.reference/claude-agent-sdk-python/src/claude_code_sdk/_client.py`

---

## Task 1: Client Structure with Complete Options

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
	"time"
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
		WithPermissionMode(PermissionBypass),
		WithSystemPrompt("You are helpful"),
	)

	if client.options.Model != "claude-opus-4" {
		t.Errorf("got model %q, want %q", client.options.Model, "claude-opus-4")
	}
	if client.options.MaxTurns != 10 {
		t.Errorf("got maxTurns %d, want %d", client.options.MaxTurns, 10)
	}
	if client.options.PermissionMode != PermissionBypass {
		t.Errorf("got permission mode %v, want %v", client.options.PermissionMode, PermissionBypass)
	}
}

func TestClientWithMCPServers(t *testing.T) {
	server := NewMCPServerBuilder("test-server").
		WithTool("echo", "Echoes input", map[string]any{
			"type": "object",
		}, func(ctx context.Context, input map[string]any) (any, error) {
			return input, nil
		}).
		Build()

	client := NewClient(
		WithMCPServer(server),
	)

	if len(client.mcpServers) != 1 {
		t.Errorf("expected 1 MCP server, got %d", len(client.mcpServers))
	}
}

func TestClientWithHooks(t *testing.T) {
	preToolUseCalled := false
	client := NewClient(
		WithPreToolUseHook(
			map[string]any{"tool_name": "Bash"},
			func(input any, toolUseID *string) (*HookOutput, error) {
				preToolUseCalled = true
				return &HookOutput{Continue: true}, nil
			},
		),
	)

	if len(client.hooks) != 1 {
		t.Errorf("expected 1 hook event, got %d", len(client.hooks))
	}
	_ = preToolUseCalled // Used when hook is invoked
}

func TestClientWithCanUseTool(t *testing.T) {
	called := false
	client := NewClient(
		WithCanUseTool(func(toolName string, input map[string]any, ctx *ToolPermissionContext) (any, error) {
			called = true
			return &PermissionResultAllow{Behavior: "allow"}, nil
		}),
	)

	if client.canUseTool == nil {
		t.Error("canUseTool callback not set")
	}
	_ = called
}
```

**Step 2: Write implementation**

Create `client.go`:

```go
package sdk

import (
	"context"
	"sync"
)

// Client is the high-level interface for Claude Agent SDK.
type Client struct {
	options *Options

	// MCP servers hosted by this client
	mcpServers map[string]*MCPServer

	// Hooks registered for this client
	hooks map[HookEvent][]HookMatcher

	// Permission callback
	canUseTool CanUseToolCallback

	// Transport and query
	transport Transport
	query     *Query

	// Session management
	sessionID string

	// State
	connected bool
	mu        sync.Mutex
}

// NewClient creates a new SDK client.
func NewClient(opts ...Option) *Client {
	options := DefaultOptions()
	ApplyOptions(options, opts...)

	client := &Client{
		options:    options,
		mcpServers: make(map[string]*MCPServer),
		hooks:      make(map[HookEvent][]HookMatcher),
	}

	// Apply client-specific options
	for _, opt := range opts {
		if clientOpt, ok := opt.(clientOption); ok {
			clientOpt.applyClient(client)
		}
	}

	return client
}

// clientOption is an option that applies to the client.
type clientOption interface {
	applyClient(*Client)
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

// SessionID returns the current session ID.
func (c *Client) SessionID() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sessionID
}

// WithMCPServer adds an MCP server to the client.
func WithMCPServer(server *MCPServer) Option {
	return &mcpServerOption{server: server}
}

type mcpServerOption struct {
	server *MCPServer
}

func (o *mcpServerOption) Apply(opts *Options) {}

func (o *mcpServerOption) applyClient(c *Client) {
	c.mcpServers[o.server.Name] = o.server
}

// WithPreToolUseHook adds a pre-tool-use hook.
func WithPreToolUseHook(matcher map[string]any, callback HookCallback) Option {
	return &hookOption{
		event:    HookPreToolUse,
		matcher:  matcher,
		callback: callback,
	}
}

// WithPostToolUseHook adds a post-tool-use hook.
func WithPostToolUseHook(matcher map[string]any, callback HookCallback) Option {
	return &hookOption{
		event:    HookPostToolUse,
		matcher:  matcher,
		callback: callback,
	}
}

// WithStopHook adds a stop hook.
func WithStopHook(matcher map[string]any, callback HookCallback) Option {
	return &hookOption{
		event:    HookStop,
		matcher:  matcher,
		callback: callback,
	}
}

type hookOption struct {
	event    HookEvent
	matcher  map[string]any
	callback HookCallback
}

func (o *hookOption) Apply(opts *Options) {}

func (o *hookOption) applyClient(c *Client) {
	c.hooks[o.event] = append(c.hooks[o.event], HookMatcher{
		Matcher: o.matcher,
		Hooks:   []HookCallback{o.callback},
	})
}

// WithCanUseTool sets the tool permission callback.
func WithCanUseTool(callback CanUseToolCallback) Option {
	return &canUseToolOption{callback: callback}
}

type canUseToolOption struct {
	callback CanUseToolCallback
}

func (o *canUseToolOption) Apply(opts *Options) {}

func (o *canUseToolOption) applyClient(c *Client) {
	c.canUseTool = o.callback
}
```

**Step 3: Run tests**

```bash
go test -run "TestNewClient|TestClientWithMCPServers|TestClientWithHooks|TestClientWithCanUseTool" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add Client structure with complete options"
```

---

## Task 2: Connect Method with Initialization

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestClient_Connect(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	// Simulate init response
	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
			"data": map[string]any{
				"version":    "2.0.0",
				"session_id": "test_session_123",
			},
		}
	}()

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("client should be connected")
	}

	client.Close()
}

func TestClient_ConnectWithPrompt(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	ctx := context.Background()
	err := client.ConnectWithPrompt(ctx, "Hello Claude!")
	if err != nil {
		t.Errorf("ConnectWithPrompt failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("client should be connected")
	}

	client.Close()
}

func TestClient_Resume(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(
		WithResume("previous_session_id"),
	)
	client.transport = transport

	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
			"data": map[string]any{
				"session_id": "previous_session_id",
			},
		}
	}()

	ctx := context.Background()
	err := client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect with resume failed: %v", err)
	}

	if client.SessionID() != "previous_session_id" {
		t.Error("session ID not set from resume")
	}

	client.Close()
}
```

**Step 2: Write implementation**

Add to `client.go`:

```go
// Connect establishes a connection to Claude in streaming mode.
func (c *Client) Connect(ctx context.Context) error {
	return c.connect(ctx, "", true)
}

// ConnectWithPrompt establishes a connection with an initial prompt (non-streaming).
func (c *Client) ConnectWithPrompt(ctx context.Context, prompt string) error {
	return c.connect(ctx, prompt, false)
}

// connect is the internal connection method.
func (c *Client) connect(ctx context.Context, prompt string, streaming bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Create transport if not provided (for testing)
	if c.transport == nil {
		if streaming {
			c.transport = NewStreamingTransport(c.options)
		} else {
			c.transport = NewSubprocessTransport(prompt, c.options)
		}
	}

	// Connect transport
	if err := c.transport.Connect(ctx); err != nil {
		return err
	}

	// Create query
	c.query = NewQuery(c.transport, streaming)

	// Set permission callback
	if c.canUseTool != nil {
		c.query.SetCanUseTool(c.canUseTool)
	}

	// Register MCP servers
	for _, server := range c.mcpServers {
		c.query.RegisterMCPServer(server)
	}

	// Start query
	if err := c.query.Start(ctx); err != nil {
		c.transport.Close()
		return err
	}

	// Initialize if streaming
	if streaming {
		result, err := c.query.Initialize(c.hooks)
		if err != nil {
			c.transport.Close()
			return err
		}

		// Extract session ID from init response
		if result != nil {
			if sid, ok := result["session_id"].(string); ok {
				c.sessionID = sid
			}
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

	// Unregister MCP servers
	if c.query != nil {
		for name := range c.mcpServers {
			c.query.UnregisterMCPServer(name)
		}
		c.query.Close()
		c.query = nil
	}

	if c.transport != nil {
		c.transport.Close()
		c.transport = nil
	}

	return nil
}

// WithResume sets the session ID to resume.
func WithResume(sessionID string) Option {
	return func(o *Options) {
		o.Resume = sessionID
	}
}

// WithContinue enables continuing the last conversation.
func WithContinue() Option {
	return func(o *Options) {
		o.ContinueConversation = true
	}
}
```

**Step 3: Run tests**

```bash
go test -run "TestClient_Connect|TestClient_Resume" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add Connect method with initialization"
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
			"duration_ms":    float64(100),
			"duration_api_ms": float64(80),
			"is_error":       false,
			"num_turns":      float64(1),
			"session_id":     "test_123",
			"total_cost_usd": float64(0.001),
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

	// Verify assistant message
	if asst, ok := messages[0].(*AssistantMessage); ok {
		if asst.Text() != "Hello!" {
			t.Errorf("got text %q, want Hello!", asst.Text())
		}
	} else {
		t.Errorf("expected AssistantMessage, got %T", messages[0])
	}

	// Verify result message
	if result, ok := messages[1].(*ResultMessage); ok {
		if !result.IsSuccess() {
			t.Error("expected success result")
		}
		if result.Cost() != 0.001 {
			t.Errorf("got cost %f, want 0.001", result.Cost())
		}
	} else {
		t.Errorf("expected ResultMessage, got %T", messages[1])
	}
}
```

**Step 2: Write implementation**

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

// WithTransport sets a custom transport (for testing).
func WithTransport(t Transport) Option {
	return func(o *Options) {
		o.customTransport = t
	}
}
```

**Step 3: Run tests**

```bash
go test -run TestQuery_OneShot -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add Query function for one-shot queries"
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
			"duration_ms":    float64(100),
			"is_error":       false,
			"num_turns":      float64(1),
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

func TestQueryStream_Cancellation(t *testing.T) {
	transport := NewMockTransport()
	go func() {
		// Don't send result - let context cancel
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		}
		time.Sleep(100 * time.Millisecond)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	msgChan, errChan := QueryStream(ctx, "Hello", WithTransport(transport))

	// Drain messages
	for range msgChan {
	}

	// Should get context error
	err := <-errChan
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}
```

**Step 2: Write implementation**

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

**Step 3: Run tests**

```bash
go test -run TestQueryStream -v
```

Expected: PASS

**Step 4: Commit**

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
func TestClient_SendQuery(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Send query
	err := client.SendQuery("Hello")
	if err != nil {
		t.Errorf("SendQuery failed: %v", err)
	}

	// Verify query was written
	written := transport.Written()
	if len(written) < 1 {
		t.Error("expected query to be written")
	}
}

func TestClient_ReceiveMessage(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
		}
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	msg, err := client.ReceiveMessage()
	if err != nil {
		t.Errorf("ReceiveMessage failed: %v", err)
	}

	if _, ok := msg.(*AssistantMessage); !ok {
		t.Errorf("expected AssistantMessage, got %T", msg)
	}
}

func TestClient_ReceiveAll(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
		}
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		}
		transport.messages <- map[string]any{
			"type":    "result",
			"subtype": "success",
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.SendQuery("Hello"); err != nil {
		t.Fatal(err)
	}

	messages, err := client.ReceiveAll()
	if err != nil {
		t.Errorf("ReceiveAll failed: %v", err)
	}

	if len(messages) != 2 { // assistant + result
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}
```

**Step 2: Write implementation**

Add to `client.go`:

```go
// SendQuery sends a query in streaming mode.
func (c *Client) SendQuery(prompt string, sessionID ...string) error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		return &ConnectionError{Message: "not connected"}
	}
	q := c.query
	sid := c.sessionID
	c.mu.Unlock()

	if len(sessionID) > 0 {
		sid = sessionID[0]
	}
	if sid == "" {
		sid = "default"
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
		return msg, nil
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

// ReceiveResponse sends a query and receives all response messages.
func (c *Client) ReceiveResponse(prompt string) ([]Message, error) {
	if err := c.SendQuery(prompt); err != nil {
		return nil, err
	}
	return c.ReceiveAll()
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

// ResultReceived returns true if a result has been received.
func (c *Client) ResultReceived() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.query != nil {
		return c.query.ResultReceived()
	}
	return false
}

// LastResult returns the last result message.
func (c *Client) LastResult() *ResultMessage {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.query != nil {
		return c.query.LastResult()
	}
	return nil
}
```

**Step 3: Run tests**

```bash
go test -run "TestClient_SendQuery|TestClient_ReceiveMessage|TestClient_ReceiveAll" -v
```

Expected: PASS

**Step 4: Commit**

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
func TestWithClient(t *testing.T) {
	transport := NewMockTransport()

	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
		}
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		}
		transport.messages <- map[string]any{
			"type":    "result",
			"subtype": "success",
		}
	}()

	ctx := context.Background()
	var receivedMessages []Message

	err := WithClient(ctx, WithTransport(transport), func(c *Client) error {
		if err := c.SendQuery("Hello"); err != nil {
			return err
		}

		messages, err := c.ReceiveAll()
		if err != nil {
			return err
		}

		receivedMessages = messages
		return nil
	})

	if err != nil {
		t.Errorf("WithClient failed: %v", err)
	}

	if len(receivedMessages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(receivedMessages))
	}
}

func TestClient_Run(t *testing.T) {
	transport := NewMockTransport()

	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
		}
	}()

	client := NewClient(WithTransport(transport))
	ctx := context.Background()

	runCalled := false
	err := client.Run(ctx, func() error {
		runCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	if !runCalled {
		t.Error("run function was not called")
	}

	if client.IsConnected() {
		t.Error("client should be disconnected after Run")
	}
}
```

**Step 2: Write implementation**

Add to `client.go`:

```go
// ClientFunc is a function that uses a client.
type ClientFunc func(*Client) error

// WithClient creates a client, connects, runs the function, and ensures cleanup.
func WithClient(ctx context.Context, opts []Option, fn ClientFunc) error {
	client := NewClient(opts...)

	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()

	return fn(client)
}

// Run connects and runs a function with the client.
func (c *Client) Run(ctx context.Context, fn func() error) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	defer c.Close()
	return fn()
}

// Convenience function signature for WithClient
// WithClient(ctx, WithTransport(t), func(c *Client) error { ... })
func WithClientOpts(ctx context.Context, fn ClientFunc, opts ...Option) error {
	return WithClient(ctx, opts, fn)
}
```

**Step 3: Run tests**

```bash
go test -run "TestWithClient|TestClient_Run" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add context manager pattern"
```

---

## Task 7: Message Helper Methods

**Files:**
- Modify: `types.go`
- Create: `types_test.go`

**Step 1: Write failing test**

Create `types_test.go`:

```go
package sdk

import (
	"testing"
)

func TestAssistantMessage_Text(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{TextContent: "Hello "},
			&TextBlock{TextContent: "World"},
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
			&TextBlock{TextContent: "Let me help"},
			&ToolUseBlock{ID: "tool_1", Name: "Bash", ToolInput: map[string]any{"command": "ls"}},
			&ToolUseBlock{ID: "tool_2", Name: "Read", ToolInput: map[string]any{"path": "/tmp"}},
		},
	}

	tools := msg.ToolCalls()
	if len(tools) != 2 {
		t.Errorf("expected 2 tool calls, got %d", len(tools))
	}

	if tools[0].Name != "Bash" {
		t.Errorf("expected first tool Bash, got %s", tools[0].Name)
	}
}

func TestAssistantMessage_Thinking(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&ThinkingBlock{ThinkingContent: "Let me think about this..."},
			&TextBlock{TextContent: "Here's my answer"},
		},
	}

	thinking := msg.Thinking()
	if thinking != "Let me think about this..." {
		t.Errorf("got thinking %q", thinking)
	}
}

func TestResultMessage_IsSuccess(t *testing.T) {
	tests := []struct {
		msg  *ResultMessage
		want bool
	}{
		{&ResultMessage{Subtype: "success", IsError: false}, true},
		{&ResultMessage{Subtype: "error", IsError: true}, false},
		{&ResultMessage{Subtype: "success", IsError: true}, false},
	}

	for _, tt := range tests {
		got := tt.msg.IsSuccess()
		if got != tt.want {
			t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
		}
	}
}

func TestResultMessage_Cost(t *testing.T) {
	cost := 0.005
	msg := &ResultMessage{TotalCostUSD: &cost}

	if msg.Cost() != 0.005 {
		t.Errorf("got cost %f, want 0.005", msg.Cost())
	}

	msg2 := &ResultMessage{}
	if msg2.Cost() != 0 {
		t.Errorf("got cost %f, want 0", msg2.Cost())
	}
}

func TestUserMessage_Text(t *testing.T) {
	msg := &UserMessage{
		Content: []ContentBlock{
			&TextBlock{TextContent: "Hello Claude!"},
		},
	}

	if msg.Text() != "Hello Claude!" {
		t.Errorf("got %q, want Hello Claude!", msg.Text())
	}
}
```

**Step 2: Write implementation**

Add to `types.go`:

```go
import "strings"

// Text returns all text content concatenated.
func (m *AssistantMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if textBlock, ok := block.(*TextBlock); ok {
			parts = append(parts, textBlock.TextContent)
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
			return thinkingBlock.ThinkingContent
		}
	}
	return ""
}

// HasToolCalls returns true if the message contains tool calls.
func (m *AssistantMessage) HasToolCalls() bool {
	for _, block := range m.Content {
		if _, ok := block.(*ToolUseBlock); ok {
			return true
		}
	}
	return false
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

// Text returns all text content from user message.
func (m *UserMessage) Text() string {
	var parts []string
	for _, block := range m.Content {
		if textBlock, ok := block.(*TextBlock); ok {
			parts = append(parts, textBlock.TextContent)
		}
	}
	return strings.Join(parts, "")
}
```

**Step 3: Run tests**

```bash
go test -run "TestAssistantMessage|TestResultMessage|TestUserMessage" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add types.go types_test.go
git commit -m "feat: add message helper methods"
```

---

## Task 8: Async Iterator Pattern

**Files:**
- Modify: `client.go`
- Modify: `client_test.go`

**Step 1: Write failing test**

Add to `client_test.go`:

```go
func TestClient_Messages(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient()
	client.transport = transport

	go func() {
		time.Sleep(10 * time.Millisecond)
		transport.messages <- map[string]any{
			"type":    "system",
			"subtype": "init",
		}
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		}
		transport.messages <- map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "World!"},
				},
			},
		}
		transport.messages <- map[string]any{
			"type":    "result",
			"subtype": "success",
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.SendQuery("Hello"); err != nil {
		t.Fatal(err)
	}

	var texts []string
	for msg := range client.Messages() {
		if asst, ok := msg.(*AssistantMessage); ok {
			texts = append(texts, asst.Text())
		}
		if _, ok := msg.(*ResultMessage); ok {
			break
		}
	}

	if len(texts) != 2 {
		t.Errorf("expected 2 texts, got %d: %v", len(texts), texts)
	}
}
```

**Step 2: Write implementation**

Add to `client.go`:

```go
// Messages returns a channel that yields messages until closed or error.
// Use this for iterating over responses in streaming mode.
func (c *Client) Messages() <-chan Message {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		ch := make(chan Message)
		close(ch)
		return ch
	}
	q := c.query
	c.mu.Unlock()

	return q.Messages()
}

// RawMessages returns a channel of raw message maps.
func (c *Client) RawMessages() <-chan map[string]any {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		ch := make(chan map[string]any)
		close(ch)
		return ch
	}
	q := c.query
	c.mu.Unlock()

	return q.RawMessages()
}

// Errors returns the error channel.
func (c *Client) Errors() <-chan error {
	c.mu.Lock()
	if !c.connected || c.query == nil {
		c.mu.Unlock()
		ch := make(chan error, 1)
		close(ch)
		return ch
	}
	q := c.query
	c.mu.Unlock()

	return q.Errors()
}
```

**Step 3: Run tests**

```bash
go test -run TestClient_Messages -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add client.go client_test.go
git commit -m "feat: add async iterator pattern"
```

---

## Summary

After completing Plan 04, you have:

- [x] Client structure with complete options
- [x] MCP server registration
- [x] Hook registration (PreToolUse, PostToolUse, Stop)
- [x] CanUseTool callback
- [x] Connect method with initialization
- [x] Resume and Continue session support
- [x] Query function (one-shot)
- [x] QueryStream function
- [x] Streaming client methods (SendQuery, ReceiveMessage, ReceiveAll, ReceiveResponse)
- [x] Control methods (Interrupt, SetPermissionMode, SetModel, RewindFiles)
- [x] Context manager pattern (WithClient, Run)
- [x] Message helper methods (Text, ToolCalls, Thinking, IsSuccess, Cost)
- [x] Async iterator pattern (Messages, RawMessages, Errors)

**Key Features:**
- Complete hook integration at client level
- MCP server hosting with tool registration
- Session resume/continue support
- Rich message helper methods
- Multiple iteration patterns (ReceiveAll, Messages channel)
- Context manager for automatic cleanup

**Next:** Plan 05 - Integration & Examples
