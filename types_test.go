package sdk

import (
	"encoding/json"
	"testing"
)

func TestTextBlock(t *testing.T) {
	block := TextBlock{Text: "hello"}
	if block.BlockType() != "text" {
		t.Errorf("got %q, want %q", block.BlockType(), "text")
	}
}

func TestToolUseBlock(t *testing.T) {
	block := ToolUseBlock{
		ID:    "tool_123",
		Name:  "Bash",
		Input: map[string]any{"command": "ls"},
	}
	if block.BlockType() != "tool_use" {
		t.Errorf("got %q, want %q", block.BlockType(), "tool_use")
	}
}

func TestThinkingBlock(t *testing.T) {
	block := ThinkingBlock{
		Thinking:  "Let me think...",
		Signature: "sig123",
	}
	if block.BlockType() != "thinking" {
		t.Errorf("got %q, want %q", block.BlockType(), "thinking")
	}
}

func TestToolResultBlock(t *testing.T) {
	block := ToolResultBlock{
		ToolUseID: "tool_123",
		Content:   "output",
		IsError:   false,
	}
	if block.BlockType() != "tool_result" {
		t.Errorf("got %q, want %q", block.BlockType(), "tool_result")
	}
}

func TestContentBlockJSON(t *testing.T) {
	input := `{"type":"text","text":"hello"}`
	var raw map[string]any
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		t.Fatal(err)
	}

	block, err := ParseContentBlock(raw)
	if err != nil {
		t.Fatal(err)
	}

	textBlock, ok := block.(*TextBlock)
	if !ok {
		t.Fatalf("expected *TextBlock, got %T", block)
	}
	if textBlock.Text != "hello" {
		t.Errorf("got %q, want %q", textBlock.Text, "hello")
	}
}

func TestUserMessage(t *testing.T) {
	msg := &UserMessage{Content: "hello"}
	if msg.MessageType() != "user" {
		t.Errorf("got %q, want %q", msg.MessageType(), "user")
	}
}

func TestAssistantMessage(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{&TextBlock{Text: "hello"}},
		Model:   "claude-sonnet-4-5",
	}
	if msg.MessageType() != "assistant" {
		t.Errorf("got %q, want %q", msg.MessageType(), "assistant")
	}
}

func TestSystemMessage(t *testing.T) {
	msg := &SystemMessage{
		Subtype: "init",
		Data:    map[string]any{"version": "1.0"},
	}
	if msg.MessageType() != "system" {
		t.Errorf("got %q, want %q", msg.MessageType(), "system")
	}
}

func TestResultMessage(t *testing.T) {
	msg := &ResultMessage{
		Subtype:      "success",
		DurationMS:   1000,
		DurationAPI:  800,
		IsError:      false,
		NumTurns:     3,
		SessionID:    "sess_123",
		TotalCostUSD: floatPtr(0.05),
	}
	if msg.MessageType() != "result" {
		t.Errorf("got %q, want %q", msg.MessageType(), "result")
	}
	if !msg.IsSuccess() {
		t.Error("expected IsSuccess() to return true")
	}
	if msg.Cost() != 0.05 {
		t.Errorf("got cost %f, want 0.05", msg.Cost())
	}
}

func TestStreamEvent(t *testing.T) {
	msg := &StreamEvent{
		UUID:      "uuid_123",
		SessionID: "sess_123",
		Event:     map[string]any{"type": "content_block_delta"},
	}
	if msg.MessageType() != "stream_event" {
		t.Errorf("got %q, want %q", msg.MessageType(), "stream_event")
	}
}

func TestAssistantMessageHelpers(t *testing.T) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "Hello "},
			&TextBlock{Text: "world"},
			&ThinkingBlock{Thinking: "Let me think..."},
			&ToolUseBlock{ID: "tool_1", Name: "Bash", Input: map[string]any{"command": "ls"}},
		},
		Model: "claude-sonnet-4-5",
	}

	if msg.Text() != "Hello world" {
		t.Errorf("got %q, want %q", msg.Text(), "Hello world")
	}

	if msg.Thinking() != "Let me think..." {
		t.Errorf("got %q, want %q", msg.Thinking(), "Let me think...")
	}

	tools := msg.ToolCalls()
	if len(tools) != 1 {
		t.Errorf("got %d tool calls, want 1", len(tools))
	}
}

func floatPtr(f float64) *float64 { return &f }

func TestHookEvent(t *testing.T) {
	events := []HookEvent{
		HookPreToolUse,
		HookPostToolUse,
		HookUserPromptSubmit,
		HookStop,
		HookSubagentStop,
		HookPreCompact,
	}

	if len(events) != 6 {
		t.Errorf("expected 6 hook events, got %d", len(events))
	}
}

func TestHookInput(t *testing.T) {
	input := &PreToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "sess_123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user",
			HookEventName:  "PreToolUse",
		},
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "ls"},
	}

	if input.SessionID != "sess_123" {
		t.Errorf("got %q, want %q", input.SessionID, "sess_123")
	}
	if input.HookEventName != "PreToolUse" {
		t.Errorf("got %q, want %q", input.HookEventName, "PreToolUse")
	}
}

func TestHookOutput(t *testing.T) {
	cont := true
	output := &HookOutput{
		Continue:   &cont,
		Decision:   "allow",
	}
	if output.Continue == nil || !*output.Continue {
		t.Error("expected Continue to be true")
	}
}

func boolPtr(b bool) *bool { return &b }
func intPtr(i int) *int    { return &i }
