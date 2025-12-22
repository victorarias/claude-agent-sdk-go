package types

import (
	"testing"
)

// TestPreToolUseCallback tests type-safe PreToolUse callback.
func TestPreToolUseCallback(t *testing.T) {
	// RED: This test will fail because PreToolUseCallback doesn't exist yet

	var callbackInvoked bool
	var capturedInput *PreToolUseHookInput
	var capturedToolUseID *string

	callback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		callbackInvoked = true
		capturedInput = input
		capturedToolUseID = toolUseID
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	// Create test input
	input := &PreToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "test-session",
			TranscriptPath: "/path/to/transcript",
			Cwd:            "/working/dir",
			PermissionMode: "default",
			HookEventName:  "PreToolUse",
		},
		ToolName:  "Bash",
		ToolInput: map[string]any{"command": "echo hello"},
	}

	toolUseID := "test-tool-id"
	ctx := &HookContext{}

	// Invoke the callback
	output, err := callback(input, &toolUseID, ctx)

	if err != nil {
		t.Fatalf("callback returned error: %v", err)
	}

	if !callbackInvoked {
		t.Error("callback was not invoked")
	}

	if capturedInput != input {
		t.Error("input was not captured correctly")
	}

	if capturedToolUseID != &toolUseID {
		t.Error("toolUseID was not captured correctly")
	}

	if output == nil || output.Continue == nil || !*output.Continue {
		t.Error("expected Continue=true in output")
	}

	// Test that we can access typed fields without type assertions
	if capturedInput.ToolName != "Bash" {
		t.Errorf("expected ToolName=Bash, got %s", capturedInput.ToolName)
	}

	if capturedInput.ToolInput["command"] != "echo hello" {
		t.Errorf("expected command=echo hello, got %v", capturedInput.ToolInput["command"])
	}
}

// TestPostToolUseCallback tests type-safe PostToolUse callback.
func TestPostToolUseCallback(t *testing.T) {
	// RED: This test will fail because PostToolUseCallback doesn't exist yet

	var callbackInvoked bool
	var capturedInput *PostToolUseHookInput

	callback := func(input *PostToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		callbackInvoked = true
		capturedInput = input
		return &HookOutput{SuppressOutput: true}, nil
	}

	input := &PostToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:      "test-session",
			TranscriptPath: "/path/to/transcript",
			Cwd:            "/working/dir",
			HookEventName:  "PostToolUse",
		},
		ToolName:     "Read",
		ToolInput:    map[string]any{"file_path": "/test.txt"},
		ToolResponse: "file contents here",
	}

	toolUseID := "tool-123"
	ctx := &HookContext{}

	output, err := callback(input, &toolUseID, ctx)

	if err != nil {
		t.Fatalf("callback returned error: %v", err)
	}

	if !callbackInvoked {
		t.Error("callback was not invoked")
	}

	// Test typed field access
	if capturedInput.ToolName != "Read" {
		t.Errorf("expected ToolName=Read, got %s", capturedInput.ToolName)
	}

	if capturedInput.ToolResponse != "file contents here" {
		t.Errorf("expected ToolResponse='file contents here', got %v", capturedInput.ToolResponse)
	}

	if !output.SuppressOutput {
		t.Error("expected SuppressOutput=true")
	}
}

// TestUserPromptSubmitCallback tests type-safe UserPromptSubmit callback.
func TestUserPromptSubmitCallback(t *testing.T) {
	// RED: This test will fail because UserPromptSubmitCallback doesn't exist yet

	var capturedPrompt string

	callback := func(input *UserPromptSubmitHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		capturedPrompt = input.Prompt
		return &HookOutput{Continue: boolPtr(false)}, nil
	}

	input := &UserPromptSubmitHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:     "test-session",
			HookEventName: "UserPromptSubmit",
		},
		Prompt: "test user prompt",
	}

	output, err := callback(input, nil, &HookContext{})

	if err != nil {
		t.Fatalf("callback returned error: %v", err)
	}

	if capturedPrompt != "test user prompt" {
		t.Errorf("expected prompt='test user prompt', got %s", capturedPrompt)
	}

	if output.Continue == nil || *output.Continue {
		t.Error("expected Continue=false")
	}
}

