// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

// TestTextBlock tests TextBlock methods.
func TestTextBlock(t *testing.T) {
	t.Run("BlockType returns text", func(t *testing.T) {
		block := &TextBlock{TextContent: "hello"}
		if block.BlockType() != "text" {
			t.Errorf("Expected BlockType to return 'text', got %q", block.BlockType())
		}
	})

	t.Run("Type returns text", func(t *testing.T) {
		block := &TextBlock{TextContent: "hello"}
		if block.Type() != "text" {
			t.Errorf("Expected Type to return 'text', got %q", block.Type())
		}
	})

	t.Run("Text returns content", func(t *testing.T) {
		block := &TextBlock{TextContent: "hello world"}
		if block.Text() != "hello world" {
			t.Errorf("Expected Text to return 'hello world', got %q", block.Text())
		}
	})
}

// TestThinkingBlock tests ThinkingBlock methods.
func TestThinkingBlock(t *testing.T) {
	t.Run("BlockType returns thinking", func(t *testing.T) {
		block := &ThinkingBlock{ThinkingContent: "reasoning"}
		if block.BlockType() != "thinking" {
			t.Errorf("Expected BlockType to return 'thinking', got %q", block.BlockType())
		}
	})

	t.Run("Type returns thinking", func(t *testing.T) {
		block := &ThinkingBlock{ThinkingContent: "reasoning"}
		if block.Type() != "thinking" {
			t.Errorf("Expected Type to return 'thinking', got %q", block.Type())
		}
	})

	t.Run("Thinking returns content", func(t *testing.T) {
		block := &ThinkingBlock{ThinkingContent: "deep thought"}
		if block.Thinking() != "deep thought" {
			t.Errorf("Expected Thinking to return 'deep thought', got %q", block.Thinking())
		}
	})
}

// TestToolUseBlock tests ToolUseBlock methods.
func TestToolUseBlock(t *testing.T) {
	t.Run("BlockType returns tool_use", func(t *testing.T) {
		block := &ToolUseBlock{ID: "1", Name: "test"}
		if block.BlockType() != "tool_use" {
			t.Errorf("Expected BlockType to return 'tool_use', got %q", block.BlockType())
		}
	})

	t.Run("Type returns tool_use", func(t *testing.T) {
		block := &ToolUseBlock{ID: "1", Name: "test"}
		if block.Type() != "tool_use" {
			t.Errorf("Expected Type to return 'tool_use', got %q", block.Type())
		}
	})

	t.Run("Input returns tool input", func(t *testing.T) {
		input := map[string]any{"param": "value"}
		block := &ToolUseBlock{ID: "1", Name: "test", ToolInput: input}
		result := block.Input()
		if result["param"] != "value" {
			t.Errorf("Expected Input to return map with param=value, got %v", result)
		}
	})
}

// TestToolResultBlock tests ToolResultBlock methods.
func TestToolResultBlock(t *testing.T) {
	t.Run("BlockType returns tool_result", func(t *testing.T) {
		block := &ToolResultBlock{ToolUseID: "1"}
		if block.BlockType() != "tool_result" {
			t.Errorf("Expected BlockType to return 'tool_result', got %q", block.BlockType())
		}
	})

	t.Run("Type returns tool_result", func(t *testing.T) {
		block := &ToolResultBlock{ToolUseID: "1"}
		if block.Type() != "tool_result" {
			t.Errorf("Expected Type to return 'tool_result', got %q", block.Type())
		}
	})

	t.Run("Content returns result content", func(t *testing.T) {
		block := &ToolResultBlock{ToolUseID: "1", ResultContent: "success"}
		if block.Content() != "success" {
			t.Errorf("Expected Content to return 'success', got %q", block.Content())
		}
	})
}

