// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package mcp

import (
	"encoding/json"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestMCPHandler_Initialize(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	resp := handler.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(*MCPInitializeResult)
	if !ok {
		t.Fatalf("Expected *MCPInitializeResult, got %T", resp.Result)
	}

	if result.ProtocolVersion != MCPProtocolVersion {
		t.Errorf("Protocol version mismatch: got %s", result.ProtocolVersion)
	}

	if result.ServerInfo.Name != "test" {
		t.Errorf("Server name mismatch: got %s", result.ServerInfo.Name)
	}
}

func TestMCPHandler_ToolsList(t *testing.T) {
	server := types.NewMCPServerBuilder("test").
		WithTool("greet", "Greet someone", nil, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{Content: []types.MCPContent{{Type: "text", Text: "hello"}}}, nil
		}).
		WithTool("calculate", "Do math", nil, func(args map[string]any) (*types.MCPToolResult, error) {
			return &types.MCPToolResult{Content: []types.MCPContent{{Type: "text", Text: "4"}}}, nil
		}).
		Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	resp := handler.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(*MCPToolsListResult)
	if !ok {
		t.Fatalf("Expected *MCPToolsListResult, got %T", resp.Result)
	}

	if len(result.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(result.Tools))
	}
}

func TestMCPHandler_ToolsCall(t *testing.T) {
	server := types.NewMCPServerBuilder("test").
		WithTool("greet", "Greet someone", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}, func(args map[string]any) (*types.MCPToolResult, error) {
			name, _ := args["name"].(string) //nolint:errcheck // Test code, type is guaranteed
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: "Hello, " + name + "!"}},
			}, nil
		}).
		Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "greet",
			"arguments": map[string]any{
				"name": "World",
			},
		},
	}

	resp := handler.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	result, ok := resp.Result.(*MCPToolCallResult)
	if !ok {
		t.Fatalf("Expected *MCPToolCallResult, got %T", resp.Result)
	}

	if len(result.Content) == 0 {
		t.Fatal("Expected content in result")
	}

	// Check content contains greeting
	found := false
	for _, c := range result.Content {
		if c.Type == "text" && c.Text == "Hello, World!" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected greeting in result, got: %v", result.Content)
	}
}

func TestMCPHandler_ToolsCall_NotFound(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name": "nonexistent",
		},
	}

	resp := handler.HandleRequest(req)
	if resp.Error == nil {
		t.Fatal("Expected error for nonexistent tool")
	}

	if resp.Error.Code != MCPErrorMethodNotFound {
		t.Errorf("Expected method not found error, got code %d", resp.Error.Code)
	}
}

func TestMCPHandler_UnknownMethod(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown/method",
	}

	resp := handler.HandleRequest(req)
	if resp.Error == nil {
		t.Fatal("Expected error for unknown method")
	}

	if resp.Error.Code != MCPErrorMethodNotFound {
		t.Errorf("Expected method not found error, got code %d", resp.Error.Code)
	}
}

func TestMCPHandler_HandleBytes(t *testing.T) {
	server := types.NewMCPServerBuilder("test").
		WithTool("echo", "Echo input", nil, func(args map[string]any) (*types.MCPToolResult, error) {
			text, _ := args["text"].(string)
			return &types.MCPToolResult{
				Content: []types.MCPContent{{Type: "text", Text: text}},
			}, nil
		}).
		Build()
	handler := NewMCPHandler(server)

	reqBytes, _ := json.Marshal(&MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      "echo",
			"arguments": map[string]any{"text": "hello"},
		},
	})

	respBytes, err := handler.HandleBytes(reqBytes)
	if err != nil {
		t.Fatalf("HandleBytes failed: %v", err)
	}

	var resp MCPResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
}

func TestMCPHandler_Ping(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	req := &MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "ping",
	}

	resp := handler.HandleRequest(req)
	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}
	if resp.Result == nil {
		t.Error("Expected non-nil result for ping")
	}
}

func TestMCPHandler_Notification(t *testing.T) {
	server := types.NewMCPServerBuilder("test").Build()
	handler := NewMCPHandler(server)

	// Notifications have no ID
	req := &MCPRequest{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	resp := handler.HandleRequest(req)
	// Notifications should return nil response
	if resp != nil {
		t.Errorf("Expected nil response for notification, got %+v", resp)
	}
}
