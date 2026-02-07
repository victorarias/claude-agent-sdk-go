// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

func TestUserMessage_JSONParityFields(t *testing.T) {
	msg := UserMessage{
		UUID:        "user-1",
		SessionID:   "sess-1",
		IsSynthetic: true,
		IsReplay:    true,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded["uuid"] != "user-1" {
		t.Fatalf("expected uuid=user-1, got %v", decoded["uuid"])
	}
	if decoded["session_id"] != "sess-1" {
		t.Fatalf("expected session_id=sess-1, got %v", decoded["session_id"])
	}
	if decoded["isSynthetic"] != true {
		t.Fatalf("expected isSynthetic=true, got %v", decoded["isSynthetic"])
	}
	if decoded["isReplay"] != true {
		t.Fatalf("expected isReplay=true, got %v", decoded["isReplay"])
	}
}

func TestAssistantMessage_JSONParityFields(t *testing.T) {
	msg := AssistantMessage{
		UUID:      "assistant-1",
		SessionID: "sess-1",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded["uuid"] != "assistant-1" {
		t.Fatalf("expected uuid=assistant-1, got %v", decoded["uuid"])
	}
	if decoded["session_id"] != "sess-1" {
		t.Fatalf("expected session_id=sess-1, got %v", decoded["session_id"])
	}
}

func TestResultMessage_JSONParityFields(t *testing.T) {
	stopReason := "end_turn"
	msg := ResultMessage{
		Subtype:    "success",
		UUID:       "result-1",
		SessionID:  "sess-1",
		StopReason: &stopReason,
		ModelUsage: map[string]ModelUsage{
			"claude-sonnet-4-5": {
				InputTokens:              10,
				OutputTokens:             5,
				CacheReadInputTokens:     1,
				CacheCreationInputTokens: 2,
				WebSearchRequests:        0,
				CostUSD:                  0.001,
				ContextWindow:            200000,
				MaxOutputTokens:          8192,
			},
		},
		PermissionDenials: []PermissionDenial{
			{
				ToolName:  "Bash",
				ToolUseID: "tool-1",
				ToolInput: map[string]any{"command": "rm -rf /"},
			},
		},
		Errors: []string{"permission denied"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded["uuid"] != "result-1" {
		t.Fatalf("expected uuid=result-1, got %v", decoded["uuid"])
	}
	if decoded["stop_reason"] != "end_turn" {
		t.Fatalf("expected stop_reason=end_turn, got %v", decoded["stop_reason"])
	}
	if _, ok := decoded["modelUsage"]; !ok {
		t.Fatal("expected modelUsage field")
	}
	if _, ok := decoded["permission_denials"]; !ok {
		t.Fatal("expected permission_denials field")
	}
	if _, ok := decoded["errors"]; !ok {
		t.Fatal("expected errors field")
	}
}
