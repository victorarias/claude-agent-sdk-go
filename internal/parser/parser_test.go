package parser

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestParseMessage_System(t *testing.T) {
	raw := map[string]any{
		"type":    "system",
		"subtype": "init",
		"data": map[string]any{
			"version":    "2.0.0",
			"session_id": "test_123",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	sys, ok := msg.(*SystemMessage)
	if !ok {
		t.Fatalf("expected *SystemMessage, got %T", msg)
	}

	if sys.Subtype != "init" {
		t.Errorf("got subtype %q, want init", sys.Subtype)
	}

	if sys.SessionID != "test_123" {
		t.Errorf("got session_id %q, want test_123", sys.SessionID)
	}
}

func TestParseMessage_Assistant(t *testing.T) {
	raw := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "Hello!"},
			},
			"model": "claude-sonnet-4-5",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	asst, ok := msg.(*AssistantMessage)
	if !ok {
		t.Fatalf("expected *AssistantMessage, got %T", msg)
	}

	if asst.Model != "claude-sonnet-4-5" {
		t.Errorf("got model %q, want claude-sonnet-4-5", asst.Model)
	}

	if asst.Text() != "Hello!" {
		t.Errorf("got text %q, want Hello!", asst.Text())
	}
}

func TestParseMessage_Result(t *testing.T) {
	raw := map[string]any{
		"type":           "result",
		"subtype":        "success",
		"duration_ms":    float64(1000),
		"session_id":     "test_123",
		"total_cost_usd": float64(0.001),
		"num_turns":      float64(5),
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	result, ok := msg.(*ResultMessage)
	if !ok {
		t.Fatalf("expected *ResultMessage, got %T", msg)
	}

	if result.Subtype != "success" {
		t.Errorf("got subtype %q, want success", result.Subtype)
	}

	if result.TotalCostUSD == nil || *result.TotalCostUSD != 0.001 {
		t.Errorf("got cost %v, want 0.001", result.TotalCostUSD)
	}
}

func TestParseMessage_User(t *testing.T) {
	raw := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": "Hello Claude!",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	user, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	if user.Text() != "Hello Claude!" {
		t.Errorf("got text %q, want Hello Claude!", user.Text())
	}
}

func TestParseMessage_StreamEvent(t *testing.T) {
	raw := map[string]any{
		"type":       "stream_event",
		"uuid":       "event_123",
		"session_id": "sess_456",
		"event": map[string]any{
			"type":  "content_block_delta",
			"index": float64(0),
			"delta": map[string]any{
				"type": "text_delta",
				"text": "Hello",
			},
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	event, ok := msg.(*StreamEvent)
	if !ok {
		t.Fatalf("expected *StreamEvent, got %T", msg)
	}

	if event.UUID != "event_123" {
		t.Errorf("got uuid %q, want event_123", event.UUID)
	}
	if event.SessionID != "sess_456" {
		t.Errorf("got session_id %q, want sess_456", event.SessionID)
	}
	if event.EventType != "content_block_delta" {
		t.Errorf("got event_type %q, want content_block_delta", event.EventType)
	}
	if event.Index == nil || *event.Index != 0 {
		t.Error("expected index 0")
	}
}

func TestParseMessage_WithParentToolUseID(t *testing.T) {
	raw := map[string]any{
		"type":               "user",
		"uuid":               "msg_123",
		"parent_tool_use_id": "tool_456",
		"message": map[string]any{
			"role":    "user",
			"content": "Subagent response",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	user, ok := msg.(*UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	if user.UUID != "msg_123" {
		t.Errorf("got uuid %q, want msg_123", user.UUID)
	}
	if user.ParentToolUseID == nil || *user.ParentToolUseID != "tool_456" {
		t.Error("expected parent_tool_use_id tool_456")
	}
}

func TestParseContentBlock(t *testing.T) {
	tests := []struct {
		name string
		raw  map[string]any
		want string // content type
	}{
		{
			name: "text",
			raw:  map[string]any{"type": "text", "text": "hello"},
			want: "text",
		},
		{
			name: "thinking",
			raw:  map[string]any{"type": "thinking", "thinking": "hmm"},
			want: "thinking",
		},
		{
			name: "tool_use",
			raw: map[string]any{
				"type":  "tool_use",
				"id":    "tool_1",
				"name":  "Bash",
				"input": map[string]any{"command": "ls"},
			},
			want: "tool_use",
		},
		{
			name: "tool_result",
			raw: map[string]any{
				"type":        "tool_result",
				"tool_use_id": "tool_1",
				"content":     "file1.txt\nfile2.txt",
			},
			want: "tool_result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block, err := parseContentBlock(tt.raw)
			if err != nil {
				t.Fatalf("parseContentBlock failed: %v", err)
			}
			if block.Type() != tt.want {
				t.Errorf("got type %q, want %q", block.Type(), tt.want)
			}
		})
	}
}

func TestParseMessage_UnknownType(t *testing.T) {
	raw := map[string]any{
		"type": "unknown_type",
	}

	_, err := ParseMessage(raw)
	if err == nil {
		t.Error("expected error for unknown message type")
	}
}
