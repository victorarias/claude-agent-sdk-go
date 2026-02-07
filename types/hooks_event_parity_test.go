// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import "testing"

func TestHookEventConstants_AdditionalEvents(t *testing.T) {
	tests := map[HookEvent]string{
		HookSessionStart:  "SessionStart",
		HookSessionEnd:    "SessionEnd",
		HookSetup:         "Setup",
		HookTeammateIdle:  "TeammateIdle",
		HookTaskCompleted: "TaskCompleted",
	}

	for got, want := range tests {
		if string(got) != want {
			t.Fatalf("expected hook constant %s, got %s", want, got)
		}
	}
}

func TestToGenericCallback_MapInputConversion_NewHookTypes(t *testing.T) {
	called := false
	cb := ToGenericCallback(func(input *SetupHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		called = true
		if input.Trigger != "init" {
			t.Fatalf("expected trigger=init, got %s", input.Trigger)
		}
		return NewSetupOutput("ready"), nil
	})

	raw := map[string]any{
		"session_id":      "sess-1",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/tmp",
		"hook_event_name": "Setup",
		"trigger":         "init",
	}

	output, err := cb(raw, nil, &HookContext{})
	if err != nil {
		t.Fatalf("unexpected callback error: %v", err)
	}
	if !called {
		t.Fatal("expected callback to be invoked")
	}
	if output == nil || output.HookSpecific == nil {
		t.Fatal("expected hook-specific output")
	}
	if output.HookSpecific["hookEventName"] != "Setup" {
		t.Fatalf("expected hookEventName=Setup, got %v", output.HookSpecific["hookEventName"])
	}
}

func TestHookBuilder_WithAdditionalTypedCallbacks(t *testing.T) {
	builder := NewHookBuilder().
		ForEvent(HookTaskCompleted).
		WithCallback(func(input *TaskCompletedHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
			return &HookOutput{SystemMessage: "task complete"}, nil
		})

	matcher := builder.Build()
	if len(matcher.Hooks) != 1 {
		t.Fatalf("expected one callback, got %d", len(matcher.Hooks))
	}

	// Ensure typed callback can receive raw map payloads (as in control protocol).
	_, err := matcher.Hooks[0](map[string]any{
		"session_id":      "sess-1",
		"transcript_path": "/tmp/transcript.jsonl",
		"cwd":             "/tmp",
		"hook_event_name": "TaskCompleted",
		"task_id":         "task-1",
		"task_subject":    "Build feature",
	}, nil, &HookContext{})
	if err != nil {
		t.Fatalf("expected callback to accept map payload, got error: %v", err)
	}
}
