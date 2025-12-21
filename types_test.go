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
