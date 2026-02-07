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
	decisionReason := "path_outside_workspace"
	agentID := "agent-1"
	description := "Request to run bash"
	req := &SDKControlPermissionRequest{
		Subtype:               "can_use_tool",
		ToolName:              "Bash",
		Input:                 map[string]any{"command": "ls"},
		PermissionSuggestions: []PermissionUpdate{{Type: PermissionAddRules}},
		BlockedPath:           stringPtr("/some/path"),
		DecisionReason:        &decisionReason,
		ToolUseID:             "tool-123",
		AgentID:               &agentID,
		Description:           &description,
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
	if result["tool_use_id"] != "tool-123" {
		t.Errorf("Expected tool_use_id 'tool-123', got %v", result["tool_use_id"])
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
		Agents: map[string]map[string]any{
			"researcher": {
				"description": "Research assistant",
				"prompt":      "Gather evidence",
			},
		},
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

func TestSDKControlSetModelRequest(t *testing.T) {
	req := &SDKControlSetModelRequest{
		Subtype: "set_model",
		Model:   "claude-sonnet-4-5",
	}

	var _ SDKControlRequest = req

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"subtype":"set_model","model":"claude-sonnet-4-5"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestSDKControlSetMaxThinkingTokensRequest(t *testing.T) {
	maxTokens := 8192
	req := &SDKControlSetMaxThinkingTokensRequest{
		Subtype:           "set_max_thinking_tokens",
		MaxThinkingTokens: &maxTokens,
	}

	var _ SDKControlRequest = req

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"subtype":"set_max_thinking_tokens","max_thinking_tokens":8192}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
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

func TestSDKControlRewindFilesRequest_WithDryRun(t *testing.T) {
	dryRun := true
	req := &SDKControlRewindFilesRequest{
		Subtype:       "rewind_files",
		UserMessageID: "msg_789",
		DryRun:        &dryRun,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"subtype":"rewind_files","user_message_id":"msg_789","dry_run":true}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
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

func TestSDKControlMcpStatusRequest(t *testing.T) {
	req := &SDKControlMcpStatusRequest{Subtype: "mcp_status"}

	var _ SDKControlRequest = req

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"subtype":"mcp_status"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestSDKControlMcpSetServersRequest(t *testing.T) {
	req := &SDKControlMcpSetServersRequest{
		Subtype: "mcp_set_servers",
		Servers: map[string]any{
			"calc": map[string]any{"type": "stdio", "command": "node"},
		},
	}

	var _ SDKControlRequest = req

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed["subtype"] != "mcp_set_servers" {
		t.Errorf("expected subtype mcp_set_servers, got %v", parsed["subtype"])
	}
	if _, ok := parsed["servers"]; !ok {
		t.Error("expected servers payload")
	}
}

func TestSDKControlMcpReconnectRequest(t *testing.T) {
	req := &SDKControlMcpReconnectRequest{
		Subtype:    "mcp_reconnect",
		ServerName: "calc",
	}

	var _ SDKControlRequest = req

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"subtype":"mcp_reconnect","serverName":"calc"}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestSDKControlMcpToggleRequest(t *testing.T) {
	req := &SDKControlMcpToggleRequest{
		Subtype:    "mcp_toggle",
		ServerName: "calc",
		Enabled:    true,
	}

	var _ SDKControlRequest = req

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	expected := `{"subtype":"mcp_toggle","serverName":"calc","enabled":true}`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
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
			json: `{"subtype":"can_use_tool","tool_name":"Bash","input":{"command":"ls"},"permission_suggestions":null,"blocked_path":null,"decision_reason":"outside_workspace","tool_use_id":"tool-1","agent_id":"agent-1","description":"bash request"}`,
			expected: &SDKControlPermissionRequest{
				Subtype:  "can_use_tool",
				ToolName: "Bash",
				Input:    map[string]any{"command": "ls"},
			},
		},
		{
			name: "initialize",
			json: `{"subtype":"initialize","hooks":null,"agents":{"researcher":{"description":"Research assistant","prompt":"Gather evidence"}}}`,
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
		{
			name: "set_model",
			json: `{"subtype":"set_model","model":"claude-opus-4"}`,
			expected: &SDKControlSetModelRequest{
				Subtype: "set_model",
				Model:   "claude-opus-4",
			},
		},
		{
			name: "set_max_thinking_tokens",
			json: `{"subtype":"set_max_thinking_tokens","max_thinking_tokens":4096}`,
			expected: &SDKControlSetMaxThinkingTokensRequest{
				Subtype: "set_max_thinking_tokens",
			},
		},
		{
			name: "mcp_status",
			json: `{"subtype":"mcp_status"}`,
			expected: &SDKControlMcpStatusRequest{
				Subtype: "mcp_status",
			},
		},
		{
			name: "mcp_set_servers",
			json: `{"subtype":"mcp_set_servers","servers":{"calc":{"type":"stdio","command":"node"}}}`,
			expected: &SDKControlMcpSetServersRequest{
				Subtype: "mcp_set_servers",
			},
		},
		{
			name: "mcp_reconnect",
			json: `{"subtype":"mcp_reconnect","serverName":"calc"}`,
			expected: &SDKControlMcpReconnectRequest{
				Subtype: "mcp_reconnect",
			},
		},
		{
			name: "mcp_toggle",
			json: `{"subtype":"mcp_toggle","serverName":"calc","enabled":true}`,
			expected: &SDKControlMcpToggleRequest{
				Subtype: "mcp_toggle",
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

func TestSDKControlInitializeResponse_Unmarshal(t *testing.T) {
	raw := `{
		"commands":[{"name":"review","description":"Review code","argumentHint":"<path>"}],
		"output_style":"default",
		"available_output_styles":["default","concise"],
		"models":[{"value":"claude-sonnet-4-5","displayName":"Sonnet 4.5","description":"Balanced model"}],
		"account":{"email":"user@example.com","organization":"acme","subscriptionType":"pro"}
	}`

	var resp SDKControlInitializeResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatalf("failed to unmarshal initialize response: %v", err)
	}

	if len(resp.Commands) != 1 || resp.Commands[0].Name != "review" {
		t.Fatalf("unexpected commands payload: %+v", resp.Commands)
	}
	if resp.OutputStyle != "default" {
		t.Fatalf("unexpected output style: %s", resp.OutputStyle)
	}
	if len(resp.Models) != 1 || resp.Models[0].Value != "claude-sonnet-4-5" {
		t.Fatalf("unexpected models payload: %+v", resp.Models)
	}
	if resp.Account.Email != "user@example.com" {
		t.Fatalf("unexpected account payload: %+v", resp.Account)
	}
}

// Helper function for creating string pointers.
func stringPtr(s string) *string {
	return &s
}
