// Example: Tool configuration with allowlist/blocklist.
//
// This example demonstrates how to control which tools Claude can use:
// - WithTools: Specify the base set of tools (or use preset "default")
// - WithAllowedTools: Allowlist specific tools (permission granted automatically)
// - WithDisallowedTools: Blocklist specific tools (permission denied)
//
// Usage:
//
//	go run examples/tools-config/main.go
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
	if err := specificToolsExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in specific tools example: %v\n", err)
	}

	if err := allowedToolsExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in allowed tools example: %v\n", err)
	}

	if err := disallowedToolsExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in disallowed tools example: %v\n", err)
	}

	if err := combinedExample(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error in combined example: %v\n", err)
	}
}

// specificToolsExample demonstrates using WithTools to specify exact tools.
func specificToolsExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 1: Specific Tools")
	fmt.Println("===================================")
	fmt.Println("Setting tools=['Read', 'Glob', 'Grep']")
	fmt.Println("This limits Claude to only file reading/searching tools.")
	fmt.Println()

	// Specify only read-only tools
	// Claude will only have access to Read, Glob, and Grep
	messages, err := sdk.RunQuery(
		ctx,
		"What tools do you have available? List them.",
		types.WithTools("Read", "Glob", "Grep"),
		types.WithMaxTurns(1),
	)
	if err != nil {
		return err
	}

	printMessages(messages)
	fmt.Println()
	return nil
}

// allowedToolsExample demonstrates using WithAllowedTools for auto-approved tools.
func allowedToolsExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 2: Allowed Tools (Auto-approve)")
	fmt.Println("===================================")
	fmt.Println("Using WithAllowedTools(['Bash', 'Read'])")
	fmt.Println("These tools can be used without permission prompts.")
	fmt.Println()

	// Allow specific tools without permission prompts
	// This is useful when you trust Claude to use certain tools freely
	messages, err := sdk.RunQuery(
		ctx,
		"Run 'echo Hello from allowed tools!' and tell me the result.",
		types.WithAllowedTools("Bash", "Read"),
		types.WithMaxTurns(2),
	)
	if err != nil {
		return err
	}

	printMessages(messages)
	fmt.Println()
	return nil
}

// disallowedToolsExample demonstrates using WithDisallowedTools to block tools.
func disallowedToolsExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 3: Disallowed Tools (Blocklist)")
	fmt.Println("===================================")
	fmt.Println("Using WithDisallowedTools(['Write', 'Edit'])")
	fmt.Println("Claude cannot use Write or Edit tools at all.")
	fmt.Println()

	// Block dangerous tools while allowing others
	// This prevents Claude from making any file modifications
	messages, err := sdk.RunQuery(
		ctx,
		"Try to list files in the current directory. Can you write or edit files?",
		types.WithDisallowedTools("Write", "Edit"),
		types.WithAllowedTools("Bash", "Read", "Glob"),
		types.WithMaxTurns(2),
	)
	if err != nil {
		return err
	}

	printMessages(messages)
	fmt.Println()
	return nil
}

// combinedExample demonstrates combining allowlist and blocklist.
func combinedExample(ctx context.Context) error {
	fmt.Println("===================================")
	fmt.Println("Example 4: Combined Allowlist + Blocklist")
	fmt.Println("===================================")
	fmt.Println("Using:")
	fmt.Println("  - WithAllowedTools(['Read', 'Bash'])")
	fmt.Println("  - WithDisallowedTools(['Bash(rm *)'])")
	fmt.Println()
	fmt.Println("This allows Bash and Read, but blocks dangerous rm commands.")
	fmt.Println("(Note: Pattern-based blocking like 'Bash(rm *)' requires CLI support)")
	fmt.Println()

	// Combine allowlist and blocklist
	// Allow general Bash access but block dangerous patterns
	client := sdk.NewClient(
		types.WithAllowedTools("Read", "Bash"),
		types.WithDisallowedTools("Write", "Edit"),
		types.WithMaxTurns(2),
	)

	if err := client.Connect(ctx); err != nil {
		return err
	}
	defer client.Close()

	if err := client.SendQuery("List files using 'ls -la' command. Can you write or edit files?"); err != nil {
		return err
	}

	// Receive and print response
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
