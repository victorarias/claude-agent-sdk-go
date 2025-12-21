package sdk

import (
	"encoding/json"
	"fmt"
)

// ContentBlock represents a block of content in a message.
type ContentBlock interface {
	BlockType() string
}

// TextBlock contains text content.
type TextBlock struct {
	Text string `json:"text"`
}

func (b *TextBlock) BlockType() string { return "text" }

// ThinkingBlock contains Claude's thinking content.
type ThinkingBlock struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

func (b *ThinkingBlock) BlockType() string { return "thinking" }

// ToolUseBlock represents a tool invocation.
type ToolUseBlock struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Input map[string]any `json:"input"`
}

func (b *ToolUseBlock) BlockType() string { return "tool_use" }

// ToolResultBlock contains the result of a tool execution.
type ToolResultBlock struct {
	ToolUseID string `json:"tool_use_id"`
	Content   any    `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

func (b *ToolResultBlock) BlockType() string { return "tool_result" }

// ParseContentBlock parses a raw JSON map into a ContentBlock.
func ParseContentBlock(raw map[string]any) (ContentBlock, error) {
	blockType, ok := raw["type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'type' field")
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal block: %w", err)
	}

	switch blockType {
	case "text":
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse text block: %w", err)
		}
		return &block, nil

	case "thinking":
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse thinking block: %w", err)
		}
		return &block, nil

	case "tool_use":
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_use block: %w", err)
		}
		return &block, nil

	case "tool_result":
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, fmt.Errorf("failed to parse tool_result block: %w", err)
		}
		return &block, nil

	default:
		return nil, fmt.Errorf("unknown block type: %s", blockType)
	}
}

// Message represents a message in the conversation.
type Message interface {
	MessageType() string
}

// UserMessage represents a user's message.
type UserMessage struct {
	Content         any     `json:"content"` // string or []ContentBlock
	UUID            *string `json:"uuid,omitempty"`
	ParentToolUseID *string `json:"parent_tool_use_id,omitempty"`
}

func (m *UserMessage) MessageType() string { return "user" }

// AssistantMessageError represents error types for assistant messages.
type AssistantMessageError string

const (
	ErrorAuthenticationFailed AssistantMessageError = "authentication_failed"
	ErrorBillingError         AssistantMessageError = "billing_error"
	ErrorRateLimit            AssistantMessageError = "rate_limit"
	ErrorInvalidRequest       AssistantMessageError = "invalid_request"
	ErrorServerError          AssistantMessageError = "server_error"
	ErrorUnknown              AssistantMessageError = "unknown"
)

// AssistantMessage represents Claude's response.
type AssistantMessage struct {
	Content         []ContentBlock         `json:"content"`
	Model           string                 `json:"model"`
	ParentToolUseID *string                `json:"parent_tool_use_id,omitempty"`
	Error           *AssistantMessageError `json:"error,omitempty"`
}

func (m *AssistantMessage) MessageType() string { return "assistant" }

// Text returns all text content concatenated from an AssistantMessage.
func (m *AssistantMessage) Text() string {
	var result string
	for _, block := range m.Content {
		if textBlock, ok := block.(*TextBlock); ok {
			result += textBlock.Text
		}
	}
	return result
}

// ToolCalls returns all tool use blocks from an AssistantMessage.
func (m *AssistantMessage) ToolCalls() []*ToolUseBlock {
	var tools []*ToolUseBlock
	for _, block := range m.Content {
		if toolBlock, ok := block.(*ToolUseBlock); ok {
			tools = append(tools, toolBlock)
		}
	}
	return tools
}

// Thinking returns the thinking content if present.
func (m *AssistantMessage) Thinking() string {
	for _, block := range m.Content {
		if thinkingBlock, ok := block.(*ThinkingBlock); ok {
			return thinkingBlock.Thinking
		}
	}
	return ""
}

// SystemMessage represents a system message.
type SystemMessage struct {
	Subtype string         `json:"subtype"`
	Data    map[string]any `json:"data,omitempty"`
}

func (m *SystemMessage) MessageType() string { return "system" }

// ResultMessage represents the final result of a query.
type ResultMessage struct {
	Subtype          string         `json:"subtype"`
	DurationMS       int            `json:"duration_ms"`
	DurationAPI      int            `json:"duration_api_ms"`
	IsError          bool           `json:"is_error"`
	NumTurns         int            `json:"num_turns"`
	SessionID        string         `json:"session_id"`
	TotalCostUSD     *float64       `json:"total_cost_usd,omitempty"`
	Usage            map[string]any `json:"usage,omitempty"`
	Result           *string        `json:"result,omitempty"`
	StructuredOutput any            `json:"structured_output,omitempty"`
}

func (m *ResultMessage) MessageType() string { return "result" }

// IsSuccess returns true if the result is successful.
func (m *ResultMessage) IsSuccess() bool {
	return !m.IsError && m.Subtype == "success"
}

// Cost returns the cost in USD.
func (m *ResultMessage) Cost() float64 {
	if m.TotalCostUSD != nil {
		return *m.TotalCostUSD
	}
	return 0
}

// StreamEvent represents a streaming event with partial updates.
type StreamEvent struct {
	UUID            string         `json:"uuid"`
	SessionID       string         `json:"session_id"`
	Event           map[string]any `json:"event"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
}

func (m *StreamEvent) MessageType() string { return "stream_event" }
