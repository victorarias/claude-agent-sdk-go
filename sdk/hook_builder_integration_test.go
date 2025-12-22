package sdk

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestHookBuilderIntegration demonstrates using HookBuilder with Client.
func TestHookBuilderIntegration(t *testing.T) {
	var preHookCalled bool
	var postHookCalled bool

	// Create hooks using the new builder pattern
	preHook := types.NewHookBuilder().
		ForEvent(types.HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			preHookCalled = true
			// Type-safe access to input fields, no type assertion needed!
			if input.ToolName != "Bash" {
				t.Errorf("expected tool_name=Bash, got %s", input.ToolName)
			}
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		WithTimeout(5.0).
		ToOption()

	postHook := types.NewHookBuilder().
		ForEvent(types.HookPostToolUse).
		WithMatcher(map[string]any{"tool_name": "Read"}).
		WithCallback(func(input *types.PostToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			postHookCalled = true
			// Type-safe access without assertions
			if input.ToolName != "Read" {
				t.Errorf("expected tool_name=Read, got %s", input.ToolName)
			}
			return &types.HookOutput{SuppressOutput: false}, nil
		}).
		ToOption()

	// Create client with hooks
	client := NewClient(preHook, postHook)

	// Verify hooks were registered
	if client.hooks == nil {
		t.Fatal("hooks should be registered")
	}

	if len(client.hooks[types.HookPreToolUse]) != 1 {
		t.Errorf("expected 1 PreToolUse hook, got %d", len(client.hooks[types.HookPreToolUse]))
	}

	if len(client.hooks[types.HookPostToolUse]) != 1 {
		t.Errorf("expected 1 PostToolUse hook, got %d", len(client.hooks[types.HookPostToolUse]))
	}

	// Test that the matcher was set correctly
	preToolUseHook := client.hooks[types.HookPreToolUse][0]
	if preToolUseHook.Matcher["tool_name"] != "Bash" {
		t.Errorf("expected tool_name=Bash, got %v", preToolUseHook.Matcher["tool_name"])
	}

	if preToolUseHook.Timeout == nil || *preToolUseHook.Timeout != 5.0 {
		t.Errorf("expected timeout=5.0, got %v", preToolUseHook.Timeout)
	}

	// Test the callbacks work
	input := &types.PreToolUseHookInput{
		BaseHookInput: types.BaseHookInput{
			SessionID:     "test",
			HookEventName: "PreToolUse",
		},
		ToolName: "Bash",
	}

	_, err := preToolUseHook.Hooks[0](input, nil, &types.HookContext{})
	if err != nil {
		t.Fatalf("pre hook error: %v", err)
	}

	if !preHookCalled {
		t.Error("pre hook was not called")
	}

	postInput := &types.PostToolUseHookInput{
		BaseHookInput: types.BaseHookInput{
			SessionID:     "test",
			HookEventName: "PostToolUse",
		},
		ToolName:     "Read",
		ToolResponse: "file contents",
	}

	postToolUseHook := client.hooks[types.HookPostToolUse][0]
	_, err = postToolUseHook.Hooks[0](postInput, nil, &types.HookContext{})
	if err != nil {
		t.Fatalf("post hook error: %v", err)
	}

	if !postHookCalled {
		t.Error("post hook was not called")
	}
}

// TestHookBuilderVsOldAPI demonstrates backward compatibility.
func TestHookBuilderVsOldAPI(t *testing.T) {
	// Old API (still works)
	oldAPIClient := NewClient(
		WithPreToolUseHook(map[string]any{"tool_name": "Bash"}, func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// Requires type assertion (the old way)
			typedInput, ok := input.(*types.PreToolUseHookInput)
			if !ok {
				t.Errorf("expected PreToolUseHookInput, got %T", input)
			}
			if typedInput.ToolName != "Bash" {
				t.Errorf("expected Bash, got %s", typedInput.ToolName)
			}
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}),
	)

	// New API (cleaner, type-safe)
	newAPIClient := NewClient(
		types.NewHookBuilder().
			ForEvent(types.HookPreToolUse).
			WithMatcher(map[string]any{"tool_name": "Bash"}).
			WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
				// No type assertion needed!
				if input.ToolName != "Bash" {
					t.Errorf("expected Bash, got %s", input.ToolName)
				}
				return &types.HookOutput{Continue: boolPtr(true)}, nil
			}).
			ToOption(),
	)

	// Both clients should have the hook registered
	if len(oldAPIClient.hooks[types.HookPreToolUse]) != 1 {
		t.Error("old API client should have PreToolUse hook")
	}

	if len(newAPIClient.hooks[types.HookPreToolUse]) != 1 {
		t.Error("new API client should have PreToolUse hook")
	}
}

