# Plan 03: Query/Control Protocol

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement the Query layer that handles bidirectional control protocol on top of Transport.

**Architecture:** Use channels for control request/response routing. Use goroutines for concurrent message handling. Maintain pending requests map with request IDs.

**Tech Stack:** Go 1.21+, sync, encoding/json

---

## Task 1: Query Structure

**Files:**
- Create: `query.go`
- Create: `query_test.go`

**Step 1: Write failing test**

Create `query_test.go`:

```go
package sdk

import (
	"context"
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
	transport.messages <- map[string]any{"type": "test"}

	select {
	case msg := <-query.Messages():
		if msg["type"] != "test" {
			t.Errorf("got type %v, want test", msg["type"])
		}
	case <-time.After(time.Second):
		t.Error("timeout waiting for message")
	}

	query.Close()
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestNewQuery -v
```

Expected: FAIL - Query not defined

**Step 3: Write implementation**

Create `query.go`:

```go
package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
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
	hookCallbacks map[string]HookCallback
	nextCallbackID atomic.Uint64

	// Permission callback
	canUseTool CanUseToolCallback

	// Message channels
	messages chan map[string]any
	errors   chan error

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool

	// Initialization result
	initResult map[string]any
}

// NewQuery creates a new Query.
func NewQuery(transport Transport, streaming bool) *Query {
	return &Query{
		transport:       transport,
		streaming:       streaming,
		pendingRequests: make(map[string]chan map[string]any),
		hookCallbacks:   make(map[string]HookCallback),
		messages:        make(chan map[string]any, 100),
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

// Messages returns the channel of SDK messages.
func (q *Query) Messages() <-chan map[string]any {
	return q.messages
}

// Errors returns the channel of errors.
func (q *Query) Errors() <-chan error {
	return q.errors
}

// Close stops the query.
func (q *Query) Close() error {
	if q.closed.Swap(true) {
		return nil // Already closed
	}

	if q.cancel != nil {
		q.cancel()
	}

	close(q.messages)
	q.wg.Wait()

	return nil
}

// routeMessages reads from transport and routes control vs SDK messages.
func (q *Query) routeMessages() {
	defer q.wg.Done()

	for msg := range q.transport.Messages() {
		select {
		case <-q.ctx.Done():
			return
		default:
		}

		msgType, _ := msg["type"].(string)

		switch msgType {
		case "control_response":
			q.handleControlResponse(msg)
		case "control_request":
			go q.handleControlRequest(msg)
		case "control_cancel_request":
			// TODO: Implement cancellation
		default:
			// Regular SDK message
			select {
			case q.messages <- msg:
			case <-q.ctx.Done():
				return
			}
		}
	}
}
```

**Step 4: Run tests**

```bash
go test -run "TestNewQuery|TestQuery_Start" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add Query structure with message routing"
```

---

## Task 2: Control Request/Response

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
		// Find the request ID from written data
		if len(transport.written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(transport.written[0]), &req)
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
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_SendControlRequest -v
```

Expected: FAIL - sendControlRequest not defined

**Step 3: Write implementation**

Add to `query.go`:

```go
import "time"

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
		// Send error as response
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

**Step 4: Run tests**

```bash
go test -run TestQuery_SendControlRequest -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add control request/response handling"
```

---

