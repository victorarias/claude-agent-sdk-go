// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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

	// Wait for result to be processed
	select {
	case <-query.WaitForFirstResult():
		// Result received successfully
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for result to be processed")
	}

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

func TestQuery_Initialize_WithAgents(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)
	query.SetAgents(map[string]types.AgentDefinition{
		"researcher": {
			Description: "Research assistant",
			Prompt:      "Gather facts",
			Tools:       []string{"Read"},
			DisallowedTools: []string{
				"Bash",
			},
			Model:                              types.AgentModelSonnet,
			MCPServers:                         []any{"docs-server"},
			CriticalSystemReminderExperimental: "Stay on mission",
			Skills:                             []string{"summarize"},
			MaxTurns:                           3,
		},
	})

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for init request write")
			return
		}
		written := transport.Written()
		if len(written) == 0 {
			t.Error("expected at least one write")
			return
		}

		var req map[string]any
		if err := json.Unmarshal([]byte(written[0]), &req); err != nil {
			t.Errorf("failed to parse request: %v", err)
			return
		}
		reqID := req["request_id"].(string)
		request := req["request"].(map[string]any)
		agents, ok := request["agents"].(map[string]any)
		if !ok || len(agents) != 1 {
			t.Errorf("expected agents payload in initialize request, got %v", request["agents"])
		}
		researcher, ok := agents["researcher"].(map[string]any)
		if !ok {
			t.Fatalf("expected researcher agent in payload, got %v", agents["researcher"])
		}
		if researcher["description"] != "Research assistant" {
			t.Errorf("expected description in payload, got %v", researcher["description"])
		}
		if researcher["prompt"] != "Gather facts" {
			t.Errorf("expected prompt in payload, got %v", researcher["prompt"])
		}
		if researcher["model"] != string(types.AgentModelSonnet) {
			t.Errorf("expected model in payload, got %v", researcher["model"])
		}
		if _, ok := researcher["disallowedTools"]; !ok {
			t.Error("expected disallowedTools in payload")
		}
		if _, ok := researcher["mcpServers"]; !ok {
			t.Error("expected mcpServers in payload")
		}
		if researcher["criticalSystemReminder_EXPERIMENTAL"] != "Stay on mission" {
			t.Errorf("expected criticalSystemReminder_EXPERIMENTAL in payload, got %v", researcher["criticalSystemReminder_EXPERIMENTAL"])
		}
		if _, ok := researcher["skills"]; !ok {
			t.Error("expected skills in payload")
		}
		if maxTurns, ok := researcher["maxTurns"].(float64); !ok || maxTurns != 3 {
			t.Errorf("expected maxTurns=3 in payload, got %v", researcher["maxTurns"])
		}

		transport.SendMessage(map[string]any{
			"type": "control_response",
			"response": map[string]any{
				"subtype":    "success",
				"request_id": reqID,
				"response":   map[string]any{},
			},
		})
	}()

	if _, err := query.Initialize(nil); err != nil {
		t.Fatalf("Initialize failed: %v", err)
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

func TestQuery_Interrupt(t *testing.T) {
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
			t.Error("timeout waiting for interrupt request write")
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
		// Wait for the control request to be written
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for permission mode request write")
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
		// Wait for the control request to be written
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for set model request write")
			return
		}

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

func TestQuery_SetMaxThinkingTokens(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for set max thinking tokens request write")
			return
		}

		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			request := req["request"].(map[string]any)
			if request["max_thinking_tokens"] != float64(4096) {
				t.Errorf("expected max_thinking_tokens 4096, got %v", request["max_thinking_tokens"])
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

	maxTokens := 4096
	if err := query.SetMaxThinkingTokens(&maxTokens); err != nil {
		t.Errorf("SetMaxThinkingTokens failed: %v", err)
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
		// Wait for the control request to be written
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for rewind files request write")
			return
		}

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

func TestQuery_RewindFilesWithOptions_DryRun(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for rewind files request write")
			return
		}

		written := transport.Written()
		if len(written) > 0 {
			var req map[string]any
			json.Unmarshal([]byte(written[0]), &req)
			reqID := req["request_id"].(string)

			request := req["request"].(map[string]any)
			if request["dry_run"] != true {
				t.Errorf("expected dry_run=true, got %v", request["dry_run"])
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response": map[string]any{
						"canRewind":    true,
						"filesChanged": []any{"a.go"},
						"insertions":   10,
						"deletions":    2,
					},
				},
			})
		}
	}()

	dryRun := true
	result, err := query.RewindFilesWithOptions("msg_123", &dryRun)
	if err != nil {
		t.Fatalf("RewindFilesWithOptions failed: %v", err)
	}
	if !result.CanRewind {
		t.Fatalf("expected CanRewind=true, got %+v", result)
	}
	if len(result.FilesChanged) != 1 || result.FilesChanged[0] != "a.go" {
		t.Fatalf("unexpected files changed: %+v", result.FilesChanged)
	}
}