// TestParseContentBlock tests ParseContentBlock function.
func TestParseContentBlock(t *testing.T) {
	tests := []struct {
		name        string
		input       map[string]any
		expectError bool
		expectType  string
	}{
		{
			name: "parse text block",
			input: map[string]any{
				"type": "text",
				"text": "hello",
			},
			expectError: false,
			expectType:  "text",
		},
		{
			name: "parse thinking block",
			input: map[string]any{
				"type":      "thinking",
				"thinking":  "reasoning",
				"signature": "sig123",
			},
			expectError: false,
			expectType:  "thinking",
		},
		{
			name: "parse tool_use block",
			input: map[string]any{
				"type":  "tool_use",
				"id":    "tool1",
				"name":  "calculator",
				"input": map[string]any{"x": 5},
			},
			expectError: false,
			expectType:  "tool_use",
		},
		{
			name: "parse tool_result block",
			input: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool1",
				"content":     "42",
				"is_error":    false,
			},
			expectError: false,
			expectType:  "tool_result",
		},
		{
			name: "missing type field",
			input: map[string]any{
				"text": "hello",
			},
			expectError: true,
		},
		{
			name: "unknown block type",
			input: map[string]any{
				"type": "unknown_type",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := ParseContentBlock(tt.input)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if block == nil {
					t.Fatal("Expected block but got nil")
				}
				if block.Type() != tt.expectType {
					t.Errorf("Expected type %q, got %q", tt.expectType, block.Type())
				}
			}
		})
	}
}

// TestUserMessage tests UserMessage methods.
func TestUserMessage(t *testing.T) {
	t.Run("MessageType returns user", func(t *testing.T) {
		msg := &UserMessage{}
		if msg.MessageType() != "user" {
			t.Errorf("Expected MessageType to return 'user', got %q", msg.MessageType())
		}
	})

	t.Run("Text concatenates text blocks", func(t *testing.T) {
		msg := &UserMessage{
			Content: []ContentBlock{
				&TextBlock{TextContent: "hello "},
				&TextBlock{TextContent: "world"},
				&ToolUseBlock{ID: "1", Name: "test"},
			},
		}
		expected := "hello world"
		if msg.Text() != expected {
			t.Errorf("Expected %q, got %q", expected, msg.Text())
		}
	})

	t.Run("Text returns empty string when no text blocks", func(t *testing.T) {
		msg := &UserMessage{
			Content: []ContentBlock{
				&ToolUseBlock{ID: "1", Name: "test"},
			},
		}
		if msg.Text() != "" {
			t.Errorf("Expected empty string, got %q", msg.Text())
		}
	})

	t.Run("tool_use_result marshals and unmarshals", func(t *testing.T) {
		msg := &UserMessage{
			Content: []ContentBlock{
				&TextBlock{TextContent: "ok"},
			},
			ToolUseResult: map[string]any{
				"tool_name": "Bash",
				"exit_code": float64(0),
			},
		}

		data, err := json.Marshal(msg)
		if err != nil {
			t.Fatalf("marshal failed: %v", err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}

		toolUseResult, ok := decoded["tool_use_result"].(map[string]any)
		if !ok {
			t.Fatalf("expected tool_use_result map, got %T", decoded["tool_use_result"])
		}
		if toolUseResult["tool_name"] != "Bash" {
			t.Errorf("got tool_name %v, want Bash", toolUseResult["tool_name"])
		}
	})
}

// TestAssistantMessage tests AssistantMessage methods.
func TestAssistantMessage(t *testing.T) {
	t.Run("MessageType returns assistant", func(t *testing.T) {
		msg := &AssistantMessage{}
		if msg.MessageType() != "assistant" {
			t.Errorf("Expected MessageType to return 'assistant', got %q", msg.MessageType())
		}
	})

	t.Run("Text concatenates text blocks", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{TextContent: "I can "},
				&TextBlock{TextContent: "help"},
			},
		}
		expected := "I can help"
		if msg.Text() != expected {
			t.Errorf("Expected %q, got %q", expected, msg.Text())
		}
	})

	t.Run("ToolCalls returns tool use blocks", func(t *testing.T) {
		tool1 := &ToolUseBlock{ID: "1", Name: "calc"}
		tool2 := &ToolUseBlock{ID: "2", Name: "search"}
		msg := &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{TextContent: "text"},
				tool1,
				tool2,
			},
		}
		tools := msg.ToolCalls()
		if len(tools) != 2 {
			t.Errorf("Expected 2 tool calls, got %d", len(tools))
		}
		if tools[0] != tool1 || tools[1] != tool2 {
			t.Error("Tool calls not returned in correct order")
		}
	})

	t.Run("GetThinking returns thinking content", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{TextContent: "text"},
				&ThinkingBlock{ThinkingContent: "deep thought"},
			},
		}
		if msg.GetThinking() != "deep thought" {
			t.Errorf("Expected 'deep thought', got %q", msg.GetThinking())
		}
	})

	t.Run("GetThinking returns empty when no thinking block", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{TextContent: "text"},
			},
		}
		if msg.GetThinking() != "" {
			t.Errorf("Expected empty string, got %q", msg.GetThinking())
		}
	})

	t.Run("Thinking is alias for GetThinking", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []ContentBlock{
				&ThinkingBlock{ThinkingContent: "reasoning"},
			},
		}
		if msg.Thinking() != msg.GetThinking() {
			t.Error("Thinking() and GetThinking() should return same value")
		}
	})

	t.Run("HasToolCalls returns true when tools present", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []ContentBlock{
				&ToolUseBlock{ID: "1", Name: "test"},
			},
		}
		if !msg.HasToolCalls() {
			t.Error("Expected HasToolCalls to return true")
		}
	})

	t.Run("HasToolCalls returns false when no tools", func(t *testing.T) {
		msg := &AssistantMessage{
			Content: []ContentBlock{
				&TextBlock{TextContent: "text"},
			},
		}
		if msg.HasToolCalls() {
			t.Error("Expected HasToolCalls to return false")
		}
	})
}