// TestHookBuilderMatchAll demonstrates the MatchAll feature.
func TestHookBuilderMatchAll(t *testing.T) {
	var callCount int

	hook := types.NewHookBuilder().
		ForEvent(types.HookPreToolUse).
		MatchAll().
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			callCount++
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		ToOption()

	client := NewClient(hook)

	// Verify matcher is nil (matches all)
	if client.hooks[types.HookPreToolUse][0].Matcher != nil {
		t.Error("MatchAll should result in nil matcher")
	}

	// Test with different tool names - all should match
	for _, toolName := range []string{"Bash", "Read", "Write", "Edit"} {
		input := &types.PreToolUseHookInput{
			BaseHookInput: types.BaseHookInput{
				SessionID:     "test",
				HookEventName: "PreToolUse",
			},
			ToolName: toolName,
		}

		_, err := client.hooks[types.HookPreToolUse][0].Hooks[0](input, nil, &types.HookContext{})
		if err != nil {
			t.Fatalf("hook error for %s: %v", toolName, err)
		}
	}

	if callCount != 4 {
		t.Errorf("expected 4 calls, got %d", callCount)
	}
}

// TestHookBuilderMultipleCallbacks demonstrates adding multiple callbacks.
func TestHookBuilderMultipleCallbacks(t *testing.T) {
	var callback1Called bool
	var callback2Called bool

	hook := types.NewHookBuilder().
		ForEvent(types.HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			callback1Called = true
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			callback2Called = true
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		ToOption()

	client := NewClient(hook)

	if len(client.hooks[types.HookPreToolUse][0].Hooks) != 2 {
		t.Errorf("expected 2 callbacks, got %d", len(client.hooks[types.HookPreToolUse][0].Hooks))
	}

	// Call both callbacks
	input := &types.PreToolUseHookInput{
		BaseHookInput: types.BaseHookInput{
			SessionID:     "test",
			HookEventName: "PreToolUse",
		},
		ToolName: "Bash",
	}

	for i, callback := range client.hooks[types.HookPreToolUse][0].Hooks {
		_, err := callback(input, nil, &types.HookContext{})
		if err != nil {
			t.Fatalf("callback %d error: %v", i, err)
		}
	}

	if !callback1Called {
		t.Error("callback1 was not called")
	}

	if !callback2Called {
		t.Error("callback2 was not called")
	}
}

// TestHookBuilderUserPromptSubmit tests user prompt hooks which don't use matchers.
func TestHookBuilderUserPromptSubmit(t *testing.T) {
	var capturedPrompt string

	hook := types.NewHookBuilder().
		ForEvent(types.HookUserPromptSubmit).
		WithCallback(func(input *types.UserPromptSubmitHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			capturedPrompt = input.Prompt
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		ToOption()

	client := NewClient(hook)

	// Verify no matcher is set (not applicable for UserPromptSubmit)
	if client.hooks[types.HookUserPromptSubmit][0].Matcher != nil {
		t.Error("UserPromptSubmit should not have a matcher")
	}

	// Test the callback
	input := &types.UserPromptSubmitHookInput{
		BaseHookInput: types.BaseHookInput{
			SessionID:     "test",
			HookEventName: "UserPromptSubmit",
		},
		Prompt: "Hello, Claude!",
	}

	_, err := client.hooks[types.HookUserPromptSubmit][0].Hooks[0](input, nil, &types.HookContext{})
	if err != nil {
		t.Fatalf("hook error: %v", err)
	}

	if capturedPrompt != "Hello, Claude!" {
		t.Errorf("expected prompt='Hello, Claude!', got %s", capturedPrompt)
	}
}
