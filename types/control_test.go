// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

// TestSDKControlInterruptRequest tests the interrupt request type.
func TestSDKControlInterruptRequest(t *testing.T) {
	req := &SDKControlInterruptRequest{
		Subtype: "interrupt",
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `{"subtype":"interrupt"}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}

	// Test type discrimination
	if req.ControlRequestType() != "interrupt" {
		t.Errorf("Expected type 'interrupt', got '%s'", req.ControlRequestType())
	}
}

// TestSDKControlPermissionRequest tests the permission request type.
func TestSDKControlPermissionRequest(t *testing.T) {
	req := &SDKControlPermissionRequest{
		Subtype:               "can_use_tool",
		ToolName:              "Bash",
		Input:                 map[string]any{"command": "ls"},
		PermissionSuggestions: []PermissionUpdate{{Type: PermissionAddRules}},
		BlockedPath:           stringPtr("/some/path"),
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["subtype"] != "can_use_tool" {
		t.Errorf("Expected subtype 'can_use_tool', got %v", result["subtype"])
	}
	if result["tool_name"] != "Bash" {
		t.Errorf("Expected tool_name 'Bash', got %v", result["tool_name"])
	}

	// Test type discrimination
	if req.ControlRequestType() != "can_use_tool" {
		t.Errorf("Expected type 'can_use_tool', got '%s'", req.ControlRequestType())
	}
}

// TestSDKControlInitializeRequest tests the initialize request type.
func TestSDKControlInitializeRequest(t *testing.T) {
	hooks := map[HookEvent]any{
		HookPreToolUse: "callback_id_123",
	}
	req := &SDKControlInitializeRequest{
		Subtype: "initialize",
		Hooks:   hooks,
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["subtype"] != "initialize" {
		t.Errorf("Expected subtype 'initialize', got %v", result["subtype"])
	}

	// Test type discrimination
	if req.ControlRequestType() != "initialize" {
		t.Errorf("Expected type 'initialize', got '%s'", req.ControlRequestType())
	}
}

// TestSDKControlSetPermissionModeRequest tests the set permission mode request type.
func TestSDKControlSetPermissionModeRequest(t *testing.T) {
	req := &SDKControlSetPermissionModeRequest{
		Subtype: "set_permission_mode",
		Mode:    "enabled",
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `{"subtype":"set_permission_mode","mode":"enabled"}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}

	// Test type discrimination
	if req.ControlRequestType() != "set_permission_mode" {
		t.Errorf("Expected type 'set_permission_mode', got '%s'", req.ControlRequestType())
	}
}

