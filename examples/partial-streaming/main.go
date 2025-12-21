// Example: Partial message streaming with Claude Agent SDK.
//
// This demonstrates real-time streaming with partial updates:
// - Enable include_partial_messages for character-by-character output
// - Handle StreamEvent messages for delta updates
// - Build responsive UIs with immediate feedback
//
// Usage:
//
//	go run examples/partial-streaming/main.go
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

	fmt.Println("Partial Streaming Example")
	fmt.Println("=========================")
	fmt.Println()
	fmt.Println("Watch text appear as it's generated...")
	fmt.Println()

	// Enable partial message streaming for real-time updates
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),
		sdk.WithPartialMessages(), // Enable streaming deltas
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	if err := client.SendQuery("Write a short poem about Go programming (4 lines)."); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Claude: ")

	var charCount int
	startTime := time.Now()

	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *sdk.StreamEvent:
			// Handle partial updates - text arrives incrementally
			if m.EventType == "content_block_delta" {
				if delta := m.Delta; delta != nil {
					if text, ok := delta["text"].(string); ok {
						fmt.Print(text)
						charCount += len(text)
					}
				}
			}

		case *sdk.AssistantMessage:
			// Full message - confirms the complete response
			// Already printed via stream events above

		case *sdk.ResultMessage:
			elapsed := time.Since(startTime)
			fmt.Println()
			fmt.Println()
			fmt.Println("=========================")
			fmt.Printf("Characters streamed: %d\n", charCount)
			fmt.Printf("Time elapsed: %.2fs\n", elapsed.Seconds())
			if charCount > 0 && elapsed > 0 {
				cps := float64(charCount) / elapsed.Seconds()
				fmt.Printf("Streaming rate: %.1f chars/sec\n", cps)
			}
			if m.TotalCostUSD != nil {
				fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
			}
			return
		}
	}
}
