package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
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
	transport.SendMessage(map[string]any{"type": "system", "subtype": "init"})

	select {
	case msg := <-query.Messages():
		if msg.MessageType() != "system" {
			t.Errorf("got type %v, want system", msg.MessageType())
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

	transport.SendMessage(map[string]any{"type": "test", "custom": true})

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
	transport.SendMessage(map[string]any{
		"type":       "result",
		"subtype":    "success",
		"session_id": "test_123",
	})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !query.ResultReceived() {
		t.Error("expected result to be received")
	}
}

func TestQuery_Close(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if err := query.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Should be able to close again without error
	if err := query.Close(); err != nil {
		t.Errorf("second Close failed: %v", err)
	}
}

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
		// Wait for the control request to be written
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for control request write")
			return
		}

		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"status": "ok"},
				},
			})
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
		// Wait for the control request to be written
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for control request write")
			return
		}

		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "error",
					"request_id": reqID,
					"error":      "something went wrong",
				},
			})
		}
	}()

	_, err := query.sendControlRequest(map[string]any{
		"subtype": "test",
	}, 5*time.Second)

	if err == nil {
		t.Error("expected error response")
	}
}

func TestQuery_SendControlRequest_NonStreaming(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, false) // non-streaming

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	_, err := query.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 100*time.Millisecond)

	if err == nil {
		t.Error("expected error for non-streaming mode")
	}
}

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
		// Wait for the control request to be written
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for init request write")
			return
		}

		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response": map[string]any{
						"commands": []any{"/help", "/clear"},
					},
				},
			})
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

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	hooks := map[types.HookEvent][]types.HookMatcher{
		types.HookPreToolUse: {
			{
				Matcher: map[string]any{"tool_name": "Bash"},
				Hooks: []types.HookCallback{
					func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
						cont := true
						return &types.HookOutput{Continue: &cont}, nil
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

func TestQuery_Initialize_NonStreaming(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, false) // non-streaming

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Non-streaming should return nil without error
	result, err := query.Initialize(nil)
	if err != nil {
		t.Errorf("Initialize failed: %v", err)
	}
	if result != nil {
		t.Error("expected nil result for non-streaming")
	}
}

func strPtr(s string) *string { return &s }

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

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
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

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	err := query.SetPermissionMode(types.PermissionBypass)
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

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	err := query.SetModel("claude-opus-4")
	if err != nil {
		t.Errorf("SetModel failed: %v", err)
	}
}

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
		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			// Verify user_message_id is in request
			request := req["request"].(map[string]any)
			if request["user_message_id"] != "msg_123" {
				t.Errorf("expected user_message_id msg_123, got %v", request["user_message_id"])
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
		}
	}()

	err := query.RewindFiles("msg_123")
	if err != nil {
		t.Errorf("RewindFiles failed: %v", err)
	}
}

func TestQuery_HandleHookCallback(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Register a hook callback - use atomic for thread-safe access
	var callbackCalled atomic.Bool
	query.hookMu.Lock()
	query.hookCallbacks["hook_1"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		callbackCalled.Store(true)
		cont := true
		return &types.HookOutput{Continue: &cont}, nil
	}
	query.hookMu.Unlock()

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Send hook callback request
	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_hook_1",
		"request": map[string]any{
			"subtype":     "hook_callback",
			"callback_id": "hook_1",
			"input": map[string]any{
				"tool_name": "Bash",
			},
		},
	})

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if !callbackCalled.Load() {
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
	query.hookCallbacks["hook_err"] = func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		return nil, fmt.Errorf("hook error")
	}
	query.hookMu.Unlock()

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_hook_err",
		"request": map[string]any{
			"subtype":     "hook_callback",
			"callback_id": "hook_err",
			"input":       map[string]any{},
		},
	})

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

func TestQuery_HandleCanUseTool(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	var called atomic.Bool
	query.SetCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
		called.Store(true)
		return &types.PermissionResultAllow{Behavior: "allow"}, nil
	})

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate can_use_tool request
	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_perm_1",
		"request": map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "Bash",
			"input":     map[string]any{"command": "ls"},
		},
	})

	// Wait for callback and response with timeout
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for response")
		case <-ticker.C:
			if called.Load() && len(transport.Written()) > 0 {
				goto done
			}
		}
	}
done:

	if !called.Load() {
		t.Error("canUseTool callback was not called")
	}

	// Verify response was sent
	written := transport.Written()
	if len(written) == 0 {
		t.Error("no response was written")
	}
}

func TestQuery_HandleCanUseTool_Deny(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	query.SetCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
		return &types.PermissionResultDeny{Behavior: "deny", Message: "not allowed"}, nil
	})

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_perm_2",
		"request": map[string]any{
			"subtype":   "can_use_tool",
			"tool_name": "Bash",
			"input":     map[string]any{"command": "rm -rf /"},
		},
	})

	time.Sleep(100 * time.Millisecond)

	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no response was written")
	}

	var resp map[string]any
	json.Unmarshal([]byte(written[0]), &resp)
	response := resp["response"].(map[string]any)
	respData := response["response"].(map[string]any)
	if respData["behavior"] != "deny" {
		t.Errorf("expected deny behavior, got %v", respData["behavior"])
	}
}