// TestSDKHookCallbackRequest tests the hook callback request type.
func TestSDKHookCallbackRequest(t *testing.T) {
	req := &SDKHookCallbackRequest{
		Subtype:    "hook_callback",
		CallbackID: "callback_123",
		Input:      map[string]any{"test": "data"},
		ToolUseID:  stringPtr("tool_456"),
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["subtype"] != "hook_callback" {
		t.Errorf("Expected subtype 'hook_callback', got %v", result["subtype"])
	}
	if result["callback_id"] != "callback_123" {
		t.Errorf("Expected callback_id 'callback_123', got %v", result["callback_id"])
	}

	// Test type discrimination
	if req.ControlRequestType() != "hook_callback" {
		t.Errorf("Expected type 'hook_callback', got '%s'", req.ControlRequestType())
	}
}

// TestSDKControlMcpMessageRequest tests the MCP message request type.
func TestSDKControlMcpMessageRequest(t *testing.T) {
	req := &SDKControlMcpMessageRequest{
		Subtype:    "mcp_message",
		ServerName: "test-server",
		Message:    map[string]any{"method": "tools/list"},
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["subtype"] != "mcp_message" {
		t.Errorf("Expected subtype 'mcp_message', got %v", result["subtype"])
	}
	if result["server_name"] != "test-server" {
		t.Errorf("Expected server_name 'test-server', got %v", result["server_name"])
	}

	// Test type discrimination
	if req.ControlRequestType() != "mcp_message" {
		t.Errorf("Expected type 'mcp_message', got '%s'", req.ControlRequestType())
	}
}

// TestSDKControlRewindFilesRequest tests the rewind files request type.
func TestSDKControlRewindFilesRequest(t *testing.T) {
	req := &SDKControlRewindFilesRequest{
		Subtype:       "rewind_files",
		UserMessageID: "msg_789",
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	expected := `{"subtype":"rewind_files","user_message_id":"msg_789"}`
	if string(data) != expected {
		t.Errorf("Expected %s, got %s", expected, string(data))
	}

	// Test type discrimination
	if req.ControlRequestType() != "rewind_files" {
		t.Errorf("Expected type 'rewind_files', got '%s'", req.ControlRequestType())
	}
}

// TestSDKControlMcpToolCallRequest tests the MCP tool call request type.
func TestSDKControlMcpToolCallRequest(t *testing.T) {
	req := &SDKControlMcpToolCallRequest{
		Subtype:    "mcp_tool_call",
		ServerName: "test-server",
		ToolName:   "calculate",
		Input:      map[string]any{"a": 1, "b": 2},
	}

	// Test that it implements SDKControlRequest interface
	var _ SDKControlRequest = req

	// Test JSON marshaling
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal to verify structure
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result["subtype"] != "mcp_tool_call" {
		t.Errorf("Expected subtype 'mcp_tool_call', got %v", result["subtype"])
	}
	if result["server_name"] != "test-server" {
		t.Errorf("Expected server_name 'test-server', got %v", result["server_name"])
	}
	if result["tool_name"] != "calculate" {
		t.Errorf("Expected tool_name 'calculate', got %v", result["tool_name"])
	}

	// Test type discrimination
	if req.ControlRequestType() != "mcp_tool_call" {
		t.Errorf("Expected type 'mcp_tool_call', got '%s'", req.ControlRequestType())
	}
}

// TestParseSDKControlRequest tests parsing raw JSON into typed structs.
func TestParseSDKControlRequest(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected SDKControlRequest
	}{
		{
			name:     "interrupt",
			json:     `{"subtype":"interrupt"}`,
			expected: &SDKControlInterruptRequest{Subtype: "interrupt"},
		},
		{
			name: "permission",
			json: `{"subtype":"can_use_tool","tool_name":"Bash","input":{"command":"ls"},"permission_suggestions":null,"blocked_path":null}`,
			expected: &SDKControlPermissionRequest{
				Subtype:  "can_use_tool",
				ToolName: "Bash",
				Input:    map[string]any{"command": "ls"},
			},
		},
		{
			name: "initialize",
			json: `{"subtype":"initialize","hooks":null}`,
			expected: &SDKControlInitializeRequest{
				Subtype: "initialize",
			},
		},
		{
			name: "set_permission_mode",
			json: `{"subtype":"set_permission_mode","mode":"enabled"}`,
			expected: &SDKControlSetPermissionModeRequest{
				Subtype: "set_permission_mode",
				Mode:    "enabled",
			},
		},
		{
			name: "hook_callback",
			json: `{"subtype":"hook_callback","callback_id":"cb_123","input":null,"tool_use_id":null}`,
			expected: &SDKHookCallbackRequest{
				Subtype:    "hook_callback",
				CallbackID: "cb_123",
			},
		},
		{
			name: "mcp_message",
			json: `{"subtype":"mcp_message","server_name":"test","message":null}`,
			expected: &SDKControlMcpMessageRequest{
				Subtype:    "mcp_message",
				ServerName: "test",
			},
		},
		{
			name: "mcp_tool_call",
			json: `{"subtype":"mcp_tool_call","server_name":"test","tool_name":"calculate","input":null}`,
			expected: &SDKControlMcpToolCallRequest{
				Subtype:    "mcp_tool_call",
				ServerName: "test",
				ToolName:   "calculate",
			},
		},
		{
			name: "rewind_files",
			json: `{"subtype":"rewind_files","user_message_id":"msg_123"}`,
			expected: &SDKControlRewindFilesRequest{
				Subtype:       "rewind_files",
				UserMessageID: "msg_123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var raw map[string]any
			if err := json.Unmarshal([]byte(tt.json), &raw); err != nil {
				t.Fatalf("Failed to unmarshal test JSON: %v", err)
			}

			result, err := ParseSDKControlRequest(raw)
			if err != nil {
				t.Fatalf("ParseSDKControlRequest failed: %v", err)
			}

			if result.ControlRequestType() != tt.expected.ControlRequestType() {
				t.Errorf("Expected type %s, got %s",
					tt.expected.ControlRequestType(),
					result.ControlRequestType())
			}
		})
	}
}

// TestParseSDKControlRequest_InvalidSubtype tests error handling for unknown subtypes.
func TestParseSDKControlRequest_InvalidSubtype(t *testing.T) {
	raw := map[string]any{
		"subtype": "unknown_type",
	}

	_, err := ParseSDKControlRequest(raw)
	if err == nil {
		t.Error("Expected error for unknown subtype, got nil")
	}
}

// Helper function for creating string pointers.
func stringPtr(s string) *string {
	return &s
}
