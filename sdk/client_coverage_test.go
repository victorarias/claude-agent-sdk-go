// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestClient_Options(t *testing.T) {
	opts := types.WithModel("test-model")
	client := NewClient(opts)

	if client.Options() == nil {
		t.Fatal("Options() returned nil")
	}

	if client.Options().Model != "test-model" {
		t.Errorf("expected model 'test-model', got %q", client.Options().Model)
	}
}

func TestClient_Disconnect(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

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

	if !client.IsConnected() {
		t.Error("client should be connected")
	}

	// Test Disconnect
	err := client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect failed: %v", err)
	}

	if client.IsConnected() {
		t.Error("client should be disconnected")
	}

	// Disconnect when already disconnected should not error
	err = client.Disconnect()
	if err != nil {
		t.Errorf("Disconnect on disconnected client failed: %v", err)
	}
}

func TestWithUserPromptSubmitHook(t *testing.T) {
	called := false
	client := NewClient(
		WithUserPromptSubmitHook(
			func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				called = true
				return &types.HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[types.HookUserPromptSubmit]) != 1 {
		t.Errorf("expected 1 user prompt submit hook, got %d", len(client.hooks[types.HookUserPromptSubmit]))
	}
	_ = called
}

func TestWithSubagentStopHook(t *testing.T) {
	called := false
	client := NewClient(
		WithSubagentStopHook(
			func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				called = true
				return &types.HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[types.HookSubagentStop]) != 1 {
		t.Errorf("expected 1 subagent stop hook, got %d", len(client.hooks[types.HookSubagentStop]))
	}
	_ = called
}

func TestWithPreCompactHook(t *testing.T) {
	called := false
	client := NewClient(
		WithPreCompactHook(
			func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				called = true
				return &types.HookOutput{}, nil
			},
		),
	)

	if len(client.hooks[types.HookPreCompact]) != 1 {
		t.Errorf("expected 1 pre-compact hook, got %d", len(client.hooks[types.HookPreCompact]))
	}
	_ = called
}

func TestClient_ReceiveResponse(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	// Respond to initialize and query
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

			// Check if it's a control request (Initialize)
			if reqType, ok := req["type"].(string); ok && reqType == "control_request" {
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
			} else {
				// It's a query message
				time.Sleep(10 * time.Millisecond)
				transport.SendMessage(map[string]any{
					"type": "assistant",
					"message": map[string]any{
						"content": []any{
							map[string]any{"type": "text", "text": "Response"},
						},
					},
				})
				transport.SendMessage(map[string]any{
					"type":    "result",
					"subtype": "success",
				})
				return
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	messages, err := client.ReceiveResponse("Hello")
	if err != nil {
		t.Errorf("ReceiveResponse failed: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestClient_Interrupt(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	// Respond to initialize and interrupt
	go func() {
		defer close(done)
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

			// Respond to all control requests
			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{"session_id": "test_session"},
				},
			})

			// If it's the interrupt request, we're done
			if requestData, ok := req["request"].(map[string]any); ok {
				if requestData["subtype"] == "interrupt" {
					return
				}
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Test Interrupt
	err := client.Interrupt()
	if err != nil {
		t.Errorf("Interrupt failed: %v", err)
	}

	<-done
}

func TestClient_Interrupt_NotConnected(t *testing.T) {
	client := NewClient()

	// Test Interrupt when not connected
	err := client.Interrupt()
	if err == nil {
		t.Error("expected error when interrupting disconnected client")
	}

	if _, ok := err.(*types.ConnectionError); !ok {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestClient_SetPermissionMode(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	// Respond to initialize and SetPermissionMode
	go func() {
		defer close(done)
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

			// If it's the set_permission_mode request, we're done
			if requestData, ok := req["request"].(map[string]any); ok {
				if requestData["subtype"] == "set_permission_mode" {
					return
				}
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err := client.SetPermissionMode(types.PermissionBypass)
	if err != nil {
		t.Errorf("SetPermissionMode failed: %v", err)
	}

	<-done
}

func TestClient_SetPermissionMode_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.SetPermissionMode(types.PermissionBypass)
	if err == nil {
		t.Error("expected error when setting permission mode on disconnected client")
	}

	if _, ok := err.(*types.ConnectionError); !ok {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestClient_SetModel(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	// Respond to initialize and SetModel
	go func() {
		defer close(done)
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

			// If it's the set_model request, we're done
			if requestData, ok := req["request"].(map[string]any); ok {
				if requestData["subtype"] == "set_model" {
					return
				}
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err := client.SetModel("claude-opus-4")
	if err != nil {
		t.Errorf("SetModel failed: %v", err)
	}

	<-done
}

func TestClient_SetModel_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.SetModel("claude-opus-4")
	if err == nil {
		t.Error("expected error when setting model on disconnected client")
	}

	if _, ok := err.(*types.ConnectionError); !ok {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestClient_ClearModel(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan error, 1)
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

			requestData, ok := req["request"].(map[string]any)
			if !ok {
				continue
			}
			if requestData["subtype"] == "set_model" {
				if _, ok := requestData["model"]; ok {
					done <- fmt.Errorf("expected set_model without model field, got %v", requestData["model"])
					return
				}
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   map[string]any{},
				},
			})

			if requestData["subtype"] == "set_model" {
				done <- nil
				return
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.ClearModel(); err != nil {
		t.Fatalf("ClearModel failed: %v", err)
	}

	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestClient_ClearModel_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.ClearModel()
	if err == nil {
		t.Error("expected error when clearing model on disconnected client")
	}

	if _, ok := err.(*types.ConnectionError); !ok {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestClient_SetMaxThinkingTokens(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	go func() {
		defer close(done)
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

			if requestData, ok := req["request"].(map[string]any); ok {
				if requestData["subtype"] == "set_max_thinking_tokens" {
					return
				}
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	maxTokens := 2048
	if err := client.SetMaxThinkingTokens(&maxTokens); err != nil {
		t.Errorf("SetMaxThinkingTokens failed: %v", err)
	}

	<-done
}

func TestClient_RewindFiles(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	// Respond to initialize and RewindFiles
	go func() {
		defer close(done)
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

			// If it's the rewind_files request, we're done
			if requestData, ok := req["request"].(map[string]any); ok {
				if requestData["subtype"] == "rewind_files" {
					return
				}
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	err := client.RewindFiles("msg_123")
	if err != nil {
		t.Errorf("RewindFiles failed: %v", err)
	}

	<-done
}

func TestClient_RewindFiles_NotConnected(t *testing.T) {
	client := NewClient()

	err := client.RewindFiles("msg_123")
	if err == nil {
		t.Error("expected error when rewinding files on disconnected client")
	}

	if _, ok := err.(*types.ConnectionError); !ok {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestClient_GetMCPStatus(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	go func() {
		defer close(done)
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

			responseData := map[string]any{"session_id": "test_session"}
			if requestData, ok := req["request"].(map[string]any); ok {
				if requestData["subtype"] == "mcp_status" {
					responseData = map[string]any{
						"mcpServers": []any{
							map[string]any{"name": "calc", "status": "connected"},
						},
					}
				}
			}

			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response":   responseData,
				},
			})

			if requestData, ok := req["request"].(map[string]any); ok && requestData["subtype"] == "mcp_status" {
				return
			}
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	status, err := client.GetMCPStatus()
	if err != nil {
		t.Fatalf("GetMCPStatus failed: %v", err)
	}
	if status["mcpServers"] == nil {
		t.Fatalf("expected mcpServers response, got %v", status)
	}
	<-done
}

func TestClient_GetMCPStatus_NotConnected(t *testing.T) {
	client := NewClient()
	_, err := client.GetMCPStatus()
	if err == nil {
		t.Fatal("expected error when fetching MCP status on disconnected client")
	}
	if _, ok := err.(*types.ConnectionError); !ok {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestClient_MCPServerControlMethods(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	go func() {
		defer close(done)
		expected := map[string]bool{
			"mcp_reconnect":   false,
			"mcp_toggle":      false,
			"mcp_set_servers": false,
		}
		allSeen := func() bool {
			for _, seen := range expected {
				if !seen {
					return false
				}
			}
			return true
		}
		for !allSeen() {
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

			response := map[string]any{}
			if requestData, ok := req["request"].(map[string]any); ok {
				if subtype, ok := requestData["subtype"].(string); ok {
					if _, tracked := expected[subtype]; tracked {
						expected[subtype] = true
					}
					if subtype == "mcp_set_servers" {
						response = map[string]any{
							"added":   []any{"calc"},
							"removed": []any{},
							"errors":  map[string]any{},
						}
					}
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

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.ReconnectMCPServer("calc"); err != nil {
		t.Fatalf("ReconnectMCPServer failed: %v", err)
	}
	if err := client.ToggleMCPServer("calc", false); err != nil {
		t.Fatalf("ToggleMCPServer failed: %v", err)
	}
	result, err := client.SetMCPServers(map[string]any{
		"calc": map[string]any{"type": "stdio", "command": "node"},
	})
	if err != nil {
		t.Fatalf("SetMCPServers failed: %v", err)
	}
	if len(result.Added) != 1 || result.Added[0] != "calc" {
		t.Fatalf("unexpected SetMCPServers result: %+v", result)
	}

	<-done
}

func TestClient_SetMCPServers_WithSDKInstancePayload(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	done := make(chan bool)
	go func() {
		defer close(done)
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
			requestData, ok := req["request"].(map[string]any)
			if !ok {
				continue
			}
			if requestData["subtype"] != "mcp_set_servers" {
				// Respond to initialize or other control requests.
				reqID, _ := req["request_id"].(string)
				transport.SendMessage(map[string]any{
					"type": "control_response",
					"response": map[string]any{
						"subtype":    "success",
						"request_id": reqID,
						"response":   map[string]any{},
					},
				})
				continue
			}

			servers, _ := requestData["servers"].(map[string]any)
			local, ok := servers["local"].(map[string]any)
			if !ok {
				t.Errorf("expected local sdk server in payload, got %v", servers["local"])
			} else {
				if local["type"] != "sdk" {
					t.Errorf("expected local server type sdk, got %v", local["type"])
				}
				if local["name"] != "local" {
					t.Errorf("expected local server name field, got %v", local["name"])
				}
			}

			reqID, _ := req["request_id"].(string)
			transport.SendMessage(map[string]any{
				"type": "control_response",
				"response": map[string]any{
					"subtype":    "success",
					"request_id": reqID,
					"response": map[string]any{
						"added":   []any{"local"},
						"removed": []any{},
						"errors":  map[string]any{},
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

	local := types.NewMCPServerBuilder("local").Build()
	if _, err := client.SetMCPServers(map[string]any{
		"local": local,
	}); err != nil {
		t.Fatalf("SetMCPServers failed: %v", err)
	}

	<-done
}

func TestClient_ServerInfo(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

	// Respond to initialize with info
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
						"session_id": "test_session",
						"version":    "2.0.0",
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

	info := client.ServerInfo()
	if info == nil {
		t.Fatal("ServerInfo returned nil")
	}

	if sessionID, ok := info["session_id"].(string); !ok || sessionID != "test_session" {
		t.Errorf("expected session_id 'test_session', got %v", info["session_id"])
	}
}

func TestClient_ServerInfo_NotConnected(t *testing.T) {
	client := NewClient()

	info := client.ServerInfo()
	if info != nil {
		t.Errorf("expected nil from ServerInfo when not connected, got %v", info)
	}
}

func TestClient_InitializationMetadataMethods(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

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
							"email": "user@example.com",
						},
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

	init, err := client.InitializationResult()
	if err != nil {
		t.Fatalf("InitializationResult failed: %v", err)
	}
	if len(init.Commands) != 1 {
		t.Fatalf("unexpected init commands: %+v", init.Commands)
	}

	commands, err := client.SupportedCommands()
	if err != nil || len(commands) != 1 {
		t.Fatalf("SupportedCommands failed: %v, %+v", err, commands)
	}

	models, err := client.SupportedModels()
	if err != nil || len(models) != 1 {
		t.Fatalf("SupportedModels failed: %v, %+v", err, models)
	}

	account, err := client.AccountInfo()
	if err != nil {
		t.Fatalf("AccountInfo failed: %v", err)
	}
	if account.Email != "user@example.com" {
		t.Fatalf("unexpected account info: %+v", account)
	}
}

func TestClient_ResultReceived(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

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

			// Send a result message
			time.Sleep(10 * time.Millisecond)
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

	if err := client.SendQuery("test"); err != nil {
		t.Fatal(err)
	}

	// Wait for result
	time.Sleep(50 * time.Millisecond)

	if !client.ResultReceived() {
		t.Error("expected ResultReceived to be true")
	}
}

func TestClient_ResultReceived_NotConnected(t *testing.T) {
	client := NewClient()

	if client.ResultReceived() {
		t.Error("expected ResultReceived to be false when not connected")
	}
}

func TestClient_LastResult(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

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

			// Send a result message
			time.Sleep(10 * time.Millisecond)
			transport.SendMessage(map[string]any{
				"type":       "result",
				"subtype":    "success",
				"session_id": "test_123",
			})
			return
		}
	}()

	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	if err := client.SendQuery("test"); err != nil {
		t.Fatal(err)
	}

	// Wait for result
	time.Sleep(50 * time.Millisecond)

	result := client.LastResult()
	if result == nil {
		t.Error("expected LastResult to return a result")
	}
}

func TestClient_LastResult_NotConnected(t *testing.T) {
	client := NewClient()

	result := client.LastResult()
	if result != nil {
		t.Errorf("expected nil from LastResult when not connected, got %v", result)
	}
}

func TestClient_RawMessages(t *testing.T) {
	transport := NewMockTransport()
	client := NewClient(types.WithTransport(transport))

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

			// Send raw messages
			time.Sleep(10 * time.Millisecond)
			transport.SendMessage(map[string]any{
				"type": "test",
				"data": "value",
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

	if err := client.SendQuery("test"); err != nil {
		t.Fatal(err)
	}

	var count int
	for msg := range client.RawMessages() {
		count++
		if msg["type"] == "result" {
			break
		}
	}

	if count < 1 {
		t.Error("expected to receive at least one raw message")
	}
}

func TestClient_RawMessages_NotConnected(t *testing.T) {
	client := NewClient()

	ch := client.RawMessages()
	if ch == nil {
		t.Fatal("expected non-nil channel from RawMessages")
	}

	// Channel should be closed immediately
	_, ok := <-ch
	if ok {
		t.Error("expected closed channel from RawMessages when not connected")
	}
}

func TestClient_Errors(t *testing.T) {
	client := NewClient()

	ch := client.Errors()
	if ch == nil {
		t.Fatal("expected non-nil channel from Errors")
	}

	// Channel should be closed immediately when not connected
	_, ok := <-ch
	if ok {
		t.Error("expected closed channel from Errors when not connected")
	}
}
