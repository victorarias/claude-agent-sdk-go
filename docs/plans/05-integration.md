# Plan 05: Integration & Examples

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create examples, documentation, and integration tests for the SDK.

**Architecture:** Create example programs demonstrating common use cases. Add integration tests with mock CLI. Document public API.

**Tech Stack:** Go 1.21+

---

## Task 1: Simple Query Example

**Files:**
- Create: `examples/simple/main.go`

**Step 1: Create example**

Create `examples/simple/main.go`:

```go
// Example: Simple one-shot query with Claude Agent SDK.
//
// Usage:
//   go run examples/simple/main.go "What is 2+2?"
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: simple <prompt>")
		os.Exit(1)
	}

	prompt := os.Args[1]

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Query Claude
	messages, err := sdk.Query(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print response
	for _, msg := range messages {
		switch m := msg.(type) {
		case *sdk.AssistantMessage:
			fmt.Println(m.Text())
		case *sdk.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
		}
	}
}
```

**Step 2: Verify it compiles**

```bash
go build ./examples/simple/
```

Expected: No errors

**Step 3: Commit**

```bash
git add examples/simple/main.go
git commit -m "docs: add simple query example"
```

---

## Task 2: Streaming Example

**Files:**
- Create: `examples/streaming/main.go`

**Step 1: Create example**

Create `examples/streaming/main.go`:

```go
// Example: Streaming conversation with Claude Agent SDK.
//
// Usage:
//   go run examples/streaming/main.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	ctx := context.Background()
	scanner := bufio.NewScanner(os.Stdin)

	// Create client with options
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),
		sdk.WithMaxTurns(10),
	)

	// Connect in streaming mode
	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Connected to Claude. Type 'quit' to exit.")
	fmt.Println()

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "quit" || input == "exit" {
			break
		}

		// Send query
		if err := client.SendQuery(input); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
			continue
		}

		// Receive and print response
		fmt.Print("Claude: ")
		for {
			msg, err := client.ReceiveMessage()
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
				break
			}

			switch m := msg.(type) {
			case *sdk.AssistantMessage:
				fmt.Print(m.Text())
			case *sdk.ResultMessage:
				fmt.Println()
				if m.TotalCostUSD != nil {
					fmt.Printf("[Cost: $%.4f]\n", *m.TotalCostUSD)
				}
				goto nextPrompt
			}
		}
	nextPrompt:
		fmt.Println()
	}

	fmt.Println("Goodbye!")
}
```

**Step 2: Verify it compiles**

```bash
go build ./examples/streaming/
```

Expected: No errors

**Step 3: Commit**

```bash
git add examples/streaming/main.go
git commit -m "docs: add streaming conversation example"
```

---

## Task 3: Tool Permission Example

**Files:**
- Create: `examples/permissions/main.go`

**Step 1: Create example**

Create `examples/permissions/main.go`:

```go
// Example: Custom tool permissions with Claude Agent SDK.
//
// This example shows how to intercept and approve/deny tool calls.
//
// Usage:
//   go run examples/permissions/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	ctx := context.Background()

	// Create client
	client := sdk.NewClient()

	// Connect in streaming mode
	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Set up permission callback
	// Note: In a real implementation, this would be set via options
	// before connecting. This is just to illustrate the pattern.

	fmt.Println("Tool Permission Example")
	fmt.Println()
	fmt.Println("This example demonstrates how to control tool permissions.")
	fmt.Println()

	// In a real implementation:
	// - Client would accept a CanUseToolCallback in options
	// - The callback would be called for each tool invocation
	// - Return PermissionResultAllow to approve
	// - Return PermissionResultDeny to block

	exampleCallback := func(toolName string, input map[string]any, ctx *sdk.ToolPermissionContext) (any, error) {
		fmt.Printf("\n[Permission Request]\n")
		fmt.Printf("  Tool: %s\n", toolName)
		fmt.Printf("  Input: %v\n", input)

		// Example: Block dangerous commands
		if toolName == "Bash" {
			if cmd, ok := input["command"].(string); ok {
				if strings.Contains(cmd, "rm -rf") {
					fmt.Println("  Decision: DENIED (dangerous command)")
					return &sdk.PermissionResultDeny{
						Behavior: "deny",
						Message:  "Dangerous command blocked",
					}, nil
				}
			}
		}

		fmt.Println("  Decision: ALLOWED")
		return &sdk.PermissionResultAllow{
			Behavior: "allow",
		}, nil
	}

	// Print the callback signature for documentation
	_ = exampleCallback

	fmt.Println("Permission callback registered (simulated).")
	fmt.Println("In production, dangerous commands would be blocked.")
}
```

