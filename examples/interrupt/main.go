// Example: Interrupt capability with Claude Agent SDK.
//
// This demonstrates how to interrupt a running Claude operation:
// - Send interrupt signal to stop long-running tasks
// - Handle Ctrl+C gracefully
// - Use context cancellation for timeouts
//
// Usage:
//
//	go run examples/interrupt/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	fmt.Println("Interrupt Example")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("This will start a long task. Press Ctrl+C to interrupt.")
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithMaxTurns(10),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Start a long-running task
	longTask := "Please count from 1 to 50, saying each number on a new line. Take your time."
	fmt.Printf("Starting task: %s\n\n", longTask)

	if err := client.SendQuery(longTask); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Handle messages in a goroutine
	var wg sync.WaitGroup
	done := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(done)

		for {
			msg, err := client.ReceiveMessage()
			if err != nil {
				fmt.Printf("\n[Message loop ended: %v]\n", err)
				return
			}

			switch m := msg.(type) {
			case *types.AssistantMessage:
				fmt.Print(m.Text())
			case *types.ResultMessage:
				fmt.Println()
				if m.IsSuccess() {
					fmt.Println("\n[Task completed normally]")
				} else {
					fmt.Printf("\n[Task ended: %s]\n", m.Subtype)
				}
				return
			}
		}
	}()

	// Wait for signal or completion
	select {
	case <-sigChan:
		fmt.Println("\n\n[Received interrupt signal, sending interrupt to Claude...]")

		// Send interrupt to Claude
		if err := client.Interrupt(); err != nil {
			fmt.Printf("[Interrupt failed: %v]\n", err)
		} else {
			fmt.Println("[Interrupt sent successfully]")
		}

		// Give Claude time to respond to interrupt
		select {
		case <-done:
			fmt.Println("[Task interrupted cleanly]")
		case <-time.After(5 * time.Second):
			fmt.Println("[Timeout waiting for interrupt response]")
			cancel()
		}

	case <-done:
		// Task completed normally
	}

	// Wait for message handler to finish
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		fmt.Println("[Cleanup complete]")
	case <-time.After(5 * time.Second):
		fmt.Println("[Timeout waiting for cleanup]")
	}
}
