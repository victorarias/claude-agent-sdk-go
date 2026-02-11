// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

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

	sys, ok := msg.(*types.SystemMessage)
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

	asst, ok := msg.(*types.AssistantMessage)
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

func TestParseMessage_Assistant_MetadataFields(t *testing.T) {
	raw := map[string]any{
		"type":       "assistant",
		"uuid":       "assistant_1",
		"session_id": "sess_1",
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

	asst, ok := msg.(*types.AssistantMessage)
	if !ok {
		t.Fatalf("expected *AssistantMessage, got %T", msg)
	}
	if asst.UUID != "assistant_1" {
		t.Errorf("got uuid %q, want assistant_1", asst.UUID)
	}
	if asst.SessionID != "sess_1" {
		t.Errorf("got session_id %q, want sess_1", asst.SessionID)
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

	result, ok := msg.(*types.ResultMessage)
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

func TestParseMessage_Result_ExpandedFields(t *testing.T) {
	raw := map[string]any{
		"type":            "result",
		"subtype":         "success",
		"uuid":            "result_1",
		"duration_ms":     float64(1000),
		"duration_api_ms": float64(600),
		"is_error":        false,
		"num_turns":       float64(2),
		"session_id":      "sess_1",
		"stop_reason":     "end_turn",
		"result":          "ok",
		"structured_output": map[string]any{
			"value": "done",
		},
		"usage": map[string]any{
			"input_tokens": float64(10),
		},
		"modelUsage": map[string]any{
			"claude-sonnet-4-5": map[string]any{
				"inputTokens":              float64(10),
				"outputTokens":             float64(5),
				"cacheReadInputTokens":     float64(1),
				"cacheCreationInputTokens": float64(2),
				"webSearchRequests":        float64(0),
				"costUSD":                  float64(0.001),
				"contextWindow":            float64(200000),
				"maxOutputTokens":          float64(8192),
			},
		},
		"permission_denials": []any{
			map[string]any{
				"tool_name":   "Bash",
				"tool_use_id": "tool_1",
				"tool_input":  map[string]any{"command": "rm -rf /"},
			},
		},
		"errors": []any{"denied"},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	result, ok := msg.(*types.ResultMessage)
	if !ok {
		t.Fatalf("expected *ResultMessage, got %T", msg)
	}
	if result.UUID != "result_1" {
		t.Fatalf("expected uuid=result_1, got %s", result.UUID)
	}
	if result.StopReason == nil || *result.StopReason != "end_turn" {
		t.Fatalf("unexpected stop_reason: %v", result.StopReason)
	}
	if result.Result == nil || *result.Result != "ok" {
		t.Fatalf("unexpected result: %v", result.Result)
	}
	if len(result.ModelUsage) != 1 {
		t.Fatalf("expected 1 modelUsage entry, got %d", len(result.ModelUsage))
	}
	if len(result.PermissionDenials) != 1 || result.PermissionDenials[0].ToolUseID != "tool_1" {
		t.Fatalf("unexpected permission_denials: %+v", result.PermissionDenials)
	}
	if len(result.Errors) != 1 || result.Errors[0] != "denied" {
		t.Fatalf("unexpected errors: %+v", result.Errors)
	}
	structured, ok := result.StructuredOutput.(map[string]any)
	if !ok {
		t.Fatalf("expected structured output map, got %T", result.StructuredOutput)
	}
	if structured["value"] != "done" {
		t.Fatalf("got structured_output.value %v, want done", structured["value"])
	}
	if result.Usage == nil || result.Usage["input_tokens"] != float64(10) {
		t.Fatalf("unexpected usage payload: %+v", result.Usage)
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

	user, ok := msg.(*types.UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	if user.Text() != "Hello Claude!" {
		t.Errorf("got text %q, want Hello Claude!", user.Text())
	}
}

func TestParseMessage_User_MetadataFields(t *testing.T) {
	raw := map[string]any{
		"type":        "user",
		"uuid":        "user_1",
		"session_id":  "sess_1",
		"isSynthetic": true,
		"isReplay":    true,
		"message": map[string]any{
			"role":    "user",
			"content": "Hello Claude!",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	user, ok := msg.(*types.UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}
	if user.UUID != "user_1" {
		t.Errorf("got uuid %q, want user_1", user.UUID)
	}
	if user.SessionID != "sess_1" {
		t.Errorf("got session_id %q, want sess_1", user.SessionID)
	}
	if !user.IsSynthetic {
		t.Error("expected isSynthetic=true")
	}
	if !user.IsReplay {
		t.Error("expected isReplay=true")
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

	event, ok := msg.(*types.StreamEvent)
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

func TestParseMessage_StreamEventTopLevelFields(t *testing.T) {
	raw := map[string]any{
		"type":               "stream_event",
		"uuid":               "event_top_1",
		"session_id":         "sess_top_1",
		"event_type":         "content_block_delta",
		"index":              float64(2),
		"delta":              map[string]any{"type": "text_delta", "text": "abc"},
		"parent_tool_use_id": "tool_parent_1",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	event, ok := msg.(*types.StreamEvent)
	if !ok {
		t.Fatalf("expected *StreamEvent, got %T", msg)
	}
	if event.EventType != "content_block_delta" {
		t.Fatalf("expected event_type=content_block_delta, got %q", event.EventType)
	}
	if event.Index == nil || *event.Index != 2 {
		t.Fatalf("expected index=2, got %+v", event.Index)
	}
	if event.Delta == nil || event.Delta["text"] != "abc" {
		t.Fatalf("unexpected delta payload: %+v", event.Delta)
	}
	if event.ParentToolUseID == nil || *event.ParentToolUseID != "tool_parent_1" {
		t.Fatalf("unexpected parent_tool_use_id: %+v", event.ParentToolUseID)
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

	user, ok := msg.(*types.UserMessage)
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

func TestParseMessage_UserWithToolUseResult(t *testing.T) {
	raw := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role":    "user",
			"content": "Done",
		},
		"tool_use_result": map[string]any{
			"tool_name": "Bash",
			"exit_code": float64(0),
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	user, ok := msg.(*types.UserMessage)
	if !ok {
		t.Fatalf("expected *UserMessage, got %T", msg)
	}

	if user.ToolUseResult == nil {
		t.Fatal("expected tool_use_result to be parsed")
	}
	if user.ToolUseResult["tool_name"] != "Bash" {
		t.Errorf("got tool_name %v, want Bash", user.ToolUseResult["tool_name"])
	}
}

func TestParseMessage_AuthStatus(t *testing.T) {
	raw := map[string]any{
		"type":             "auth_status",
		"isAuthenticating": true,
		"output":           []any{"Starting login"},
		"session_id":       "sess_1",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	auth, ok := msg.(*types.AuthStatusMessage)
	if !ok {
		t.Fatalf("expected *AuthStatusMessage, got %T", msg)
	}
	if !auth.IsAuthenticating {
		t.Fatal("expected IsAuthenticating=true")
	}
	if len(auth.Output) != 1 || auth.Output[0] != "Starting login" {
		t.Fatalf("unexpected output payload: %+v", auth.Output)
	}
}

func TestParseMessage_ToolProgress(t *testing.T) {
	raw := map[string]any{
		"type":                 "tool_progress",
		"tool_use_id":          "tool_1",
		"tool_name":            "Bash",
		"elapsed_time_seconds": float64(3.5),
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	progress, ok := msg.(*types.ToolProgressMessage)
	if !ok {
		t.Fatalf("expected *ToolProgressMessage, got %T", msg)
	}
	if progress.ToolUseID != "tool_1" {
		t.Fatalf("expected tool_use_id=tool_1, got %s", progress.ToolUseID)
	}
	if progress.ElapsedTimeSeconds != 3.5 {
		t.Fatalf("expected elapsed_time_seconds=3.5, got %v", progress.ElapsedTimeSeconds)
	}
}

func TestParseMessage_ToolUseSummary(t *testing.T) {
	raw := map[string]any{
		"type":    "tool_use_summary",
		"summary": "Summarized operations",
		"preceding_tool_use_ids": []any{
			"tool_1",
			"tool_2",
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	summary, ok := msg.(*types.ToolUseSummaryMessage)
	if !ok {
		t.Fatalf("expected *ToolUseSummaryMessage, got %T", msg)
	}
	if summary.Summary != "Summarized operations" {
		t.Fatalf("unexpected summary: %s", summary.Summary)
	}
	if len(summary.PrecedingToolUseIDs) != 2 {
		t.Fatalf("unexpected tool id list: %+v", summary.PrecedingToolUseIDs)
	}
}

func TestParseMessage_SystemTaskNotification(t *testing.T) {
	raw := map[string]any{
		"type":        "system",
		"subtype":     "task_notification",
		"task_id":     "task_1",
		"status":      "completed",
		"output_file": "/tmp/task.out",
		"summary":     "Task complete",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	task, ok := msg.(*types.TaskNotificationMessage)
	if !ok {
		t.Fatalf("expected *TaskNotificationMessage, got %T", msg)
	}
	if task.TaskID != "task_1" {
		t.Fatalf("expected task_id=task_1, got %s", task.TaskID)
	}
}

func TestParseMessage_System_MetadataFields(t *testing.T) {
	raw := map[string]any{
		"type":       "system",
		"subtype":    "init",
		"uuid":       "sys_1",
		"session_id": "sess_1",
		"version":    "2.0.0",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	sys, ok := msg.(*types.SystemMessage)
	if !ok {
		t.Fatalf("expected *SystemMessage, got %T", msg)
	}
	if sys.UUID != "sys_1" {
		t.Fatalf("expected uuid=sys_1, got %s", sys.UUID)
	}
	if sys.SessionID != "sess_1" {
		t.Fatalf("expected session_id=sess_1, got %s", sys.SessionID)
	}
}

func TestParseMessage_SystemFilesPersisted(t *testing.T) {
	raw := map[string]any{
		"type":         "system",
		"subtype":      "files_persisted",
		"processed_at": "2026-02-07T00:00:00Z",
		"files": []any{
			map[string]any{"filename": "a.go", "file_id": "f1"},
		},
		"failed": []any{
			map[string]any{"filename": "b.go", "error": "permission denied"},
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	persisted, ok := msg.(*types.FilesPersistedMessage)
	if !ok {
		t.Fatalf("expected *FilesPersistedMessage, got %T", msg)
	}
	if len(persisted.Files) != 1 || persisted.Files[0].Filename != "a.go" {
		t.Fatalf("unexpected files payload: %+v", persisted.Files)
	}
	if len(persisted.Failed) != 1 || persisted.Failed[0].Filename != "b.go" {
		t.Fatalf("unexpected failed payload: %+v", persisted.Failed)
	}
}

func TestParseMessage_SystemStatus(t *testing.T) {
	raw := map[string]any{
		"type":           "system",
		"subtype":        "status",
		"status":         "compacting",
		"permissionMode": "plan",
		"uuid":           "sys_1",
		"session_id":     "sess_1",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	status, ok := msg.(*types.StatusMessage)
	if !ok {
		t.Fatalf("expected *StatusMessage, got %T", msg)
	}
	if status.Subtype != "status" {
		t.Fatalf("expected subtype=status, got %s", status.Subtype)
	}
	if status.Status != "compacting" {
		t.Fatalf("expected status=compacting, got %v", status.Status)
	}
	if status.PermissionMode != types.PermissionPlan {
		t.Fatalf("expected permissionMode=plan, got %s", status.PermissionMode)
	}
}

func TestParseMessage_SystemCompactBoundary(t *testing.T) {
	raw := map[string]any{
		"type":       "system",
		"subtype":    "compact_boundary",
		"uuid":       "sys_2",
		"session_id": "sess_2",
		"compact_metadata": map[string]any{
			"trigger":    "auto",
			"pre_tokens": float64(12345),
		},
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	compact, ok := msg.(*types.CompactBoundaryMessage)
	if !ok {
		t.Fatalf("expected *CompactBoundaryMessage, got %T", msg)
	}
	if compact.CompactMetadata.Trigger != "auto" {
		t.Fatalf("expected trigger=auto, got %s", compact.CompactMetadata.Trigger)
	}
	if compact.CompactMetadata.PreTokens != 12345 {
		t.Fatalf("expected pre_tokens=12345, got %d", compact.CompactMetadata.PreTokens)
	}
}

func TestParseMessage_SystemHookStarted(t *testing.T) {
	raw := map[string]any{
		"type":       "system",
		"subtype":    "hook_started",
		"hook_id":    "hook-1",
		"hook_name":  "pre_tool_check",
		"hook_event": "PreToolUse",
		"session_id": "sess-1",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	hook, ok := msg.(*types.HookStartedMessage)
	if !ok {
		t.Fatalf("expected *HookStartedMessage, got %T", msg)
	}
	if hook.HookID != "hook-1" || hook.HookEvent != "PreToolUse" {
		t.Fatalf("unexpected hook_started payload: %+v", hook)
	}
}

func TestParseMessage_SystemHookProgress(t *testing.T) {
	raw := map[string]any{
		"type":       "system",
		"subtype":    "hook_progress",
		"hook_id":    "hook-2",
		"hook_name":  "setup",
		"hook_event": "Setup",
		"stdout":     "checking...",
		"stderr":     "",
		"output":     "partial output",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	hook, ok := msg.(*types.HookProgressMessage)
	if !ok {
		t.Fatalf("expected *HookProgressMessage, got %T", msg)
	}
	if hook.Stdout != "checking..." || hook.Output != "partial output" {
		t.Fatalf("unexpected hook_progress payload: %+v", hook)
	}
}

func TestParseMessage_SystemHookResponse(t *testing.T) {
	raw := map[string]any{
		"type":       "system",
		"subtype":    "hook_response",
		"hook_id":    "hook-3",
		"hook_name":  "cleanup",
		"hook_event": "Stop",
		"outcome":    "success",
		"exit_code":  float64(0),
		"stdout":     "done",
		"stderr":     "",
	}

	msg, err := ParseMessage(raw)
	if err != nil {
		t.Fatalf("ParseMessage failed: %v", err)
	}

	hook, ok := msg.(*types.HookResponseMessage)
	if !ok {
		t.Fatalf("expected *HookResponseMessage, got %T", msg)
	}
	if hook.ExitCode == nil || *hook.ExitCode != 0 || hook.Outcome != "success" {
		t.Fatalf("unexpected hook_response payload: %+v", hook)
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

func TestParseContentBlock_ToolResultStructuredContent(t *testing.T) {
	raw := map[string]any{
		"type":        "tool_result",
		"tool_use_id": "tool_1",
		"content": []any{
			map[string]any{"type": "text", "text": "line1"},
			map[string]any{"type": "text", "text": "line2"},
		},
	}

	block, err := parseContentBlock(raw)
	if err != nil {
		t.Fatalf("parseContentBlock failed: %v", err)
	}

	toolResult, ok := block.(*types.ToolResultBlock)
	if !ok {
		t.Fatalf("expected *ToolResultBlock, got %T", block)
	}
	expected := `[{"text":"line1","type":"text"},{"text":"line2","type":"text"}]`
	if toolResult.ResultContent != expected {
		t.Fatalf("unexpected structured tool result content: got %q want %q", toolResult.ResultContent, expected)
	}
}

func TestParseMessage_AssistantInvalidOnlyContentFails(t *testing.T) {
	raw := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"model": "claude-sonnet-4-5",
			"content": []any{
				map[string]any{"type": "unknown_block_type"},
			},
		},
	}

	if _, err := ParseMessage(raw); err == nil {
		t.Fatal("expected parse failure for assistant with only invalid content blocks")
	}
}

func TestParseMessage_UserInvalidOnlyContentFails(t *testing.T) {
	raw := map[string]any{
		"type": "user",
		"message": map[string]any{
			"role": "user",
			"content": []any{
				map[string]any{"type": "unknown_block_type"},
			},
		},
	}

	if _, err := ParseMessage(raw); err == nil {
		t.Fatal("expected parse failure for user with only invalid content blocks")
	}
}

func TestParseContentBlock_ThinkingWithSignature(t *testing.T) {
	raw := map[string]any{
		"type":      "thinking",
		"thinking":  "some thought",
		"signature": "abc123",
	}

	block, err := parseContentBlock(raw)
	if err != nil {
		t.Fatalf("parseContentBlock failed: %v", err)
	}

	thinkingBlock, ok := block.(*types.ThinkingBlock)
	if !ok {
		t.Fatalf("expected *ThinkingBlock, got %T", block)
	}

	if thinkingBlock.ThinkingContent != "some thought" {
		t.Errorf("got thinking %q, want 'some thought'", thinkingBlock.ThinkingContent)
	}

	if thinkingBlock.Signature != "abc123" {
		t.Errorf("got signature %q, want 'abc123'", thinkingBlock.Signature)
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
