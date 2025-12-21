// Example: Error handling patterns with Claude Agent SDK.
//
// This demonstrates how to properly handle various error types:
// - CLI not found
// - Version mismatches
// - Connection failures
// - Process errors
// - Parse errors
// - Context cancellation
//
// Usage:
//
//	go run examples/error-handling/main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Error Handling Example")
	fmt.Println("======================")
	fmt.Println()

	// Example 1: Basic error handling
	fmt.Println("1. Basic Error Handling")
	fmt.Println("   Testing connection and query...")
	basicExample()

	// Example 2: Timeout handling
	fmt.Println("\n2. Timeout Handling")
	fmt.Println("   Testing with very short timeout...")
	timeoutExample()

	// Example 3: Recovery patterns
	fmt.Println("\n3. Recovery Patterns")
	fmt.Println("   Demonstrating retry logic...")
	recoveryExample()

	fmt.Println("\nAll error handling examples complete")
}

func basicExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages, err := sdk.RunQuery(ctx, "Say 'hello'")
	if err != nil {
		handleError("Query", err)
		return
	}

	for _, msg := range messages {
		if m, ok := msg.(*sdk.AssistantMessage); ok {
			fmt.Printf("   Response: %s\n", m.Text())
		}
	}
}

func timeoutExample() {
	// Very short timeout to demonstrate handling
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := sdk.RunQuery(ctx, "This will likely timeout")
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("   Timeout occurred (expected behavior)")
		} else {
			handleError("Timeout test", err)
		}
		return
	}
	fmt.Println("   Query completed (unexpectedly fast!)")
}

func recoveryExample() {
	ctx := context.Background()

	// Retry logic with exponential backoff
	var lastErr error
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("   Attempt %d/%d... ", attempt, maxRetries)

		queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		_, err := sdk.RunQuery(queryCtx, "Say 'success'")
		cancel()

		if err == nil {
			fmt.Println("Success")
			return
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err) {
			fmt.Printf("Non-retryable error: %v\n", err)
			return
		}

		fmt.Printf("Retryable error: %v\n", err)

		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<(attempt-1))
			fmt.Printf("   Waiting %v before retry...\n", delay)
			time.Sleep(delay)
		}
	}

	fmt.Printf("   All retries exhausted. Last error: %v\n", lastErr)
}

// handleError demonstrates comprehensive error handling
func handleError(operation string, err error) {
	fmt.Printf("   %s failed: ", operation)

	switch {
	// SDK-specific errors
	case errors.Is(err, sdk.ErrCLINotFound):
		fmt.Println("Claude CLI not found")
		fmt.Println("      -> Install: npm install -g @anthropic-ai/claude-code")

	case errors.Is(err, sdk.ErrCLIVersion):
		fmt.Println("CLI version too old")
		fmt.Println("      -> Update: npm update -g @anthropic-ai/claude-code")

	case errors.Is(err, sdk.ErrConnection):
		fmt.Println("Failed to connect to CLI")
		fmt.Println("      -> Check CLI installation and permissions")

	case errors.Is(err, sdk.ErrProcess):
		fmt.Println("CLI process exited unexpectedly")
		fmt.Println("      -> Check stderr for details")

	case errors.Is(err, sdk.ErrParse):
		fmt.Println("Failed to parse CLI output")
		fmt.Println("      -> This may indicate a CLI version incompatibility")

	case errors.Is(err, sdk.ErrClosed):
		fmt.Println("Transport was closed")
		fmt.Println("      -> Reconnect before sending more queries")

	// Context errors
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Println("Request timed out")
		fmt.Println("      -> Consider increasing timeout duration")

	case errors.Is(err, context.Canceled):
		fmt.Println("Request was cancelled")

	// Generic error
	default:
		fmt.Printf("%v\n", err)
	}
}

// isRetryable determines if an error should trigger a retry
func isRetryable(err error) bool {
	// Don't retry on context cancellation
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Don't retry on CLI not found - it won't magically appear
	if errors.Is(err, sdk.ErrCLINotFound) {
		return false
	}

	// Don't retry on version mismatch - needs manual update
	if errors.Is(err, sdk.ErrCLIVersion) {
		return false
	}

	// Retry on timeouts, connection issues, process exits
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, sdk.ErrConnection) ||
		errors.Is(err, sdk.ErrProcess)
}