// TestSystemMessage tests SystemMessage methods.
func TestSystemMessage(t *testing.T) {
	t.Run("MessageType returns system", func(t *testing.T) {
		msg := &SystemMessage{}
		if msg.MessageType() != "system" {
			t.Errorf("Expected MessageType to return 'system', got %q", msg.MessageType())
		}
	})
}

// TestResultMessage tests ResultMessage methods.
func TestResultMessage(t *testing.T) {
	t.Run("MessageType returns result", func(t *testing.T) {
		msg := &ResultMessage{}
		if msg.MessageType() != "result" {
			t.Errorf("Expected MessageType to return 'result', got %q", msg.MessageType())
		}
	})

	t.Run("IsSuccess returns true for successful result", func(t *testing.T) {
		msg := &ResultMessage{
			Subtype: "success",
			IsError: false,
		}
		if !msg.IsSuccess() {
			t.Error("Expected IsSuccess to return true")
		}
	})

	t.Run("IsSuccess returns false when IsError is true", func(t *testing.T) {
		msg := &ResultMessage{
			Subtype: "success",
			IsError: true,
		}
		if msg.IsSuccess() {
			t.Error("Expected IsSuccess to return false when IsError is true")
		}
	})

	t.Run("IsSuccess returns false for non-success subtype", func(t *testing.T) {
		msg := &ResultMessage{
			Subtype: "error",
			IsError: false,
		}
		if msg.IsSuccess() {
			t.Error("Expected IsSuccess to return false for non-success subtype")
		}
	})

	t.Run("Cost returns value when TotalCostUSD is set", func(t *testing.T) {
		cost := 0.05
		msg := &ResultMessage{
			TotalCostUSD: &cost,
		}
		if msg.Cost() != 0.05 {
			t.Errorf("Expected Cost to return 0.05, got %f", msg.Cost())
		}
	})

	t.Run("Cost returns zero when TotalCostUSD is nil", func(t *testing.T) {
		msg := &ResultMessage{}
		if msg.Cost() != 0 {
			t.Errorf("Expected Cost to return 0, got %f", msg.Cost())
		}
	})
}

// TestStreamEvent tests StreamEvent methods.
func TestStreamEvent(t *testing.T) {
	t.Run("MessageType returns stream_event", func(t *testing.T) {
		msg := &StreamEvent{}
		if msg.MessageType() != "stream_event" {
			t.Errorf("Expected MessageType to return 'stream_event', got %q", msg.MessageType())
		}
	})
}

// TestPermissionResult tests permission result types.
func TestPermissionResult(t *testing.T) {
	t.Run("PermissionResultAllow implements interface", func(t *testing.T) {
		var _ PermissionResult = &PermissionResultAllow{}
		// Test basic serialization
		result := PermissionResultAllow{
			Behavior: "allow",
		}
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal PermissionResultAllow: %v", err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if decoded["behavior"] != "allow" {
			t.Errorf("Expected behavior 'allow', got %v", decoded["behavior"])
		}
	})

	t.Run("PermissionResultDeny implements interface", func(t *testing.T) {
		var _ PermissionResult = &PermissionResultDeny{}
		// Test basic serialization
		result := PermissionResultDeny{
			Behavior: "deny",
			Message:  "unauthorized",
		}
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal PermissionResultDeny: %v", err)
		}

		var decoded map[string]any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if decoded["behavior"] != "deny" {
			t.Errorf("Expected behavior 'deny', got %v", decoded["behavior"])
		}
		if decoded["message"] != "unauthorized" {
			t.Errorf("Expected message 'unauthorized', got %v", decoded["message"])
		}
	})
}
