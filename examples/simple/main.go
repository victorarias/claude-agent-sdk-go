// Example: Simple one-shot query with Claude Agent SDK.
//
// This is the simplest way to use the SDK - send a prompt, get a response.
//
// Usage:
//
//	go run examples/simple/main.go "What is 2+2?"
//	go run examples/simple/main.go "Explain Go interfaces"
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: simple <prompt>")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  simple \"What is 2+2?\"")
		fmt.Println("  simple \"Write a haiku about Go\"")
		os.Exit(1)
	}

	prompt := os.Args[1]

	// Create context with timeout
	// Always use timeouts in production to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Query Claude with default options
	// This is the simplest API - one function call, complete response
	messages, err := sdk.RunQuery(ctx, prompt)
	if err != nil {
		handleError(err)
		os.Exit(1)
	}

	// Print response
	// Messages include assistant responses and a final result
	for _, msg := range messages {
		switch m := msg.(type) {
		case *sdk.AssistantMessage:
			// Print Claude's text response
			fmt.Println(m.Text())

			// Show if any tools were used
			if m.HasToolCalls() {
				fmt.Printf("\n[Used %d tool(s)]\n", len(m.ToolCalls()))
			}

		case *sdk.ResultMessage:
			// Print final cost summary
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
			if !m.IsSuccess() {
				fmt.Printf("[Status: %s]\n", m.Subtype)
			}
		}
	}
}

// handleError demonstrates proper error handling patterns
func handleError(err error) {
	switch {
	case errors.Is(err, sdk.ErrCLINotFound):
		fmt.Fprintln(os.Stderr, "Error: Claude CLI not found")
		fmt.Fprintln(os.Stderr, "Install with: npm install -g @anthropic-ai/claude-code")

	case errors.Is(err, sdk.ErrVersionMismatch):
		fmt.Fprintln(os.Stderr, "Error: Claude CLI version too old")
		fmt.Fprintln(os.Stderr, "Update with: npm update -g @anthropic-ai/claude-code")

	case errors.Is(err, context.DeadlineExceeded):
		fmt.Fprintln(os.Stderr, "Error: Request timed out")

	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
