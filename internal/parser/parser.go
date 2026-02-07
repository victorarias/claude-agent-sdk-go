// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

// Package parser provides message parsing functionality for the Claude Agent SDK.
//
// This package handles the conversion of raw JSON-RPC messages from the Claude CLI
// into strongly-typed Go message structures. It supports all message types including
// system messages, user messages, assistant messages, stream events, and result messages.
//
// The parser automatically handles type inference and validation, ensuring that all
// messages conform to the expected schema before being returned to the client.
package parser

import (
	"fmt"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// ParseMessage parses a raw message map into a typed Message.
func ParseMessage(raw map[string]any) (types.Message, error) {
	msgType, _ := raw["type"].(string)

	switch msgType {
	case "system":
		return parseSystemMessage(raw)
	case "auth_status":
		return parseAuthStatusMessage(raw)
	case "assistant":
		return parseAssistantMessage(raw)
	case "user":
		return parseUserMessage(raw)
	case "result":
		return parseResultMessage(raw)
	case "stream_event":
		return parseStreamEvent(raw)
	case "tool_progress":
		return parseToolProgressMessage(raw)
	case "tool_use_summary":
		return parseToolUseSummaryMessage(raw)
	default:
		return nil, &types.MessageParseError{
			Message: fmt.Sprintf("unknown message type: %s", msgType),
			Data:    raw,
		}
	}
}

// parseStreamEvent parses a StreamEvent for partial message updates.
func parseStreamEvent(raw map[string]any) (*types.StreamEvent, error) {
	event := &types.StreamEvent{
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

func parseSystemMessage(raw map[string]any) (types.Message, error) {
	subtype := getString(raw, "subtype")

	switch subtype {
	case "task_notification":
		return parseTaskNotificationMessage(raw), nil
	case "files_persisted":
		return parseFilesPersistedMessage(raw), nil
	case "hook_started":
		return parseHookStartedMessage(raw), nil
	case "hook_progress":
		return parseHookProgressMessage(raw), nil
	case "hook_response":
		return parseHookResponseMessage(raw), nil
	}

	msg := &types.SystemMessage{
		Subtype: subtype,
	}

	if data, ok := raw["data"].(map[string]any); ok {
		msg.SessionID = getString(data, "session_id")
		msg.Version = getString(data, "version")
		msg.Data = data
	} else {
		msg.SessionID = getString(raw, "session_id")
		msg.Version = getString(raw, "version")
		msg.Data = raw
	}

	return msg, nil
}

func parseAuthStatusMessage(raw map[string]any) (*types.AuthStatusMessage, error) {
	msg := &types.AuthStatusMessage{
		IsAuthenticating: getBool(raw, "isAuthenticating"),
		Error:            getString(raw, "error"),
		UUID:             getString(raw, "uuid"),
		SessionID:        getString(raw, "session_id"),
		Output:           getStringSlice(raw, "output"),
	}
	return msg, nil
}

func parseToolProgressMessage(raw map[string]any) (*types.ToolProgressMessage, error) {
	msg := &types.ToolProgressMessage{
		ToolUseID: getString(raw, "tool_use_id"),
		ToolName:  getString(raw, "tool_name"),
		UUID:      getString(raw, "uuid"),
		SessionID: getString(raw, "session_id"),
	}
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}
	if elapsed, ok := raw["elapsed_time_seconds"].(float64); ok {
		msg.ElapsedTimeSeconds = elapsed
	}
	return msg, nil
}

func parseToolUseSummaryMessage(raw map[string]any) (*types.ToolUseSummaryMessage, error) {
	msg := &types.ToolUseSummaryMessage{
		Summary:             getString(raw, "summary"),
		PrecedingToolUseIDs: getStringSlice(raw, "preceding_tool_use_ids"),
		UUID:                getString(raw, "uuid"),
		SessionID:           getString(raw, "session_id"),
	}
	return msg, nil
}

func parseTaskNotificationMessage(raw map[string]any) *types.TaskNotificationMessage {
	return &types.TaskNotificationMessage{
		Subtype:    getString(raw, "subtype"),
		TaskID:     getString(raw, "task_id"),
		Status:     getString(raw, "status"),
		OutputFile: getString(raw, "output_file"),
		Summary:    getString(raw, "summary"),
		UUID:       getString(raw, "uuid"),
		SessionID:  getString(raw, "session_id"),
	}
}

func parseFilesPersistedMessage(raw map[string]any) *types.FilesPersistedMessage {
	msg := &types.FilesPersistedMessage{
		Subtype:     getString(raw, "subtype"),
		ProcessedAt: getString(raw, "processed_at"),
		UUID:        getString(raw, "uuid"),
		SessionID:   getString(raw, "session_id"),
	}

	if filesRaw, ok := raw["files"].([]any); ok {
		files := make([]types.FilesPersistedFile, 0, len(filesRaw))
		for _, item := range filesRaw {
			if m, ok := item.(map[string]any); ok {
				files = append(files, types.FilesPersistedFile{
					Filename: getString(m, "filename"),
					FileID:   getString(m, "file_id"),
				})
			}
		}
		msg.Files = files
	}

	if failedRaw, ok := raw["failed"].([]any); ok {
		failed := make([]types.FilesPersistedFailure, 0, len(failedRaw))
		for _, item := range failedRaw {
			if m, ok := item.(map[string]any); ok {
				failed = append(failed, types.FilesPersistedFailure{
					Filename: getString(m, "filename"),
					Error:    getString(m, "error"),
				})
			}
		}
		msg.Failed = failed
	}

	return msg
}

func parseHookStartedMessage(raw map[string]any) *types.HookStartedMessage {
	return &types.HookStartedMessage{
		Subtype:   getString(raw, "subtype"),
		HookID:    getString(raw, "hook_id"),
		HookName:  getString(raw, "hook_name"),
		HookEvent: getString(raw, "hook_event"),
		UUID:      getString(raw, "uuid"),
		SessionID: getString(raw, "session_id"),
	}
}

func parseHookProgressMessage(raw map[string]any) *types.HookProgressMessage {
	return &types.HookProgressMessage{
		Subtype:   getString(raw, "subtype"),
		HookID:    getString(raw, "hook_id"),
		HookName:  getString(raw, "hook_name"),
		HookEvent: getString(raw, "hook_event"),
		Stdout:    getString(raw, "stdout"),
		Stderr:    getString(raw, "stderr"),
		Output:    getString(raw, "output"),
		UUID:      getString(raw, "uuid"),
		SessionID: getString(raw, "session_id"),
	}
}

func parseHookResponseMessage(raw map[string]any) *types.HookResponseMessage {
	msg := &types.HookResponseMessage{
		Subtype:   getString(raw, "subtype"),
		HookID:    getString(raw, "hook_id"),
		HookName:  getString(raw, "hook_name"),
		HookEvent: getString(raw, "hook_event"),
		Output:    getString(raw, "output"),
		Stdout:    getString(raw, "stdout"),
		Stderr:    getString(raw, "stderr"),
		Outcome:   getString(raw, "outcome"),
		UUID:      getString(raw, "uuid"),
		SessionID: getString(raw, "session_id"),
	}
	if exitCode, ok := raw["exit_code"].(float64); ok {
		n := int(exitCode)
		msg.ExitCode = &n
	}
	return msg
}

func parseAssistantMessage(raw map[string]any) (*types.AssistantMessage, error) {
	msg := &types.AssistantMessage{}

	// Extract parent_tool_use_id for subagent messages
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}

	// Extract error field for API error messages
	if errType, ok := raw["error"].(string); ok {
		err := types.AssistantMessageError(errType)
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

func parseUserMessage(raw map[string]any) (*types.UserMessage, error) {
	msg := &types.UserMessage{
		UUID: getString(raw, "uuid"),
	}

	// Extract parent_tool_use_id for subagent messages
	if parentID, ok := raw["parent_tool_use_id"].(string); ok {
		msg.ParentToolUseID = &parentID
	}
	if toolUseResult, ok := raw["tool_use_result"].(map[string]any); ok {
		msg.ToolUseResult = toolUseResult
	}

	if msgData, ok := raw["message"].(map[string]any); ok {
		msg.Role = getString(msgData, "role")

		// Content can be string or array
		switch c := msgData["content"].(type) {
		case string:
			msg.Content = []types.ContentBlock{&types.TextBlock{TextContent: c}}
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

func parseResultMessage(raw map[string]any) (*types.ResultMessage, error) {
	msg := &types.ResultMessage{
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
func parseContentBlock(raw map[string]any) (types.ContentBlock, error) {
	blockType, _ := raw["type"].(string)

	switch blockType {
	case "text":
		return &types.TextBlock{
			TextContent: getString(raw, "text"),
		}, nil

	case "thinking":
		return &types.ThinkingBlock{
			ThinkingContent: getString(raw, "thinking"),
			Signature:       getString(raw, "signature"),
		}, nil

	case "tool_use":
		input, _ := raw["input"].(map[string]any)
		return &types.ToolUseBlock{
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
		return &types.ToolResultBlock{
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

func getStringSlice(m map[string]any, key string) []string {
	raw, ok := m[key].([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
