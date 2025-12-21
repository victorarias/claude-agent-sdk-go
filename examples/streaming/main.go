// Example: Streaming conversation with Claude Agent SDK.
//
// This demonstrates multi-turn interactive conversations with streaming output.
// Text appears as Claude generates it, providing real-time feedback.
//
// Usage:
//
//	go run examples/streaming/main.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\nInterrupted. Goodbye!")
		cancel()
	}()

	scanner := bufio.NewScanner(os.Stdin)

	// Create client with options
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithMaxTurns(10),
		types.WithSystemPrompt("You are a helpful assistant. Be concise but thorough."),
	)

	// Connect in streaming mode
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Connected to Claude. Commands: 'quit', 'clear', 'cost'")
	fmt.Println("---------------------------------------------------")
	fmt.Println()

	var totalCost float64

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "quit", "exit":
			fmt.Printf("\nTotal session cost: $%.4f\n", totalCost)
			fmt.Println("Goodbye!")
			return
		case "clear":
			// Start fresh conversation
			client.Close()
			if err := client.Connect(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to reconnect: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("[Conversation cleared]")
			fmt.Println()
			continue
		case "cost":
			fmt.Printf("[Current session cost: $%.4f]\n\n", totalCost)
			continue
		}

		// Send query
		if err := client.SendQuery(input); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
			continue
		}

		// Receive and print response with streaming
		fmt.Print("\nClaude: ")
		for {
			msg, err := client.ReceiveMessage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
				break
			}

			switch m := msg.(type) {
			case *types.AssistantMessage:
				// Stream text as it arrives
				fmt.Print(m.Text())

			case *types.StreamEvent:
				// Handle partial updates for true streaming
				if m.EventType == "content_block_delta" {
					if text, ok := m.Delta["text"].(string); ok {
						fmt.Print(text)
					}
				}

			case *types.ResultMessage:
				fmt.Println()
				if m.TotalCostUSD != nil {
					cost := *m.TotalCostUSD
					totalCost += cost
					fmt.Printf("[Turn cost: $%.4f | Total: $%.4f]\n", cost, totalCost)
				}
				goto nextPrompt
			}
		}
	nextPrompt:
		fmt.Println()
	}

	fmt.Printf("\nTotal session cost: $%.4f\n", totalCost)
	fmt.Println("Goodbye!")
}
