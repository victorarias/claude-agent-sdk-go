// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"encoding/json"
	"testing"
)

func TestNewPreToolUseOutputWithContext(t *testing.T) {
	output := NewPreToolUseOutputWithContext("allow", "ok", map[string]any{"k": "v"}, "extra")
	if output.HookSpecific["hookEventName"] != "PreToolUse" {
		t.Fatalf("unexpected hook event: %v", output.HookSpecific["hookEventName"])
	}
	if output.HookSpecific["additionalContext"] != "extra" {
		t.Fatalf("expected additionalContext=extra, got %v", output.HookSpecific["additionalContext"])
	}
}

func TestNewPostToolUseOutputWithUpdate(t *testing.T) {
	update := map[string]any{"content": "updated"}
	output := NewPostToolUseOutputWithUpdate("ctx", update)
	if output.HookSpecific["updatedMCPToolOutput"] == nil {
		t.Fatal("expected updatedMCPToolOutput to be set")
	}
}

func TestNewHookOutputsForNewEvents(t *testing.T) {
	tests := []struct {
		name     string
		output   *HookOutput
		event    string
		validate func(t *testing.T, hook map[string]any)
	}{
		{
			name:   "post tool use failure",
			output: NewPostToolUseFailureOutput("failed"),
			event:  "PostToolUseFailure",
			validate: func(t *testing.T, hook map[string]any) {
				if hook["additionalContext"] != "failed" {
					t.Fatalf("unexpected additionalContext: %v", hook["additionalContext"])
				}
			},
		},
		{
			name:   "notification",
			output: NewNotificationOutput("note"),
			event:  "Notification",
			validate: func(t *testing.T, hook map[string]any) {
				if hook["additionalContext"] != "note" {
					t.Fatalf("unexpected additionalContext: %v", hook["additionalContext"])
				}
			},
		},
		{
			name:   "subagent start",
			output: NewSubagentStartOutput("start"),
			event:  "SubagentStart",
			validate: func(t *testing.T, hook map[string]any) {
				if hook["additionalContext"] != "start" {
					t.Fatalf("unexpected additionalContext: %v", hook["additionalContext"])
				}
			},
		},
		{
			name:   "permission request",
			output: NewPermissionRequestOutput(map[string]any{"allow": true}),
			event:  "PermissionRequest",
			validate: func(t *testing.T, hook map[string]any) {
				decision, ok := hook["decision"].(map[string]any)
				if !ok || decision["allow"] != true {
					t.Fatalf("unexpected decision: %v", hook["decision"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.output)
			if err != nil {
				t.Fatalf("marshal failed: %v", err)
			}
			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatalf("unmarshal failed: %v", err)
			}
			hook, ok := payload["hookSpecificOutput"].(map[string]any)
			if !ok {
				t.Fatalf("expected hookSpecificOutput map, got %T", payload["hookSpecificOutput"])
			}
			if hook["hookEventName"] != tt.event {
				t.Fatalf("expected hookEventName=%s, got %v", tt.event, hook["hookEventName"])
			}
			tt.validate(t, hook)
		})
	}
}

func TestHookBuilderSupportsNewTypedCallbacks(t *testing.T) {
	builder := NewHookBuilder().
		ForEvent(HookNotification).
		WithCallback(func(input *NotificationHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
			return NewNotificationOutput("ok"), nil
		})

	matcher := builder.Build()
	if len(matcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook callback, got %d", len(matcher.Hooks))
	}
}
