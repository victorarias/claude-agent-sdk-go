// Example: Using WithStderrCallback for debugging CLI communication.
//
// This example demonstrates how to capture stderr output from the Claude CLI
// subprocess for debugging purposes. This is particularly useful when:
//   - Diagnosing CLI communication issues
//   - Troubleshooting unexpected behavior
//   - Monitoring subprocess health
//   - Debugging MCP server interactions
//
// Usage:
//
//	go run examples/stderr-callback/main.go "What is 2+2?"
//	go run examples/stderr-callback/main.go "List files in the current directory"
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: stderr-callback <prompt>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  stderr-callback \"What is 2+2?\"")
		fmt.Println("  stderr-callback \"List files in the current directory\"")
		os.Exit(1)
	}

	prompt := os.Args[1]

	// Setup stderr logging
	// In production, you might want to use a proper logger or write to a file
	stderrLogger := log.New(os.Stderr, "[CLI STDERR] ", log.LstdFlags)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Create client with stderr callback
	// The callback receives each line of stderr output from the Claude CLI subprocess
	client := sdk.NewClient(
		types.WithStderrCallback(func(line string) {
			// Log stderr output with timestamp
			stderrLogger.Println(line)
		}),
	)

	// Connect to Claude CLI
	if err := client.Connect(ctx); err != nil {
		handleError(err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Sending query (watch stderr output above)...")
	fmt.Println()

	// Send query
	if err := client.SendQuery(prompt); err != nil {
		handleError(err)
		os.Exit(1)
	}

	// Receive and process response
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			handleError(err)
			os.Exit(1)
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			// Print Claude's response
			fmt.Print(m.Text())

			// Show tool usage
			if m.HasToolCalls() {
				fmt.Printf("\n[Used %d tool(s)]\n", len(m.ToolCalls()))
			}

		case *types.ResultMessage:
			// Print final summary
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
			if !m.IsSuccess() {
				fmt.Printf("[Status: %s]\n", m.Subtype)
			}
			return
		}
	}
}

// handleError demonstrates proper error handling patterns
func handleError(err error) {
	switch {
	case errors.Is(err, types.ErrCLINotFound):
		fmt.Fprintln(os.Stderr, "Error: Claude CLI not found")
		fmt.Fprintln(os.Stderr, "Install with: npm install -g @anthropic-ai/claude-code")

	case errors.Is(err, types.ErrCLIVersion):
		fmt.Fprintln(os.Stderr, "Error: Claude CLI version too old")
		fmt.Fprintln(os.Stderr, "Update with: npm update -g @anthropic-ai/claude-code")

	case errors.Is(err, context.DeadlineExceeded):
		fmt.Fprintln(os.Stderr, "Error: Request timed out")

	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
