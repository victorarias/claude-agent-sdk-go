package sdk

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient(
		types.WithModel("claude-opus-4"),
		types.WithMaxTurns(10),
		types.WithPermissionMode(types.PermissionBypass),
		types.WithSystemPrompt("You are helpful"),
	)

	if client.options.Model != "claude-opus-4" {
		t.Errorf("got model %q, want %q", client.options.Model, "claude-opus-4")
	}
	if client.options.MaxTurns != 10 {
		t.Errorf("got maxTurns %d, want %d", client.options.MaxTurns, 10)
	}
	if client.options.PermissionMode != types.PermissionBypass {
		t.Errorf("got permission mode %v, want %v", client.options.PermissionMode, types.PermissionBypass)
	}
}

func TestClientWithMCPServers(t *testing.T) {
	server := types.NewMCPServerBuilder("test-server").
		WithTool("echo", "Echoes input", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "hello"}},
			}, nil
		}).
		Build()

	client := NewClient(
		WithClientMCPServer(server),
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
			func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				preToolUseCalled = true
				cont := true
				return &types.HookOutput{Continue: &cont}, nil
			},
		),
	)

	if len(client.hooks) != 1 {
		t.Errorf("expected 1 hook event, got %d", len(client.hooks))
	}
	_ = preToolUseCalled // Used when hook is invoked
}

func TestClientWithPostToolUseHook(t *testing.T) {
	client := NewClient(
		WithPostToolUseHook(
			map[string]any{"tool_name": "Read"},
			func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				return &types.HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[types.HookPostToolUse]) != 1 {
		t.Errorf("expected 1 post tool use hook, got %d", len(client.hooks[types.HookPostToolUse]))
	}
}

func TestClientWithStopHook(t *testing.T) {
	client := NewClient(
		WithStopHook(
			nil, // No matcher
			func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				return &types.HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[types.HookStop]) != 1 {
		t.Errorf("expected 1 stop hook, got %d", len(client.hooks[types.HookStop]))
	}
}

func TestClientWithCanUseTool(t *testing.T) {
	called := false
	client := NewClient(
		WithCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
			called = true
			return &types.PermissionResultAllow{Behavior: "allow"}, nil
		}),
	)

	if client.canUseTool == nil {
		t.Error("canUseTool callback not set")
	}
	_ = called
}

func TestClient_SessionID(t *testing.T) {
	client := NewClient()

	// Initially empty
	if client.SessionID() != "" {
		t.Errorf("expected empty session ID, got %q", client.SessionID())
	}
}

func TestClient_IsConnected(t *testing.T) {
	client := NewClient()

	// Initially not connected
	if client.IsConnected() {
		t.Error("expected client to not be connected initially")
	}
}

func TestClientWithHookTimeout(t *testing.T) {
	timeout := 5.0
	client := NewClient(
		WithHookTimeout(types.HookPreToolUse, map[string]any{"tool_name": "Bash"}, timeout,
			func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				return &types.HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[types.HookPreToolUse]) != 1 {
		t.Errorf("expected 1 hook, got %d", len(client.hooks[types.HookPreToolUse]))
	}

	if client.hooks[types.HookPreToolUse][0].Timeout == nil || *client.hooks[types.HookPreToolUse][0].Timeout != timeout {
		t.Error("expected hook timeout to be set")
	}
}

func TestClient_Connect(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(WithTransport(transport))

	// Respond to control requests
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			written := transport.Written()
			if len(written) == 0 {
				continue
			}

			// Parse the last request to get the request_id
			var req map[string]any
			if err := json.Unmarshal([]byte(written[len(written)-1]), &req); err != nil {
				continue
			}
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response": map[string]any{
						"session_id": "test_session_123",
					},
				},
			})
			return
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

	if client.SessionID() != "test_session_123" {
		t.Errorf("expected session ID 'test_session_123', got '%s'", client.SessionID())
	}

	client.Close()
}

func TestClient_ConnectWithPrompt(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(WithTransport(transport))

	// Non-streaming mode doesn't require Initialize
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
		WithTransport(transport),
	)

	// Respond to control requests
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
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response": map[string]any{
						"session_id": "previous_session_id",
					},
				},
			})
			return
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