func TestQuery_GetMCPStatus(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for mcp status request write")
			return
		}

		written := transport.Written()
		if len(written) == 0 {
			t.Error("expected mcp status request")
			return
		}

		var req map[string]any
		if err := json.Unmarshal([]byte(written[0]), &req); err != nil {
			t.Errorf("failed to parse request: %v", err)
			return
		}
		reqID := req["request_id"].(string)

		transport.SendMessage(map[string]any{
			"type": "control_response",
			"response": map[string]any{
				"subtype":    "success",
				"request_id": reqID,
				"response": map[string]any{
					"mcpServers": []any{
						map[string]any{"name": "calc", "status": "connected"},
					},
				},
			},
		})
	}()

	status, err := query.GetMCPStatus()
	if err != nil {
		t.Fatalf("GetMCPStatus failed: %v", err)
	}
	if status["mcpServers"] == nil {
		t.Errorf("expected mcpServers in response, got %v", status)
	}
}

func TestQuery_MCPServerControlRequests(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	expectedSubtypes := map[string]bool{
		"mcp_reconnect":   false,
		"mcp_toggle":      false,
		"mcp_set_servers": false,
	}

	go func() {
		deadline := time.After(2 * time.Second)
		for {
			select {
			case <-deadline:
				return
			default:
			}
			if !transport.WaitForWrite(100 * time.Millisecond) {
				continue
			}
			written := transport.Written()
			if len(written) == 0 {
				continue
			}
			var req map[string]any
			if err := json.Unmarshal([]byte(written[len(written)-1]), &req); err != nil {
				continue
			}
			reqID, _ := req["request_id"].(string)
			request, _ := req["request"].(map[string]any)
			subtype, _ := request["subtype"].(string)
			if _, ok := expectedSubtypes[subtype]; ok {
				expectedSubtypes[subtype] = true
			}

			response := map[string]any{}
			if subtype == "mcp_set_servers" {
				response = map[string]any{
					"added":   []any{"calc"},
					"removed": []any{},
					"errors":  map[string]any{},
				}
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   response,
				},
			})
		}
	}()

	if err := query.ReconnectMCPServer("calc"); err != nil {
		t.Fatalf("ReconnectMCPServer failed: %v", err)
	}
	if err := query.ToggleMCPServer("calc", false); err != nil {
		t.Fatalf("ToggleMCPServer failed: %v", err)
	}
	result, err := query.SetMCPServers(map[string]any{
		"calc": map[string]any{"type": "stdio", "command": "node"},
	})
	if err != nil {
		t.Fatalf("SetMCPServers failed: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "calc" {
		t.Fatalf("unexpected set servers result: %+v", result)
	}

	for subtype, seen := range expectedSubtypes {
		if !seen {
			t.Errorf("expected control subtype %s to be sent", subtype)
		}
	}
}

func TestQuery_SetMCPServers_ReconcilesSDKServers(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Start with an existing SDK-hosted MCP server.
	initial := types.NewMCPServerBuilder("initial").Build()
	query.RegisterMCPServer(initial)

	replacement := types.NewMCPServerBuilder("replacement").
		WithTool("echo", "Echo tool", map[string]any{
			"type": "object",
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "ok"}},
			}, nil
		}).
		Build()

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	go func() {
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for mcp_set_servers request")
			return
		}
		written := transport.Written()
		if len(written) == 0 {
			t.Error("expected control request write")
			return
		}

		var req map[string]any
		if err := json.Unmarshal([]byte(written[len(written)-1]), &req); err != nil {
			t.Errorf("failed to parse request: %v", err)
			return
		}
		reqID, _ := req["request_id"].(string)
		request, _ := req["request"].(map[string]any)
		servers, _ := request["servers"].(map[string]any)

		// Replacement SDK server must be represented as type=sdk,name=<server>.
		replacementConfig, ok := servers["replacement"].(map[string]any)
		if !ok {
			t.Errorf("expected replacement server in request payload, got %v", servers["replacement"])
		} else {
			if replacementConfig["type"] != "sdk" {
				t.Errorf("expected replacement type sdk, got %v", replacementConfig["type"])
			}
			if replacementConfig["name"] != "replacement" {
				t.Errorf("expected replacement name field, got %v", replacementConfig["name"])
			}
		}

		// External process server should pass through.
		if _, ok := servers["remote"]; !ok {
			t.Errorf("expected remote server in request payload, got %v", servers)
		}

		transport.SendMessage(map[string]any{
			"type": "control_response",
			"response": map[string]any{
				"subtype":    "success",
				"request_id": reqID,
				"response": map[string]any{
					"added":   []any{"replacement", "remote"},
					"removed": []any{"initial"},
					"errors":  map[string]any{},
				},
			},
		})
	}()

	_, err := query.SetMCPServers(map[string]any{
		"replacement": replacement,
		"remote": map[string]any{
			"type":    "stdio",
			"command": "node",
			"args":    []string{"server.js"},
		},
	})
	if err != nil {
		t.Fatalf("SetMCPServers failed: %v", err)
	}

	// Local SDK map should now contain only replacement.
	query.mcpServersMu.RLock()
	_, hasInitial := query.mcpServers["initial"]
	_, hasReplacement := query.mcpServers["replacement"]
	query.mcpServersMu.RUnlock()
	if hasInitial {
		t.Fatal("expected initial server to be removed from local SDK MCP registry")
	}
	if !hasReplacement {
		t.Fatal("expected replacement server to be registered in local SDK MCP registry")
	}

	// Verify bridged MCP messages route to the replacement server.
	resp, err := query.handleMCPMessage("replacement", map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
	})
	if err != nil {
		t.Fatalf("handleMCPMessage failed for replacement server: %v", err)
	}
	if resp == nil {
		t.Fatal("expected mcp response from replacement server")
	}
}

