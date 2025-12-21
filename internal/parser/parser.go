package parser

import (
	"fmt"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// ParseMessage parses a raw message map into a typed Message.
func ParseMessage(raw map[string]any) (Message, error) {
	msgType, _ := raw["type"].(string)

	switch msgType {
	case "system":
		return parseSystemMessage(raw)
	case "assistant":
		return parseAssistantMessage(raw)
	case "user":
		return parseUserMessage(raw)
	case "result":
		return parseResultMessage(raw)
	case "stream_event":
		return parseStreamEvent(raw)
	default:
		return nil, &MessageParseError{
			Message: fmt.Sprintf("unknown message type: %s", msgType),
			Data:    raw,
		}
	}
}

// parseStreamEvent parses a StreamEvent for partial message updates.
func parseStreamEvent(raw map[string]any) (*StreamEvent, error) {
	event := &StreamEvent{
		UUID:      getString(raw, "uuid"),
		SessionID: getString(raw, "session_id"),
	}

	// Extract parent_tool_use_id if present
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		event.ParentToolUseID = &parentID
	}

	// Parse the nested event data
	if eventData, ok := raw["event"].(map[string]any); ok {
		event.Event = eventData
		event.EventType = getString(eventData, "type")

		// Extract index if present (for content_block events)
		if idx, ok := eventData["index"].(float64); ok {
			idxInt := int(idx)
			event.Index = &idxInt
		}

		// Extract delta if present
		if delta, ok := eventData["delta"].(map[string]any); ok {
			event.Delta = delta
		}
	}

	return event, nil
}

func parseSystemMessage(raw map[string]any) (*SystemMessage, error) {
	msg := &SystemMessage{
		Subtype: getString(raw, "subtype"),
	}

	if data, ok := raw["data"].(map[string]any); ok {
		msg.SessionID = getString(data, "session_id")
		msg.Version = getString(data, "version")
		msg.Data = data
	}

	return msg, nil
}

func parseAssistantMessage(raw map[string]any) (*AssistantMessage, error) {
	msg := &AssistantMessage{}

	// Extract parent_tool_use_id for subagent messages
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}

	// Extract error field for API error messages
	if errType, ok := raw["error"].(string); ok {
		err := AssistantMessageError(errType)
		msg.Error = &err
	}

	if msgData, ok := raw["message"].(map[string]any); ok {
		msg.Model = getString(msgData, "model")
		msg.StopReason = getString(msgData, "stop_reason")

		// Parse content blocks
		if content, ok := msgData["content"].([]any); ok {
			for _, item := range content {
				if blockRaw, ok := item.(map[string]any); ok {
					block, err := parseContentBlock(blockRaw)
					if err != nil {
						continue // Skip invalid blocks
					}
					msg.Content = append(msg.Content, block)
				}
			}
		}
	}

	return msg, nil
}

func parseUserMessage(raw map[string]any) (*UserMessage, error) {
	msg := &UserMessage{
		UUID: getString(raw, "uuid"),
	}

	// Extract parent_tool_use_id for subagent messages
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}

	if msgData, ok := raw["message"].(map[string]any); ok {
		msg.Role = getString(msgData, "role")

		// Content can be string or array
		switch c := msgData["content"].(type) {
		case string:
			msg.Content = []ContentBlock{&TextBlock{TextContent: c}}
		case []any:
			for _, item := range c {
				if blockRaw, ok := item.(map[string]any); ok {
					block, err := parseContentBlock(blockRaw)
					if err != nil {
						continue
					}
					msg.Content = append(msg.Content, block)
				}
			}
		}
	}

	return msg, nil
}

func parseResultMessage(raw map[string]any) (*ResultMessage, error) {
	msg := &ResultMessage{
		Subtype:   getString(raw, "subtype"),
		SessionID: getString(raw, "session_id"),
		IsError:   getBool(raw, "is_error"),
	}

	if dur, ok := raw["duration_ms"].(float64); ok {
		msg.DurationMS = int(dur)
	}
	if durAPI, ok := raw["duration_api_ms"].(float64); ok {
		msg.DurationAPI = int(durAPI)
	}
	if turns, ok := raw["num_turns"].(float64); ok {
		msg.NumTurns = int(turns)
	}
	if cost, ok := raw["total_cost_usd"].(float64); ok {
		msg.TotalCostUSD = &cost
	}

	return msg, nil
}

// parseContentBlock parses a raw JSON map into a ContentBlock.
// This is a simplified version that doesn't use JSON marshaling.
func parseContentBlock(raw map[string]any) (ContentBlock, error) {
	blockType, _ := raw["type"].(string)

	switch blockType {
	case "text":
		return &TextBlock{
			TextContent: getString(raw, "text"),
		}, nil

	case "thinking":
		return &ThinkingBlock{
			ThinkingContent: getString(raw, "thinking"),
		}, nil

	case "tool_use":
		input, _ := raw["input"].(map[string]any)
		return &ToolUseBlock{
			ID:        getString(raw, "id"),
			Name:      getString(raw, "name"),
			ToolInput: input,
		}, nil

	case "tool_result":
		var content string
		switch c := raw["content"].(type) {
		case string:
			content = c
		}
		return &ToolResultBlock{
			ToolUseID:     getString(raw, "tool_use_id"),
			ResultContent: content,
			IsError:       getBool(raw, "is_error"),
		}, nil

	default:
		return nil, fmt.Errorf("unknown content block type: %s", blockType)
	}
}

// Helper functions
func getString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}

func getBool(m map[string]any, key string) bool {
	v, _ := m[key].(bool)
	return v
}