// Task 8: Stream Input and Message Sending Tests

func TestQuery_SendMessage(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	err := query.SendMessage(map[string]any{
		"type": "test",
		"data": "hello",
	})
	if err != nil {
		t.Errorf("SendMessage failed: %v", err)
	}

	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no message written")
	}

	var msg map[string]any
	if err := json.Unmarshal([]byte(written[0]), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg["type"] != "test" {
		t.Errorf("expected type test, got %v", msg["type"])
	}
}

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
	if err := json.Unmarshal([]byte(written[0]), &msg); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if msg["type"] != "user" {
		t.Errorf("expected type user, got %v", msg["type"])
	}

	if msg["session_id"] != "session_123" {
		t.Errorf("expected session_id session_123, got %v", msg["session_id"])
	}

	message := msg["message"].(map[string]any)
	if message["content"] != "Hello Claude!" {
		t.Errorf("expected content Hello Claude!, got %v", message["content"])
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

func TestQuery_StreamInput_ContextCancelled(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx, cancel := context.WithCancel(context.Background())
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Create input channel that won't be read
	input := make(chan map[string]any)

	// Cancel context before streaming
	cancel()

	// Give goroutine time to see cancellation
	time.Sleep(50 * time.Millisecond)

	err := query.StreamInput(input)
	if err == nil {
		t.Error("expected context cancelled error")
	}
}

// Task 10: MCP Tool Call Handling Tests

func TestQuery_RegisterMCPServer(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	server := types.NewMCPServerBuilder("test-server").
		WithTool("echo", "Echoes input", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "hello"}},
			}, nil
		}).
		Build()

	query.RegisterMCPServer(server)

	// Verify server was registered
	if _, ok := query.mcpServers["test-server"]; !ok {
		t.Error("expected server to be registered")
	}
}

func TestQuery_UnregisterMCPServer(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	server := types.NewMCPServerBuilder("test-server").Build()
	query.RegisterMCPServer(server)
	query.UnregisterMCPServer("test-server")

	if _, ok := query.mcpServers["test-server"]; ok {
		t.Error("expected server to be unregistered")
	}
}

func TestQuery_MCPToolCall(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Register an MCP server
	server := types.NewMCPServerBuilder("test-server").
		WithTool("echo", "Echoes input", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{"type": "string"},
			},
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			msg := args["message"].(string)
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: msg}},
			}, nil
		}).
		Build()

	query.RegisterMCPServer(server)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate MCP tool call request
	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_mcp_1",
		"request": map[string]any{
			"subtype":     "mcp_tool_call",
			"server_name": "test-server",
			"tool_name":   "echo",
			"input":       map[string]any{"message": "hello"},
		},
	})

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

func TestQuery_MCPToolCall_ServerNotFound(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate MCP tool call request for non-existent server
	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_mcp_2",
		"request": map[string]any{
			"subtype":     "mcp_tool_call",
			"server_name": "nonexistent",
			"tool_name":   "echo",
			"input":       map[string]any{},
		},
	})

	time.Sleep(100 * time.Millisecond)

	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no response written")
	}

	var resp map[string]any
	json.Unmarshal([]byte(written[0]), &resp)
	response := resp["response"].(map[string]any)
	if response["subtype"] != "error" {
		t.Errorf("expected error response, got %v", response["subtype"])
	}
}

func TestQuery_MCPToolCall_ToolNotFound(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	server := types.NewMCPServerBuilder("test-server").Build()
	query.RegisterMCPServer(server)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Simulate MCP tool call request for non-existent tool
	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_mcp_3",
		"request": map[string]any{
			"subtype":     "mcp_tool_call",
			"server_name": "test-server",
			"tool_name":   "nonexistent",
			"input":       map[string]any{},
		},
	})

	time.Sleep(100 * time.Millisecond)

	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("no response written")
	}

	var resp map[string]any
	json.Unmarshal([]byte(written[0]), &resp)
	response := resp["response"].(map[string]any)
	if response["subtype"] != "error" {
		t.Errorf("expected error response, got %v", response["subtype"])
	}
}

// TestQuery_SendControlResponse_MarshalError verifies that json.Marshal errors
// are handled gracefully (nothing written to transport, no panic).
func TestQuery_SendControlResponse_MarshalError(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	// Create data with an unmarshallable type (channel)
	// This will cause json.Marshal to fail
	unmarshallableData := map[string]any{
		"channel": make(chan int),
	}

	// Call sendControlResponse with unmarshallable data
	// Should not panic and should not write anything to transport
	query.sendControlResponse("test_req_id", unmarshallableData, nil)

	// With proper error handling, nothing should be written when marshal fails
	written := transport.Written()
	if len(written) > 0 {
		t.Errorf("expected no data written when marshal fails, got %d items", len(written))
	}
}

