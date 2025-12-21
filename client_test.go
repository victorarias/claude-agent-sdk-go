package sdk

import (
	"context"
	"encoding/json"
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
		}, func(args map[string]any) (*MCPToolResult, error) {
			return &MCPToolResult{
				Content: []MCPContent{{Type: "text", Text: "hello"}},
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
			func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
				preToolUseCalled = true
				cont := true
				return &HookOutput{Continue: &cont}, nil
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
			func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
				return &HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[HookPostToolUse]) != 1 {
		t.Errorf("expected 1 post tool use hook, got %d", len(client.hooks[HookPostToolUse]))
	}
}

func TestClientWithStopHook(t *testing.T) {
	client := NewClient(
		WithStopHook(
			nil, // No matcher
			func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
				return &HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[HookStop]) != 1 {
		t.Errorf("expected 1 stop hook, got %d", len(client.hooks[HookStop]))
	}
}

func TestClientWithCanUseTool(t *testing.T) {
	called := false
	client := NewClient(
		WithCanUseTool(func(toolName string, input map[string]any, ctx *ToolPermissionContext) (PermissionResult, error) {
			called = true
			return &PermissionResultAllow{Behavior: "allow"}, nil
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
		WithHookTimeout(HookPreToolUse, map[string]any{"tool_name": "Bash"}, timeout,
			func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
				return &HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[HookPreToolUse]) != 1 {
		t.Errorf("expected 1 hook, got %d", len(client.hooks[HookPreToolUse]))
	}

	if client.hooks[HookPreToolUse][0].Timeout == nil || *client.hooks[HookPreToolUse][0].Timeout != timeout {
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
