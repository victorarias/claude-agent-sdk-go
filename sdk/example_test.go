// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk_test

import (
	"fmt"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

// ExampleNewClient demonstrates creating a client with options.
func ExampleNewClient() {
	// Create a client with custom configuration
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-20250514"),
		types.WithMaxTurns(5),
	)

	// Check the configured options
	opts := client.Options()
	fmt.Println("Model:", opts.Model)
	fmt.Println("MaxTurns:", opts.MaxTurns)

	// Output:
	// Model: claude-sonnet-4-20250514
	// MaxTurns: 5
}

// ExampleNewClient_withSystemPrompt shows how to set a system prompt.
func ExampleNewClient_withSystemPrompt() {
	client := sdk.NewClient(
		types.WithSystemPrompt("You are a helpful coding assistant."),
	)

	opts := client.Options()
	// SystemPrompt can be string or []ContentBlock
	if sp, ok := opts.SystemPrompt.(string); ok {
		fmt.Println("System prompt set:", sp != "")
	}

	// Output:
	// System prompt set: true
}

// ExampleNewClient_withWorkingDirectory shows how to set the working directory.
func ExampleNewClient_withWorkingDirectory() {
	client := sdk.NewClient(
		types.WithCwd("/tmp"),
	)

	opts := client.Options()
	fmt.Println("Working directory:", opts.Cwd)

	// Output:
	// Working directory: /tmp
}

// ExampleNewClient_withPermissionMode shows how to configure permissions.
func ExampleNewClient_withPermissionMode() {
	// Default mode - prompts for permission on tool use
	clientDefault := sdk.NewClient(
		types.WithPermissionMode(types.PermissionDefault),
	)

	// AcceptEdits mode - auto-approves file edits
	clientAcceptEdits := sdk.NewClient(
		types.WithPermissionMode(types.PermissionAcceptEdits),
	)

	// Bypass mode - auto-approves everything (use with caution)
	clientBypass := sdk.NewClient(
		types.WithPermissionMode(types.PermissionBypass),
	)

	fmt.Println("Default mode:", clientDefault.Options().PermissionMode)
	fmt.Println("AcceptEdits mode:", clientAcceptEdits.Options().PermissionMode)
	fmt.Println("Bypass mode:", clientBypass.Options().PermissionMode)

	// Output:
	// Default mode: default
	// AcceptEdits mode: acceptEdits
	// Bypass mode: bypassPermissions
}

// ExampleWithToolUseHook demonstrates registering a tool use hook.
func ExampleWithToolUseHook() {
	toolCalls := 0

	client := sdk.NewClient(
		sdk.WithToolUseHook(func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			toolCalls++
			// Allow all tool calls
			return &types.HookOutput{
				Decision: types.PermissionAllow{},
			}, nil
		}),
	)

	// Verify hook is registered
	fmt.Println("Client created with hook:", client != nil)

	// Output:
	// Client created with hook: true
}

// ExampleWithStderrCallback demonstrates capturing stderr output.
func ExampleWithStderrCallback() {
	var stderrOutput string

	client := sdk.NewClient(
		sdk.WithStderrCallback(func(line string) {
			stderrOutput += line + "\n"
		}),
	)

	// Verify client is created
	fmt.Println("Client created with stderr callback:", client != nil)

	// Output:
	// Client created with stderr callback: true
}

// ExampleOptions demonstrates creating options with multiple settings.
func ExampleOptions() {
	opts := types.DefaultOptions()
	types.WithModel("claude-sonnet-4-20250514")(opts)
	types.WithMaxTurns(10)(opts)
	types.WithCwd("/project")(opts)
	types.WithAllowedTools("Read", "Write", "Bash")(opts)

	fmt.Println("Model:", opts.Model)
	fmt.Println("MaxTurns:", opts.MaxTurns)
	fmt.Println("Cwd:", opts.Cwd)
	fmt.Println("AllowedTools count:", len(opts.AllowedTools))

	// Output:
	// Model: claude-sonnet-4-20250514
	// MaxTurns: 10
	// Cwd: /project
	// AllowedTools count: 3
}

// Example_contentBlocks shows how to work with message content blocks.
func Example_contentBlocks() {
	// Create content blocks programmatically
	textBlock := &types.TextBlock{TextContent: "Hello, world!"}
	fmt.Println("Text block type:", textBlock.Type())
	fmt.Println("Text content:", textBlock.Text())

	thinkingBlock := &types.ThinkingBlock{ThinkingContent: "Let me think about this..."}
	fmt.Println("Thinking block type:", thinkingBlock.Type())
	fmt.Println("Thinking content:", thinkingBlock.Thinking())

	toolUseBlock := &types.ToolUseBlock{
		ID:        "tool_123",
		Name:      "calculator",
		ToolInput: map[string]any{"expression": "2+2"},
	}
	fmt.Println("Tool use block type:", toolUseBlock.Type())
	fmt.Println("Tool name:", toolUseBlock.Name)

	// Output:
	// Text block type: text
	// Text content: Hello, world!
	// Thinking block type: thinking
	// Thinking content: Let me think about this...
	// Tool use block type: tool_use
	// Tool name: calculator
}

// Example_errors shows how to handle SDK errors.
func Example_errors() {
	// Check error types
	timeoutErr := &types.TimeoutError{Operation: "connect", Duration: 30 * time.Second}
	fmt.Println("Timeout error:", timeoutErr.Error())

	closedErr := &types.ClosedError{Resource: "session"}
	fmt.Println("Closed error:", closedErr.Error())

	processErr := &types.ProcessError{ExitCode: 1, Stderr: "command failed"}
	fmt.Println("Process error:", processErr.Error())

	// Output:
	// Timeout error: operation 'connect' timed out after 30s
	// Closed error: resource closed: session
	// Process error: process exited with code 1: command failed
}

// Example_mcpServers shows how to configure MCP servers.
func Example_mcpServers() {
	opts := types.DefaultOptions()
	types.WithMCPServers(map[string]types.MCPServerConfig{
		"filesystem": {
			Command: "npx",
			Args:    []string{"-y", "@anthropic/mcp-filesystem"},
			Env:     map[string]string{"HOME": "/home/user"},
		},
	})(opts)

	fmt.Println("MCP servers configured:", len(opts.MCPServers))
	if server, ok := opts.MCPServers["filesystem"]; ok {
		fmt.Println("Filesystem server command:", server.Command)
	}

	// Output:
	// MCP servers configured: 1
	// Filesystem server command: npx
}
