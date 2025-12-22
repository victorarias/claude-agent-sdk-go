// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"testing"
)

// TestHookBuilderPreToolUse tests HookBuilder for PreToolUse hooks.
func TestHookBuilderPreToolUse(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	var callbackInvoked bool

	callback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		callbackInvoked = true
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	// Use HookBuilder pattern
	hookMatcher := NewHookBuilder().
		ForEvent(HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(callback).
		WithTimeout(5.0).
		Build()

	if hookMatcher.Matcher["tool_name"] != "Bash" {
		t.Errorf("expected matcher tool_name=Bash, got %v", hookMatcher.Matcher["tool_name"])
	}

	if hookMatcher.Timeout == nil || *hookMatcher.Timeout != 5.0 {
		t.Errorf("expected timeout=5.0, got %v", hookMatcher.Timeout)
	}

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}

	// Test the callback
	input := &PreToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID:     "test",
			HookEventName: "PreToolUse",
		},
		ToolName: "Bash",
	}

	_, err := hookMatcher.Hooks[0](input, nil, &HookContext{})
	if err != nil {
		t.Fatalf("callback error: %v", err)
	}

	if !callbackInvoked {
		t.Error("callback was not invoked")
	}
}

// TestHookBuilderPostToolUse tests HookBuilder for PostToolUse hooks.
func TestHookBuilderPostToolUse(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback := func(input *PostToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{SuppressOutput: true}, nil
	}

	hookMatcher := NewHookBuilder().
		ForEvent(HookPostToolUse).
		WithMatcher(map[string]any{"tool_name": "Read"}).
		WithCallback(callback).
		Build()

	if hookMatcher.Matcher["tool_name"] != "Read" {
		t.Errorf("expected matcher tool_name=Read, got %v", hookMatcher.Matcher["tool_name"])
	}

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderMatchAll tests HookBuilder with MatchAll (nil matcher).
func TestHookBuilderMatchAll(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	hookMatcher := NewHookBuilder().
		ForEvent(HookPreToolUse).
		MatchAll().
		WithCallback(callback).
		Build()

	if hookMatcher.Matcher != nil {
		t.Errorf("expected nil matcher for MatchAll, got %v", hookMatcher.Matcher)
	}

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderMultipleCallbacks tests HookBuilder with multiple callbacks.
func TestHookBuilderMultipleCallbacks(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback1 := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	callback2 := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	hookMatcher := NewHookBuilder().
		ForEvent(HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(callback1).
		WithCallback(callback2).
		Build()

	if len(hookMatcher.Hooks) != 2 {
		t.Fatalf("expected 2 hooks, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderUserPromptSubmit tests HookBuilder for UserPromptSubmit hooks.
func TestHookBuilderUserPromptSubmit(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback := func(input *UserPromptSubmitHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	// UserPromptSubmit doesn't need a matcher
	hookMatcher := NewHookBuilder().
		ForEvent(HookUserPromptSubmit).
		WithCallback(callback).
		Build()

	if hookMatcher.Matcher != nil {
		t.Errorf("expected nil matcher for UserPromptSubmit, got %v", hookMatcher.Matcher)
	}

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderStop tests HookBuilder for Stop hooks.
func TestHookBuilderStop(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback := func(input *StopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{StopReason: "custom"}, nil
	}

	hookMatcher := NewHookBuilder().
		ForEvent(HookStop).
		WithCallback(callback).
		Build()

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderSubagentStop tests HookBuilder for SubagentStop hooks.
func TestHookBuilderSubagentStop(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback := func(input *SubagentStopHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{SystemMessage: "stopped"}, nil
	}

	hookMatcher := NewHookBuilder().
		ForEvent(HookSubagentStop).
		WithCallback(callback).
		Build()

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderPreCompact tests HookBuilder for PreCompact hooks.
func TestHookBuilderPreCompact(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback := func(input *PreCompactHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	hookMatcher := NewHookBuilder().
		ForEvent(HookPreCompact).
		WithCallback(callback).
		Build()

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderWithGenericCallback tests HookBuilder with generic callback.
func TestHookBuilderWithGenericCallback(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	genericCallback := func(input any, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	hookMatcher := NewHookBuilder().
		ForEvent(HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithGenericCallback(genericCallback).
		Build()

	if len(hookMatcher.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(hookMatcher.Hooks))
	}
}

// TestHookBuilderFluentAPI tests that the HookBuilder API is fluent.
func TestHookBuilderFluentAPI(t *testing.T) {
	// RED: This test will fail because HookBuilder doesn't exist yet

	callback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	// All methods should return the builder for chaining
	builder := NewHookBuilder()
	result1 := builder.ForEvent(HookPreToolUse)
	result2 := result1.WithMatcher(map[string]any{"tool_name": "Bash"})
	result3 := result2.WithCallback(callback)
	result4 := result3.WithTimeout(10.0)

	// Should be the same builder instance
	if builder != result1 || builder != result2 || builder != result3 || builder != result4 {
		t.Error("HookBuilder methods should return the same builder instance for chaining")
	}

	hookMatcher := result4.Build()

	if len(hookMatcher.Hooks) == 0 {
		t.Fatal("Build should return a HookMatcher with hooks")
	}
}

// TestHookBuilderToOption tests converting HookBuilder to Option.
func TestHookBuilderToOption(t *testing.T) {
	// RED: This test will fail because ToOption doesn't exist yet

	callback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	builder := NewHookBuilder().
		ForEvent(HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(callback)

	// Convert to option
	option := builder.ToOption()

	// Apply to options
	opts := DefaultOptions()
	option(opts)

	// Verify hook was registered
	if opts.Hooks == nil {
		t.Fatal("Hooks map should not be nil")
	}

	matchers, ok := opts.Hooks[HookPreToolUse]
	if !ok {
		t.Fatal("PreToolUse hooks should be registered")
	}

	if len(matchers) != 1 {
		t.Fatalf("expected 1 matcher, got %d", len(matchers))
	}

	if matchers[0].Matcher["tool_name"] != "Bash" {
		t.Errorf("expected tool_name=Bash, got %v", matchers[0].Matcher["tool_name"])
	}
}

// TestHookBuilderBuildForOptions tests BuildForOptions convenience method.
func TestHookBuilderBuildForOptions(t *testing.T) {
	// RED: This test will fail because BuildForOptions doesn't exist yet

	callback := func(input *PreToolUseHookInput, toolUseID *string, ctx *HookContext) (*HookOutput, error) {
		return &HookOutput{Continue: boolPtr(true)}, nil
	}

	opts := DefaultOptions()

	// Build directly into options
	NewHookBuilder().
		ForEvent(HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(callback).
		BuildForOptions(opts)

	// Verify hook was registered
	if opts.Hooks == nil {
		t.Fatal("Hooks map should not be nil")
	}

	matchers, ok := opts.Hooks[HookPreToolUse]
	if !ok {
		t.Fatal("PreToolUse hooks should be registered")
	}

	if len(matchers) != 1 {
		t.Fatalf("expected 1 matcher, got %d", len(matchers))
	}
}