func TestQuery_InitializationResultHelpers(t *testing.T) {
	query := NewQuery(NewMockTransport(), true)
	query.initResult = map[string]any{
		"commands": []any{
			map[string]any{
				"name":         "review",
				"description":  "Review code",
				"argumentHint": "<path>",
			},
		},
		"models": []any{
			map[string]any{
				"value":       "claude-sonnet-4-5",
				"displayName": "Sonnet 4.5",
				"description": "Balanced",
			},
		},
		"account": map[string]any{
			"email":        "user@example.com",
			"organization": "acme",
		},
	}

	init, err := query.InitializationResult()
	if err != nil {
		t.Fatalf("InitializationResult failed: %v", err)
	}
	if len(init.Commands) != 1 || init.Commands[0].Name != "review" {
		t.Fatalf("unexpected commands: %+v", init.Commands)
	}

	commands, err := query.SupportedCommands()
	if err != nil {
		t.Fatalf("SupportedCommands failed: %v", err)
	}
	if len(commands) != 1 {
		t.Fatalf("unexpected supported commands: %+v", commands)
	}

	models, err := query.SupportedModels()
	if err != nil {
		t.Fatalf("SupportedModels failed: %v", err)
	}
	if len(models) != 1 || models[0].Value != "claude-sonnet-4-5" {
		t.Fatalf("unexpected supported models: %+v", models)
	}

	account, err := query.AccountInfo()
	if err != nil {
		t.Fatalf("AccountInfo failed: %v", err)
	}
	if account.Email != "user@example.com" {
		t.Fatalf("unexpected account info: %+v", account)
	}
}

func TestQuery_SendControlRequest_FailFastWhenTransportCloses(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	errCh := make(chan error, 1)
	go func() {
		_, err := query.sendControlRequest(map[string]any{
			"subtype": "interrupt",
		}, 5*time.Second)
		errCh <- err
	}()

	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for control request write")
	}

	_ = transport.Close()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error when transport closes")
		}
		if !strings.Contains(err.Error(), "transport message stream closed") {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("control request did not fail fast")
	}
}

func TestQuery_TransportErrorDoesNotAbortRouting(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.SendError(fmt.Errorf("transient transport parse error"))

	select {
	case err := <-query.Errors():
		if err == nil || !strings.Contains(err.Error(), "transient transport parse error") {
			t.Fatalf("unexpected routed error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("expected transport error to be surfaced")
	}

	go func() {
		if !transport.WaitForWrite(time.Second) {
			t.Error("timeout waiting for control request write")
			return
		}
		written := transport.Written()
		if len(written) == 0 {
			t.Error("expected control request write")
			return
		}
		var req map[string]any
		if err := json.Unmarshal([]byte(written[len(written)-1]), &req); err != nil {
			t.Errorf("failed to parse request: %v", err)
			return
		}
		reqID, _ := req["request_id"].(string)
		transport.SendMessage(map[string]any{
			"type": "control_response",
			"response": map[string]any{
				"subtype":    "success",
				"request_id": reqID,
				"response":   map[string]any{"ok": true},
			},
		})
	}()

	resp, err := query.sendControlRequest(map[string]any{
		"subtype": "interrupt",
	}, 2*time.Second)
	if err != nil {
		t.Fatalf("expected control request success after transport error, got: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("expected success response, got: %v", resp)
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

	// Wait for response to be written
	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for hook callback response")
	}

	if !callbackCalled.Load() {
		t.Error("hook callback was not called")
	}

	// Verify response was sent
	written := transport.Written()
	if len(written) == 0 {
		t.Error("no response was written")
	}
}

func TestQuery_HandleHookCallback_TypedSetupInputFromMap(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	callbackInvoked := false
	query.hookMu.Lock()
	query.hookCallbacks["hook_setup"] = types.ToGenericCallback(func(input *types.SetupHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
		callbackInvoked = true
		if input.Trigger != "init" {
			t.Fatalf("expected trigger=init, got %s", input.Trigger)
		}
		return types.NewSetupOutput("ok"), nil
	})
	query.hookMu.Unlock()

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_hook_setup",
		"request": map[string]any{
			"subtype":     "hook_callback",
			"callback_id": "hook_setup",
			"input": map[string]any{
				"session_id":      "sess-1",
				"transcript_path": "/tmp/transcript.jsonl",
				"cwd":             "/tmp",
				"hook_event_name": "Setup",
				"trigger":         "init",
			},
		},
	})

	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for typed setup hook response")
	}
	if !callbackInvoked {
		t.Fatal("expected typed setup callback to be invoked")
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

	// Wait for error response to be written
	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for hook error response")
	}

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

	// Wait for response to be written
	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for permission deny response")
	}

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

