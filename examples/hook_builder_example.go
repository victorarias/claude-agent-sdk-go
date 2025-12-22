package main

import (
	"context"
	"fmt"
	"log"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

// This example demonstrates the improved hook registration API with:
// 1. Type-safe callbacks (no type assertions needed)
// 2. HookBuilder pattern (unified, discoverable API)

func main() {
	// Example 1: Type-safe PreToolUse hook with HookBuilder
	preToolUseHook := types.NewHookBuilder().
		ForEvent(types.HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// No type assertion needed! Input is already typed as *PreToolUseHookInput
			fmt.Printf("Tool about to be called: %s\n", input.ToolName)
			fmt.Printf("Tool input: %v\n", input.ToolInput)

			// You can inspect or modify the tool call
			if command, ok := input.ToolInput["command"].(string); ok {
				fmt.Printf("Bash command: %s\n", command)
			}

			// Allow the tool to proceed
			return &types.HookOutput{
				Continue: boolPtr(true),
			}, nil
		}).
		WithTimeout(5.0). // Optional timeout
		ToOption()

	// Example 2: Type-safe PostToolUse hook
	postToolUseHook := types.NewHookBuilder().
		ForEvent(types.HookPostToolUse).
		MatchAll(). // Match all tools, not just specific ones
		WithCallback(func(input *types.PostToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// No type assertion needed! Input is already typed as *PostToolUseHookInput
			fmt.Printf("Tool completed: %s\n", input.ToolName)
			fmt.Printf("Tool response: %v\n", input.ToolResponse)

			// You can inspect or modify the response
			return &types.HookOutput{
				SuppressOutput: false, // Don't suppress the output
			}, nil
		}).
		ToOption()

	// Example 3: Type-safe UserPromptSubmit hook
	userPromptHook := types.NewHookBuilder().
		ForEvent(types.HookUserPromptSubmit).
		WithCallback(func(input *types.UserPromptSubmitHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// No type assertion needed! Input is already typed as *UserPromptSubmitHookInput
			fmt.Printf("User prompt: %s\n", input.Prompt)

			// You can validate or modify the prompt
			return &types.HookOutput{
				Continue: boolPtr(true),
			}, nil
		}).
		ToOption()

	// Example 4: Multiple callbacks on the same hook
	multiCallbackHook := types.NewHookBuilder().
		ForEvent(types.HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Read"}).
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			fmt.Println("First callback: Logging read operation")
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			fmt.Println("Second callback: Validating file path")
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		ToOption()

	// Create client with all hooks
	client := sdk.NewClient(
		preToolUseHook,
		postToolUseHook,
		userPromptHook,
		multiCallbackHook,
		types.WithPermissionMode(types.PermissionBypass),
	)

	ctx := context.Background()

	// Connect and use the client
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("Client connected successfully with type-safe hooks!")
}

// Comparison with old API (for reference):

func oldAPIExample() {
	// OLD WAY: Required type assertions
	oldWayHook := sdk.WithPreToolUseHook(
		map[string]any{"tool_name": "Bash"},
		func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// Required type assertion - error-prone!
			typedInput, ok := input.(*types.PreToolUseHookInput)
			if !ok {
				return nil, fmt.Errorf("invalid input type: %T", input)
			}

			fmt.Printf("Tool: %s\n", typedInput.ToolName)
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		},
	)

	// NEW WAY: No type assertions, cleaner API
	newWayHook := types.NewHookBuilder().
		ForEvent(types.HookPreToolUse).
		WithMatcher(map[string]any{"tool_name": "Bash"}).
		WithCallback(func(input *types.PreToolUseHookInput, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// No type assertion needed!
			fmt.Printf("Tool: %s\n", input.ToolName)
			return &types.HookOutput{Continue: boolPtr(true)}, nil
		}).
		ToOption()

	_ = oldWayHook
	_ = newWayHook
}

func boolPtr(b bool) *bool {
	return &b
}