// TestStopHookCallback tests type-safe Stop callback.
func TestStopHookCallback(t *testing.T) {
	// RED: This test will fail because StopHookCallback doesn't exist yet

	callback := func(input *StopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		if input.StopHookActive {
			return &HookOutput{StopReason: "custom stop"}, nil
		}
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	input := &StopHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:     "test-session",
			HookEventName: "Stop",
		},
		StopHookActive: true,
	}

	output, err := callback(input, nil, &HookContext{})

	if err != nil {
		t.Fatalf("callback returned error: %v", err)
	}

	if output.StopReason != "custom stop" {
		t.Errorf("expected StopReason='custom stop', got %s", output.StopReason)
	}
}

// TestSubagentStopCallback tests type-safe SubagentStop callback.
func TestSubagentStopCallback(t *testing.T) {
	// RED: This test will fail because SubagentStopCallback doesn't exist yet

	callback := func(input *SubagentStopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{
			SystemMessage: "subagent stopped",
		}, nil
	}

	input := &SubagentStopHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:     "test-session",
			HookEventName: "SubagentStop",
		},
		StopHookActive: false,
	}

	output, err := callback(input, nil, &HookContext{})

	if err != nil {
		t.Fatalf("callback returned error: %v", err)
	}

	if output.SystemMessage != "subagent stopped" {
		t.Errorf("expected SystemMessage='subagent stopped', got %s", output.SystemMessage)
	}
}

// TestPreCompactCallback tests type-safe PreCompact callback.
func TestPreCompactCallback(t *testing.T) {
	// RED: This test will fail because PreCompactCallback doesn't exist yet

	callback := func(input *PreCompactHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		customInstructions := "keep all code blocks"
		return &HookOutput{
			HookSpecific: map[string]any{
				"customInstructions": customInstructions,
			},
		}, nil
	}

	trigger := "token_limit"
	input := &PreCompactHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:     "test-session",
			HookEventName: "PreCompact",
		},
		Trigger:            trigger,
		CustomInstructions: nil,
	}

	output, err := callback(input, nil, &HookContext{})

	if err != nil {
		t.Fatalf("callback returned error: %v", err)
	}

	if input.Trigger != "token_limit" {
		t.Errorf("expected Trigger='token_limit', got %s", input.Trigger)
	}

	if output.HookSpecific == nil {
		t.Error("expected HookSpecific to be set")
	}
}

// TestTypeSafeCallbackConversionToGeneric tests that type-safe callbacks
// can be converted to the generic HookCallback type for backward compatibility.
func TestTypeSafeCallbackConversionToGeneric(t *testing.T) {
	// RED: This test will fail because conversion helpers don't exist yet

	preToolUseCallback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	// Convert to generic callback
	genericCallback := ToGenericCallback(preToolUseCallback)

	// Invoke with 'any' type (how it's called internally)
	var input any = &PreToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:     "test",
			HookEventName: "PreToolUse",
		},
		ToolName: "Bash",
	}

	output, err := genericCallback(input, nil, &HookContext{})

	if err != nil {
		t.Fatalf("converted callback returned error: %v", err)
	}

	if output == nil || output.Continue == nil || !*output.Continue {
		t.Error("expected Continue=true from converted callback")
	}
}

// TestTypeSafeCallbackErrorHandling tests that type-safe callbacks
// handle type mismatches appropriately.
func TestTypeSafeCallbackErrorHandling(t *testing.T) {
	// RED: This test will fail because error handling doesn't exist yet

	preToolUseCallback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	// Convert to generic
	genericCallback := ToGenericCallback(preToolUseCallback)

	// Try to invoke with wrong input type
	wrongInput := &PostToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:     "test",
			HookEventName: "PostToolUse",
		},
		ToolName: "Read",
	}

	_, err := genericCallback(wrongInput, nil, &HookContext{})

	// Should get a type error
	if err == nil {
		t.Error("expected error when calling with wrong input type")
	}
}