func TestQuery_HandleCanUseTool_ContextAndToolUseID(t *testing.T) {
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	var capturedCtx *types.ToolPermissionContext
	query.SetCanUseTool(func(toolName string, input map[string]any, ctx *types.ToolPermissionContext) (types.PermissionResult, error) {
		capturedCtx = ctx
		return &types.PermissionResultAllow{
			Behavior: "allow",
		}, nil
	})

	ctx := context.Background()
	if err := query.Start(ctx); err != nil {
		t.Fatal(err)
	}
	defer query.Close()

	transport.SendMessage(map[string]any{
		"type":       "control_request",
		"request_id": "req_perm_ctx",
		"request": map[string]any{
			"subtype":         "can_use_tool",
			"tool_name":       "Bash",
			"input":           map[string]any{"command": "ls"},
			"decision_reason": "outside_workspace",
			"tool_use_id":     "tool_abc",
			"agent_id":        "agent_xyz",
			"description":     "run bash",
		},
	})

	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for permission response")
	}

	if capturedCtx == nil {
		t.Fatal("expected tool permission context to be captured")
	}
	if capturedCtx.ToolUseID != "tool_abc" {
		t.Fatalf("expected ToolUseID=tool_abc, got %s", capturedCtx.ToolUseID)
	}
	if capturedCtx.DecisionReason == nil || *capturedCtx.DecisionReason != "outside_workspace" {
		t.Fatalf("expected decision reason in context, got %+v", capturedCtx.DecisionReason)
	}
	if capturedCtx.AgentID == nil || *capturedCtx.AgentID != "agent_xyz" {
		t.Fatalf("expected agent ID in context, got %+v", capturedCtx.AgentID)
	}
	if capturedCtx.Description == nil || *capturedCtx.Description != "run bash" {
		t.Fatalf("expected description in context, got %+v", capturedCtx.Description)
	}

	written := transport.Written()
	if len(written) == 0 {
		t.Fatal("expected permission response")
	}
	var resp map[string]any
	json.Unmarshal([]byte(written[0]), &resp)
	response := resp["response"].(map[string]any)
	respData := response["response"].(map[string]any)
	if respData["toolUseID"] != "tool_abc" {
		t.Fatalf("expected toolUseID passthrough, got %v", respData["toolUseID"])
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
		t.Error("expected context canceled error")
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

	// Wait for response to be written
	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for MCP tool call response")
	}

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

	// Wait for error response to be written
	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for server not found error response")
	}

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

	// Wait for error response to be written
	if !transport.WaitForWrite(time.Second) {
		t.Fatal("timeout waiting for tool not found error response")
	}

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

// TestQuery_ChannelBufferConstants verifies that channel buffer size constants exist,
// are properly used in NewQuery, and have the expected values.
func TestQuery_ChannelBufferConstants(t *testing.T) {
	// Verify constants are defined with expected values
	if MessageChannelBuffer != 100 {
		t.Errorf("MessageChannelBuffer = %d, want 100", MessageChannelBuffer)
	}
	if RawMessageChannelBuffer != 100 {
		t.Errorf("RawMessageChannelBuffer = %d, want 100", RawMessageChannelBuffer)
	}

	// Verify constants are actually used by creating a Query
	// and checking the channel capacities
	transport := NewMockTransport()
	query := NewQuery(transport, true)

	// Verify the channels were created with correct buffer sizes
	if cap(query.messages) != MessageChannelBuffer {
		t.Errorf("messages channel capacity = %d, want %d", cap(query.messages), MessageChannelBuffer)
	}
	if cap(query.rawMessages) != RawMessageChannelBuffer {
		t.Errorf("rawMessages channel capacity = %d, want %d", cap(query.rawMessages), RawMessageChannelBuffer)
	}
}
