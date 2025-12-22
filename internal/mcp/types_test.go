// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package mcp

import (
	"encoding/json"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestMCPToolDefinition_JSON(t *testing.T) {
	tool := MCPToolDefinition{
		Name:        "calculate",
		Description: "Perform calculations",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "Math expression",
				},
			},
			"required": []string{"expression"},
		},
	}

	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded MCPToolDefinition
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != tool.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, tool.Name)
	}
	if decoded.Description != tool.Description {
		t.Errorf("Description mismatch: got %s, want %s", decoded.Description, tool.Description)
	}
}

func TestMCPRequest_JSON(t *testing.T) {
	tests := []struct {
		name    string
		request MCPRequest
	}{
		{
			name: "initialize",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
				Params: map[string]any{
					"protocolVersion": MCPProtocolVersion,
					"capabilities":    map[string]any{},
					"clientInfo": map[string]any{
						"name":    "claude-agent-sdk-go",
						"version": types.Version,
					},
				},
			},
		},
		{
			name: "tools/list",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      2,
				Method:  "tools/list",
			},
		},
		{
			name: "tools/call",
			request: MCPRequest{
				JSONRPC: "2.0",
				ID:      3,
				Method:  "tools/call",
				Params: map[string]any{
					"name": "calculate",
					"arguments": map[string]any{
						"expression": "2+2",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var decoded MCPRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if decoded.Method != tt.request.Method {
				t.Errorf("Method mismatch: got %s, want %s", decoded.Method, tt.request.Method)
			}
			if decoded.JSONRPC != "2.0" {
				t.Errorf("JSONRPC mismatch: got %s, want 2.0", decoded.JSONRPC)
			}
		})
	}
}

func TestMCPResponse_JSON(t *testing.T) {
	tests := []struct {
		name     string
		response MCPResponse
	}{
		{
			name: "success",
			response: MCPResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result: map[string]any{
					"content": []any{
						map[string]any{
							"type": "text",
							"text": "Result: 4",
						},
					},
				},
			},
		},
		{
			name: "error",
			response: MCPResponse{
				JSONRPC: "2.0",
				ID:      2,
				Error: &MCPError{
					Code:    MCPErrorInvalidRequest,
					Message: "Invalid Request",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var decoded MCPResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if tt.response.Error != nil && decoded.Error == nil {
				t.Error("Expected error but got nil")
			}
			if tt.response.Result != nil && decoded.Result == nil {
				t.Error("Expected result but got nil")
			}
		})
	}
}

func TestMCPError_JSON(t *testing.T) {
	err := &MCPError{
		Code:    MCPErrorMethodNotFound,
		Message: "Method not found",
		Data:    "unknown/method",
	}

	data, err2 := json.Marshal(err)
	if err2 != nil {
		t.Fatalf("Marshal failed: %v", err2)
	}

	var decoded MCPError
	if err2 := json.Unmarshal(data, &decoded); err2 != nil {
		t.Fatalf("Unmarshal failed: %v", err2)
	}

	if decoded.Code != MCPErrorMethodNotFound {
		t.Errorf("Code mismatch: got %d, want %d", decoded.Code, MCPErrorMethodNotFound)
	}
	if decoded.Message != "Method not found" {
		t.Errorf("Message mismatch: got %s", decoded.Message)
	}
}

func TestMCPNotification_JSON(t *testing.T) {
	notification := MCPNotification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]any{},
	}

	data, err := json.Marshal(notification)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded MCPNotification
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Method != notification.Method {
		t.Errorf("Method mismatch: got %s, want %s", decoded.Method, notification.Method)
	}
}

func TestNewMCPResponse(t *testing.T) {
	resp := NewMCPResponse(1, map[string]any{"key": "value"})

	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC should be 2.0, got %s", resp.JSONRPC)
	}
	if resp.ID != 1 {
		t.Errorf("ID should be 1, got %v", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("Error should be nil")
	}
}

func TestNewMCPErrorResponse(t *testing.T) {
	err := NewMCPError(MCPErrorInternal, "Internal error", nil)
	resp := NewMCPErrorResponse(2, err)

	if resp.JSONRPC != "2.0" {
		t.Errorf("JSONRPC should be 2.0, got %s", resp.JSONRPC)
	}
	if resp.ID != 2 {
		t.Errorf("ID should be 2, got %v", resp.ID)
	}
	if resp.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if resp.Error.Code != MCPErrorInternal {
		t.Errorf("Error code should be %d, got %d", MCPErrorInternal, resp.Error.Code)
	}
}

func TestMCPInitializeResult_JSON(t *testing.T) {
	result := MCPInitializeResult{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: &MCPCapabilities{
			Tools: &MCPToolsCapability{
				ListChanged: true,
			},
		},
		ServerInfo: MCPServerInfo{
			Name:    "test-server",
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded MCPInitializeResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.ProtocolVersion != MCPProtocolVersion {
		t.Errorf("ProtocolVersion mismatch: got %s", decoded.ProtocolVersion)
	}
	if decoded.ServerInfo.Name != "test-server" {
		t.Errorf("ServerInfo.Name mismatch: got %s", decoded.ServerInfo.Name)
	}
}

func TestMCPToolsListResult_JSON(t *testing.T) {
	result := MCPToolsListResult{
		Tools: []MCPToolDefinition{
			{
				Name:        "tool1",
				Description: "First tool",
				InputSchema: map[string]any{"type": "object"},
			},
			{
				Name:        "tool2",
				Description: "Second tool",
				InputSchema: map[string]any{"type": "object"},
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded MCPToolsListResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(decoded.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(decoded.Tools))
	}
}
