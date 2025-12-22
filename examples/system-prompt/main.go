// Example: Configuring system prompts.
//
// System prompts control Claude's behavior and personality. You can:
// - Use a custom system prompt to define behavior
// - Use preset prompts like "claude_code"
// - Append additional instructions to the default prompt
//
// Usage:
//
//	go run examples/system-prompt/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("System Prompt Examples")
	fmt.Println("=====================")
	fmt.Println()

	// Example 1: Custom system prompt
	// Completely replaces the default system prompt
	fmt.Println("1. Custom System Prompt")
	fmt.Println("   Using WithSystemPrompt() to define custom behavior")
	customExample(ctx)

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Example 2: System prompt preset
	// Uses a predefined preset like "claude_code"
	fmt.Println("2. System Prompt Preset")
	fmt.Println("   Using WithSystemPromptPreset() with the claude_code preset")
	presetExample(ctx)

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Example 3: Append to system prompt
	// Adds additional instructions to the default prompt
	fmt.Println("3. Append to System Prompt")
	fmt.Println("   Using WithAppendSystemPrompt() to add context")
	appendExample(ctx)
}

// customExample demonstrates using a completely custom system prompt
func customExample(ctx context.Context) {
	// Define a custom system prompt
	// This replaces the default prompt entirely
	customPrompt := `You are a helpful assistant specialized in Go programming.
You should:
- Always provide idiomatic Go code
- Explain error handling patterns
- Suggest Go best practices
- Keep responses concise and practical`

	// Create client with custom system prompt
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPrompt(customPrompt),
		// Bypass permissions for example simplicity
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Ask a question to see the custom behavior
	if err := client.SendQuery("What's the best way to handle errors in Go?"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Println(m.Text())
		case *types.ResultMessage:
			return
		}
	}
}

// presetExample demonstrates using a system prompt preset
func presetExample(ctx context.Context) {
	// Use the claude_code preset
	// This preset includes instructions optimized for coding tasks
	preset := types.SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		// Optionally append additional instructions to the preset
		Append: stringPtr("Always include inline comments in code examples."),
	}

	// Create client with preset system prompt
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSystemPromptPreset(preset),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Ask a coding question
	if err := client.SendQuery("Write a simple HTTP server in Go"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Println(m.Text())
		case *types.ResultMessage:
			return
		}
	}
}

// appendExample demonstrates appending to the default system prompt
func appendExample(ctx context.Context) {
	// Append additional context to the default system prompt
	// This is useful when you want to add domain-specific context
	// without replacing the entire default prompt
	additionalContext := `Additional context: This is for a banking application.
Security and data privacy are critical.
Always consider edge cases for financial calculations.`

	// Create client with appended system prompt
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithAppendSystemPrompt(additionalContext),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Ask a question that will benefit from the additional context
	if err := client.SendQuery("How should I calculate interest on a savings account?"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Println(m.Text())
		case *types.ResultMessage:
			return
		}
	}
}

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}
