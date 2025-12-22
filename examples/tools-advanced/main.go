// Example: Advanced tools configuration patterns.
//
// This example demonstrates advanced patterns for configuring tools:
// - Using ToolsPreset with Type and Preset fields
// - Combining tools configuration with permission modes
// - Using sandbox configuration with tools
// - Advanced permission patterns for specific tool configurations
//
// Usage:
//
//	go run examples/tools-advanced/main.go
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

	// Run all examples
	if err := toolsPresetExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in tools preset example: %v\n", err)
	}

	if err := sandboxWithToolsExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in sandbox with tools example: %v\n", err)
	}

	if err := permissionModesExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in permission modes example: %v\n", err)
	}

	if err := customToolPermissionsExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in custom tool permissions example: %v\n", err)
	}
}

// toolsPresetExample demonstrates using ToolsPreset configuration.
func toolsPresetExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 1: Tools Preset")
	fmt.Println("===================================")
	fmt.Println("Using ToolsPreset to configure tools with a preset configuration.")
	fmt.Println("This uses the 'claude_code' preset which provides the standard set")
	fmt.Println("of tools available in Claude Code.")
	fmt.Println()

	// Use a tools preset instead of listing individual tools
	// This is equivalent to the default tools in Claude Code
	messages, err := sdk.RunQuery(
		ctx,
		"What tools do you have available? Just list the categories briefly.",
		types.WithToolsPreset(types.ToolsPreset{
			Type:   "preset",
			Preset: "claude_code",
		}),
		types.WithMaxTurns(1),
	)
	if err != nil {
		return err
	}

	printMessages(messages)
	fmt.Println()
	return nil
}

// sandboxWithToolsExample demonstrates using sandbox configuration with tools.
func sandboxWithToolsExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 2: Sandbox with Tools")
	fmt.Println("===================================")
	fmt.Println("Configuring tools with sandbox restrictions.")
	fmt.Println("This example enables sandbox mode to isolate command execution.")
	fmt.Println()

	// Configure sandbox with specific tools
	// The sandbox provides network isolation and file access control
	messages, err := sdk.RunQuery(
		ctx,
		"Run 'echo Hello from sandbox' and tell me if you can access the network.",
		types.WithTools("Bash", "Read", "Glob"),
		types.WithAllowedTools("Bash", "Read", "Glob"),
		types.WithSandbox(types.SandboxSettings{
			Enabled:                  true,
			AutoAllowBashIfSandboxed: true, // Auto-allow Bash since it's sandboxed
			Network: &types.SandboxNetworkConfig{
				AllowLocalBinding: false, // Prevent local network binding
			},
		}),
		types.WithMaxTurns(2),
	)
	if err != nil {
		return err
	}

	printMessages(messages)
	fmt.Println()
	return nil
}

// permissionModesExample demonstrates different permission modes with tools.
func permissionModesExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 3: Permission Modes")
	fmt.Println("===================================")
	fmt.Println("Using different permission modes to control tool execution.")
	fmt.Println()

	// Plan mode: Claude creates an execution plan before running tools
	fmt.Println("Using 'plan' permission mode...")
	fmt.Println("Claude will create a plan before executing any tools.")
	fmt.Println()

	messages, err := sdk.RunQuery(
		ctx,
		"List the files in the current directory.",
		types.WithPermissionMode(types.PermissionPlan),
		types.WithTools("Bash", "Read", "Glob"),
		types.WithMaxTurns(2),
	)
	if err != nil {
		return err
	}

	printMessages(messages)
	fmt.Println()

	// AcceptEdits mode: Auto-approve edit operations
	fmt.Println("\n---")
	fmt.Println("Using 'acceptEdits' permission mode...")
	fmt.Println("This would auto-approve Write and Edit operations.")
	fmt.Println("(Note: Not executing to avoid file modifications)")
	fmt.Println()

	return nil
}

// customToolPermissionsExample demonstrates advanced permission patterns.
func customToolPermissionsExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 4: Custom Tool Permissions")
	fmt.Println("===================================")
	fmt.Println("Using custom permission callbacks for fine-grained control.")
	fmt.Println()

	// Track tool usage
	toolUsage := make(map[string]int)

	// Create client with custom permission logic
	client := sdk.NewClient(
		types.WithTools("Bash", "Read", "Glob", "Grep"),
		types.WithMaxTurns(3),
		sdk.WithCanUseTool(func(toolName string, input map[string]any, permCtx *types.ToolPermissionContext) (types.PermissionResult, error) {
			// Track usage
			toolUsage[toolName]++

			// Allow read-only tools automatically
			if toolName == "Read" || toolName == "Glob" || toolName == "Grep" {
				fmt.Printf("[Auto-approved: %s (usage: %d)]\n", toolName, toolUsage[toolName])
				return &types.PermissionResultAllow{Behavior: "allow"}, nil
			}

			// For Bash, only allow safe commands
			if toolName == "Bash" {
				if cmd, ok := input["command"].(string); ok {
					// Check if command is safe (starts with safe commands)
					safeCommands := []string{"ls", "pwd", "echo", "date", "whoami"}
					for _, safe := range safeCommands {
						if len(cmd) >= len(safe) && cmd[:len(safe)] == safe {
							fmt.Printf("[Auto-approved: Bash - safe command '%s' (usage: %d)]\n", cmd, toolUsage[toolName])
							return &types.PermissionResultAllow{Behavior: "allow"}, nil
						}
					}

					// Deny unsafe commands
					fmt.Printf("[Denied: Bash - unsafe command '%s']\n", cmd)
					return &types.PermissionResultDeny{
						Behavior: "deny",
						Message:  fmt.Sprintf("Command not in safe list: %s", cmd),
					}, nil
				}
			}

			// Default: deny
			fmt.Printf("[Denied: %s - not in auto-approve list]\n", toolName)
			return &types.PermissionResultDeny{
				Behavior: "deny",
				Message:  fmt.Sprintf("Tool %s requires manual approval", toolName),
			}, nil
		}),
	)

	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()

	if err := client.SendQuery("List the current directory and tell me the current time."); err != nil {
		return err
	}

	// Receive and print response
	fmt.Println("\n---- Response ----")
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			return err
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
			fmt.Println()

			// Print tool usage statistics
			fmt.Println("\n---- Tool Usage Statistics ----")
			for tool, count := range toolUsage {
				fmt.Printf("%s: %d calls\n", tool, count)
			}
			fmt.Println()
			return nil
		}
	}
}

// printMessages is a helper to print messages from a query.
func printMessages(messages []types.Message) {
	for _, msg := range messages {
		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
		}
	}
}
