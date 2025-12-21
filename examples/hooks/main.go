// Example: Using hooks to monitor and control tool execution.
//
// Hooks allow you to:
// - Log all tool calls for auditing
// - Modify tool inputs before execution
// - Block certain operations
// - Track performance metrics
//
// Usage:
//
//	go run examples/hooks/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Track metrics
	var toolCallCount int
	toolDurations := make(map[string]time.Duration)

	// Create client with hooks
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),

		// PreToolUse hook: Called BEFORE each tool execution
		// Can modify inputs, block execution, or log the attempt
		sdk.WithPreToolUseHook(nil, func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// Type assert to get the pre-tool-use input
			preInput, ok := input.(*types.PreToolUseHookInput)
			if !ok {
				return nil, nil
			}

			toolCallCount++
			fmt.Printf("\n[PreToolUse #%d] %s\n", toolCallCount, preInput.ToolName)

			// Log input (truncate long values)
			for k, v := range preInput.ToolInput {
				str := fmt.Sprintf("%v", v)
				if len(str) > 50 {
					str = str[:47] + "..."
				}
				fmt.Printf("   %s: %s\n", k, str)
			}

			// Example: Block dangerous bash commands
			if preInput.ToolName == "Bash" {
				if cmd, ok := preInput.ToolInput["command"].(string); ok {
					dangerous := []string{"rm -rf", "sudo", "> /dev/", "mkfs"}
					for _, d := range dangerous {
						if strings.Contains(cmd, d) {
							fmt.Printf("   BLOCKED: dangerous command detected\n")
							return &types.HookOutput{
								Decision: "block",
								Reason:   fmt.Sprintf("Blocked: command contains '%s'", d),
							}, nil
						}
					}
				}
			}

			// Example: Block writes to certain paths
			if preInput.ToolName == "Write" || preInput.ToolName == "Edit" {
				if path, ok := preInput.ToolInput["file_path"].(string); ok {
					protected := []string{"/etc/", "/usr/", ".env", "credentials"}
					for _, p := range protected {
						if strings.Contains(path, p) {
							fmt.Printf("   BLOCKED: protected path\n")
							return &types.HookOutput{
								Decision: "block",
								Reason:   fmt.Sprintf("Cannot write to protected path: %s", path),
							}, nil
						}
					}
				}
			}

			fmt.Printf("   Allowed\n")
			return nil, nil
		}),

		// PostToolUse hook: Called AFTER each tool execution
		// Can log results, track metrics, or trigger side effects
		sdk.WithPostToolUseHook(nil, func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			// Type assert to get the post-tool-use input
			postInput, ok := input.(*types.PostToolUseHookInput)
			if !ok {
				return nil, nil
			}

			fmt.Printf("[PostToolUse] %s", postInput.ToolName)

			// Log success/failure
			if postInput.ToolResponse != nil {
				// Show truncated output
				str := fmt.Sprintf("%v", postInput.ToolResponse)
				if len(str) > 100 {
					str = str[:97] + "..."
				}
				fmt.Printf(" -> %s\n", str)
			} else {
				fmt.Println()
			}

			return nil, nil
		}),

		// Stop hook: Called when Claude wants to stop
		sdk.WithStopHook(nil, func(input any, toolUseID *string, ctx *types.HookContext) (*types.HookOutput, error) {
			stopInput, ok := input.(*types.StopHookInput)
			if !ok {
				return nil, nil
			}

			fmt.Printf("\n[Stop] Stop hook active: %v\n", stopInput.StopHookActive)
			return nil, nil
		}),
	)

	// Connect
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Hooks Example - All tool calls are logged")
	fmt.Println("------------------------------------------")
	fmt.Println()

	// Send a query that will use tools
	prompt := "List the files in the current directory and show me the contents of README.md if it exists."
	fmt.Printf("Prompt: %s\n\n", prompt)

	if err := client.SendQuery(prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	fmt.Println("Response:")
	fmt.Println("---------")
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			fmt.Println()
			goto done
		}
	}
done:

	// Print metrics summary
	fmt.Println()
	fmt.Println("------------------------------------------")
	fmt.Printf("Summary: %d tool calls\n", toolCallCount)
	if len(toolDurations) > 0 {
		fmt.Println("Tool durations:")
		for tool, dur := range toolDurations {
			fmt.Printf("  - %s: %.2fs\n", tool, dur.Seconds())
		}
	}
}