## Task 3: Initialize Method

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
		if len(transport.written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(transport.written[0]), &req)
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
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_Initialize -v
```

Expected: FAIL - Initialize not defined

**Step 3: Write implementation**

Add to `query.go`:

```go
// Initialize sends the initialization request to the CLI.
func (q *Query) Initialize(hooks map[HookEvent][]HookMatcher) (map[string]any, error) {
	if !q.streaming {
		return nil, nil
	}

	// Build hooks configuration
	hooksConfig := make(map[string]any)
	for event, matchers := range hooks {
		if len(matchers) == 0 {
			continue
		}

		var matcherConfigs []map[string]any
		for _, matcher := range matchers {
			callbackIDs := make([]string, len(matcher.Hooks))
			for i, callback := range matcher.Hooks {
				callbackID := fmt.Sprintf("hook_%d", q.nextCallbackID.Add(1))
				q.hookCallbacks[callbackID] = callback
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
		hooksConfig[string(event)] = matcherConfigs
	}

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

	q.initResult = result
	return result, nil
}

// InitResult returns the initialization result.
func (q *Query) InitResult() map[string]any {
	return q.initResult
}
```

**Step 4: Run tests**

```bash
go test -run TestQuery_Initialize -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add Initialize method"
```

---

## Task 4: Interrupt Method

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

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

	// Simulate response
	go func() {
		time.Sleep(10 * time.Millisecond)
		if len(transport.written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(transport.written[0]), &req)
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

	// Verify request was sent
	if len(transport.written) == 0 {
		t.Error("no request was written")
	}

	var req map[string]any
	json.Unmarshal([]byte(transport.written[0]), &req)
	request := req["request"].(map[string]any)
	if request["subtype"] != "interrupt" {
		t.Errorf("expected interrupt subtype, got %v", request["subtype"])
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_Interrupt -v
```

Expected: FAIL - Interrupt not defined

**Step 3: Write implementation**

Add to `query.go`:

```go
// Interrupt sends an interrupt signal to the CLI.
func (q *Query) Interrupt() error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 30*time.Second)
	return err
}
```

**Step 4: Run tests**

```bash
go test -run TestQuery_Interrupt -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add Interrupt method"
```

---

## Task 5: SetPermissionMode Method

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
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
		if len(transport.written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(transport.written[0]), &req)
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

	var req map[string]any
	json.Unmarshal([]byte(transport.written[0]), &req)
	request := req["request"].(map[string]any)
	if request["mode"] != "bypassPermissions" {
		t.Errorf("expected bypassPermissions, got %v", request["mode"])
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_SetPermissionMode -v
```

Expected: FAIL - SetPermissionMode not defined

**Step 3: Write implementation**

Add to `query.go`:

```go
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
```

**Step 4: Run tests**

```bash
go test -run TestQuery_SetPermissionMode -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add SetPermissionMode and SetModel methods"
```

---

## Task 6: Handle Hook Callbacks

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
	query.hookCallbacks["hook_1"] = func(input any, toolUseID *string) (*HookOutput, error) {
		callbackCalled = true
		return &HookOutput{Continue: true}, nil
	}

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
	if len(transport.written) == 0 {
		t.Error("no response was written")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_HandleHookCallback -v
```

Expected: FAIL - handleControlRequest not handling hook_callback

**Step 3: Write implementation**

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

	callback, exists := q.hookCallbacks[callbackID]
	if !exists {
		return nil, fmt.Errorf("hook callback not found: %s", callbackID)
	}

	output, err := callback(input, toolUseID)
	if err != nil {
		return nil, err
	}

	// Convert HookOutput to response
	result := make(map[string]any)
	if output != nil {
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
	}

	return result, nil
}

// handleCanUseTool handles tool permission requests.
func (q *Query) handleCanUseTool(request map[string]any) (map[string]any, error) {
	if q.canUseTool == nil {
		return nil, fmt.Errorf("canUseTool callback not provided")
	}

	toolName, _ := request["tool_name"].(string)
	input, _ := request["input"].(map[string]any)
	suggestions, _ := request["permission_suggestions"].([]any)

	ctx := &ToolPermissionContext{}
	// Convert suggestions if present
	_ = suggestions // TODO: Convert to PermissionUpdate slice

	result, err := q.canUseTool(toolName, input, ctx)
	if err != nil {
		return nil, err
	}

	// Convert result to response
	switch r := result.(type) {
	case *PermissionResultAllow:
		resp := map[string]any{
			"behavior": "allow",
		}
		if r.UpdatedInput != nil {
			resp["updatedInput"] = r.UpdatedInput
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

**Step 4: Run tests**

```bash
go test -run TestQuery_HandleHookCallback -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add hook callback handling"
```

---

## Task 7: Stream Input

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
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

	// Verify messages were written
	if len(transport.written) != 2 {
		t.Errorf("expected 2 messages written, got %d", len(transport.written))
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_StreamInput -v
```

Expected: FAIL - StreamInput not defined

**Step 3: Write implementation**

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

		data, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := q.transport.Write(string(data)); err != nil {
			return fmt.Errorf("failed to write message: %w", err)
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
```

**Step 4: Run tests**

```bash
go test -run TestQuery_StreamInput -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add StreamInput and SendMessage methods"
```

---

## Task 8: SetCanUseTool

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
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

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_SetCanUseTool -v
```

Expected: FAIL - SetCanUseTool not defined

**Step 3: Write implementation**

Add to `query.go`:

```go
// SetCanUseTool sets the tool permission callback.
func (q *Query) SetCanUseTool(callback CanUseToolCallback) {
	q.canUseTool = callback
}
```

**Step 4: Run tests**

```bash
go test -run TestQuery_SetCanUseTool -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add SetCanUseTool method"
```

---

## Task 9: RewindFiles Method

**Files:**
- Modify: `query.go`
- Modify: `query_test.go`

**Step 1: Write failing test**

Add to `query_test.go`:

```go
func TestQuery_RewindFiles(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		time.Sleep(10 * time.Millisecond)
		if len(transport.written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(transport.written[0]), &req)
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

	err := query.RewindFiles("msg_123")
	if err != nil {
		t.Errorf("RewindFiles failed: %v", err)
	}

	var req map[string]any
	json.Unmarshal([]byte(transport.written[0]), &req)
	request := req["request"].(map[string]any)
	if request["user_message_id"] != "msg_123" {
		t.Errorf("expected msg_123, got %v", request["user_message_id"])
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestQuery_RewindFiles -v
```

Expected: FAIL - RewindFiles not defined

**Step 3: Write implementation**

Add to `query.go`:

```go
// RewindFiles rewinds tracked files to a specific user message.
func (q *Query) RewindFiles(userMessageID string) error {
	_, err := q.sendControlRequest(map[string]any{
		"subtype":         "rewind_files",
		"user_message_id": userMessageID,
	}, 30*time.Second)
	return err
}
```

**Step 4: Run tests**

```bash
go test -run TestQuery_RewindFiles -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add query.go query_test.go
git commit -m "feat: add RewindFiles method"
```

---

## Summary

After completing Plan 03, you have:

- [x] Query structure with message routing
- [x] Control request/response handling
- [x] Initialize method
- [x] Interrupt method
- [x] SetPermissionMode and SetModel methods
- [x] Hook callback handling
- [x] Stream input methods
- [x] SetCanUseTool method
- [x] RewindFiles method

**Next:** Plan 04 - Client API