**Step 2: Verify it compiles**

```bash
go build ./examples/permissions/
```

Expected: No errors

**Step 3: Commit**

```bash
git add examples/permissions/main.go
git commit -m "docs: add tool permissions example"
```

---

## Task 4: Integration Test

**Files:**
- Create: `integration_test.go`

**Step 1: Create integration test**

Create `integration_test.go`:

```go
//go:build integration

package sdk

import (
	"context"
	"os"
	"testing"
	"time"
)

// Integration tests require the Claude CLI to be installed.
// Run with: go test -tags=integration -v

func TestIntegration_SimpleQuery(t *testing.T) {
	if os.Getenv("CLAUDE_TEST_INTEGRATION") == "" {
		t.Skip("Set CLAUDE_TEST_INTEGRATION=1 to run integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	messages, err := Query(ctx, "What is 2+2? Reply with just the number.")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("No messages returned")
	}

	// Should have at least an assistant message and result
	var gotAssistant, gotResult bool
	for _, msg := range messages {
		switch msg.(type) {
		case *AssistantMessage:
			gotAssistant = true
		case *ResultMessage:
			gotResult = true
		}
	}

	if !gotAssistant {
		t.Error("No assistant message received")
	}
	if !gotResult {
		t.Error("No result message received")
	}
}

func TestIntegration_StreamingConversation(t *testing.T) {
	if os.Getenv("CLAUDE_TEST_INTEGRATION") == "" {
		t.Skip("Set CLAUDE_TEST_INTEGRATION=1 to run integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := NewClient(
		WithMaxTurns(2),
	)

	if err := client.Connect(ctx, ""); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	// First query
	if err := client.SendQuery("What is 1+1? Reply with just the number."); err != nil {
		t.Fatalf("SendQuery failed: %v", err)
	}

	messages, err := client.ReceiveAll()
	if err != nil {
		t.Fatalf("ReceiveAll failed: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("No messages received")
	}

	// Verify we got a result
	lastMsg := messages[len(messages)-1]
	if _, ok := lastMsg.(*ResultMessage); !ok {
		t.Errorf("Expected ResultMessage, got %T", lastMsg)
	}
}
```

**Step 2: Verify it compiles**

```bash
go build -tags=integration ./...
```

Expected: No errors

**Step 3: Commit**

```bash
git add integration_test.go
git commit -m "test: add integration tests"
```

---

## Task 5: Package Documentation

**Files:**
- Modify: `sdk.go`
- Create: `doc.go`

**Step 1: Create doc.go**

Create `doc.go`:

```go
// Package sdk provides a Go client for the Claude Agent SDK.
//
// The SDK spawns the Claude CLI as a subprocess and communicates via
// JSON streaming for bidirectional control protocol.
//
// # Quick Start
//
// Simple one-shot query:
//
//	ctx := context.Background()
//	messages, err := sdk.Query(ctx, "What is 2+2?")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, msg := range messages {
//	    if m, ok := msg.(*sdk.AssistantMessage); ok {
//	        fmt.Println(m.Text())
//	    }
//	}
//
// Streaming query:
//
//	ctx := context.Background()
//	msgChan, errChan := sdk.QueryStream(ctx, "Explain quantum computing")
//	for msg := range msgChan {
//	    if m, ok := msg.(*sdk.AssistantMessage); ok {
//	        fmt.Print(m.Text())
//	    }
//	}
//	if err := <-errChan; err != nil {
//	    log.Fatal(err)
//	}
//
// Interactive conversation:
//
//	client := sdk.NewClient(sdk.WithModel("claude-sonnet-4-5"))
//	if err := client.Connect(ctx, ""); err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	client.SendQuery("Hello!")
//	for {
//	    msg, err := client.ReceiveMessage()
//	    if err != nil {
//	        break
//	    }
//	    switch m := msg.(type) {
//	    case *sdk.AssistantMessage:
//	        fmt.Print(m.Text())
//	    case *sdk.ResultMessage:
//	        fmt.Printf("\nCost: $%.4f\n", m.Cost())
//	        return
//	    }
//	}
//
// # Configuration
//
// Use functional options to configure the client:
//
//	client := sdk.NewClient(
//	    sdk.WithModel("claude-opus-4"),
//	    sdk.WithMaxTurns(10),
//	    sdk.WithPermissionMode(sdk.PermissionBypass),
//	    sdk.WithCwd("/path/to/project"),
//	    sdk.WithEnv(map[string]string{"API_KEY": "secret"}),
//	)
//
// # Message Types
//
// The SDK returns several message types:
//
//   - UserMessage: The user's input
//   - AssistantMessage: Claude's response (may contain text, tool calls, thinking)
//   - SystemMessage: System events (init, notifications)
//   - ResultMessage: Final result with cost and session info
//   - StreamEvent: Partial updates during streaming
//
// # Tool Permissions
//
// Control which tools Claude can use:
//
//	// Via options
//	sdk.WithTools("Read", "Bash")  // Only allow these tools
//
//	// Or via permission callback (streaming mode only)
//	// See examples/permissions for details
//
// # Error Handling
//
// The SDK defines several error types:
//
//   - CLINotFoundError: Claude CLI not installed
//   - ConnectionError: Failed to connect to CLI
//   - ProcessError: CLI process exited with error
//   - ParseError: Failed to parse CLI output
//
// Use errors.Is to check error types:
//
//	if errors.Is(err, sdk.ErrCLINotFound) {
//	    fmt.Println("Please install Claude CLI")
//	}
//
// # Requirements
//
// The Claude CLI must be installed:
//
//	npm install -g @anthropic-ai/claude-code
//
// Or provide a custom path:
//
//	sdk.WithCLIPath("/path/to/claude")
package sdk
```

**Step 2: Update sdk.go**

Update `sdk.go`:

```go
package sdk

// Version is the SDK version.
const Version = "0.1.0"

// MinimumCLIVersion is the minimum supported CLI version.
const MinimumCLIVersion = "2.0.0"
```

**Step 3: Verify docs**

```bash
go doc ./...
```

Expected: Documentation is generated

**Step 4: Commit**

```bash
git add sdk.go doc.go
git commit -m "docs: add package documentation"
```

---

## Summary

After completing all plans, you have:

**Plan 01: Types & Interfaces**
- Error types
- Content blocks (Text, Thinking, ToolUse, ToolResult)
- Message types (User, Assistant, System, Result)
- Options with functional pattern
- Hook and control protocol types
- Transport interface

**Plan 02: Transport Layer**
- CLI discovery
- Command building
- SubprocessTransport implementation
- Message reading/writing
- Error handling

**Plan 03: Query Protocol**
- Control request/response routing
- Initialize, Interrupt, SetPermissionMode
- Hook callback handling
- Stream input
- RewindFiles

**Plan 04: Client API**
- Client structure
- Query and QueryStream functions
- Streaming methods
- Context manager pattern
- Message helpers

**Plan 05: Integration**
- Simple query example
- Streaming example
- Permissions example
- Integration tests
- Package documentation

**Total: ~2,100 lines of Go code across 39 tasks**

---

## Next Steps

1. Create git repository: `git init`
2. Push to GitHub: `gh repo create`
3. Add CI/CD with GitHub Actions
4. Publish to pkg.go.dev

**Plan complete and saved.** Ready for execution!