func TestQuery_OneShot(t *testing.T) {
	transport := NewMockTransport()
	go func() {
		transport.SendMessage(map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
				"model": "claude-test",
			},
		})
		transport.SendMessage(map[string]any{
			"type":            "result",
			"subtype":         "success",
			"duration_ms":     float64(100),
			"duration_api_ms": float64(80),
			"is_error":        false,
			"num_turns":       float64(1),
			"session_id":      "test_123",
			"total_cost_usd":  float64(0.001),
		})
	}()

	ctx := context.Background()
	messages, err := RunQuery(ctx, "Hello", WithTransport(transport))

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

func TestQueryStream(t *testing.T) {
	transport := NewMockTransport()
	go func() {
		transport.SendMessage(map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
				"model": "claude-test",
			},
		})
		transport.SendMessage(map[string]any{
			"type":        "result",
			"subtype":     "success",
			"duration_ms": float64(100),
			"is_error":    false,
			"num_turns":   float64(1),
			"session_id":  "test_123",
		})
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
		// Send a message but no result - let context cancel
		transport.SendMessage(map[string]any{
			"type": "assistant",
			"message": map[string]any{
				"content": []any{
					map[string]any{"type": "text", "text": "Hello!"},
				},
			},
		})
		// Wait then close to simulate connection dropping
		time.Sleep(100 * time.Millisecond)
		transport.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	msgChan, errChan := QueryStream(ctx, "Hello", WithTransport(transport))

	// Drain messages (should get at least one before timeout)
	var count int
	for range msgChan {
		count++
	}

	// Should get context error or nil (if channel closed first)
	select {
	case err := <-errChan:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("expected nil or DeadlineExceeded, got %v", err)
		}
	default:
	}

	if count < 1 {
		t.Errorf("expected at least 1 message, got %d", count)
	}
}

func TestClient_SendQuery(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(WithTransport(transport))

	// Respond to initialize
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
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "test_session"},
				},
			})
			return
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
}

func TestClient_ReceiveMessage(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(WithTransport(transport))

	// Respond to initialize and send assistant message
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
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			// Send init response
			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "test_session"},
				},
			})

			// Then send assistant message
			time.Sleep(10 * time.Millisecond)
			transport.SendMessage(map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello!"},
					},
				},
			})
			return
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
	client := NewClient(WithTransport(transport))

	// Respond to initialize and send messages
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
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			// Send init response
			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "test_session"},
				},
			})

			// Then send assistant and result messages
			time.Sleep(10 * time.Millisecond)
			transport.SendMessage(map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello!"},
					},
				},
			})
			transport.SendMessage(map[string]any{
				"type":    "result",
				"subtype": "success",
			})
			return
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	messages, err := client.ReceiveAll()
	if err != nil {
		t.Errorf("ReceiveAll failed: %v", err)
	}

	if len(messages) != 2 { // assistant + result
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestWithClient(t *testing.T) {
	transport := NewMockTransport()

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
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "test_session"},
				},
			})

			time.Sleep(10 * time.Millisecond)
			transport.SendMessage(map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello!"},
					},
				},
			})
			transport.SendMessage(map[string]any{
				"type":    "result",
				"subtype": "success",
			})
			return
		}
	}()

	ctx := context.Background()
	var receivedMessages []Message

	err := WithClient(ctx, []Option{WithTransport(transport)}, func(c *Client) error {
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
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})
			return
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

func TestClient_Messages(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(WithTransport(transport))

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
			reqID, ok := req["request_id"].(string)
			if !ok {
				continue
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "test_session"},
				},
			})

			time.Sleep(10 * time.Millisecond)
			transport.SendMessage(map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "Hello!"},
					},
				},
			})
			transport.SendMessage(map[string]any{
				"type": "assistant",
				"message": map[string]any{
					"content": []any{
						map[string]any{"type": "text", "text": "World!"},
					},
				},
			})
			transport.SendMessage(map[string]any{
				"type":    "result",
				"subtype": "success",
			})
			return
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
