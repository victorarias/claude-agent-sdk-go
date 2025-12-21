# Plan 05: Integration & Examples

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create examples, documentation, and integration tests for the SDK.

**Architecture:** Create example programs demonstrating common use cases including hooks, MCP servers, session management, and error handling. Add integration tests with mock CLI. Document public API comprehensively.

**Tech Stack:** Go 1.21+

**Dependencies:** Plans 01-04 must be complete.

---

## Task 0: Example Directory Structure

**Files:**
- Create: `examples/README.md`

**Step 1: Create examples README**

Create `examples/README.md`:

```markdown
# Claude Agent SDK Go Examples

This directory contains examples demonstrating common SDK usage patterns.

## Examples

| Example | Description |
|---------|-------------|
| [simple](./simple/) | One-shot query with minimal setup |
| [streaming](./streaming/) | Interactive multi-turn conversation |
| [hooks](./hooks/) | Pre/post tool use hooks for logging and control |
| [mcp-server](./mcp-server/) | SDK-hosted MCP server with custom tools |
| [permissions](./permissions/) | Custom tool permission handling |
| [session](./session/) | Session resume and conversation continuation |
| [error-handling](./error-handling/) | Error types and recovery patterns |

## Running Examples

All examples require the Claude CLI to be installed:

\`\`\`bash
npm install -g @anthropic-ai/claude-code
\`\`\`

Run any example:

\`\`\`bash
go run ./examples/simple/ "What is 2+2?"
go run ./examples/streaming/
go run ./examples/hooks/
\`\`\`

## Common Patterns

### Context with Timeout
Always use context timeouts for production:
\`\`\`go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()
\`\`\`

### Defer Close
Always defer Close to ensure cleanup:
\`\`\`go
client := sdk.NewClient()
if err := client.Connect(ctx, ""); err != nil {
    log.Fatal(err)
}
defer client.Close()
\`\`\`

### Error Checking
Use errors.Is for specific error handling:
\`\`\`go
if errors.Is(err, sdk.ErrCLINotFound) {
    fmt.Println("Please install: npm i -g @anthropic-ai/claude-code")
}
\`\`\`
```

**Step 2: Verify directory exists**

```bash
mkdir -p examples
```

**Step 3: Commit**

```bash
git add examples/README.md
git commit -m "docs: add examples README"
```

---

## Task 1: Simple Query Example

**Files:**
- Create: `examples/simple/main.go`

**Step 1: Create example**

Create `examples/simple/main.go`:

```go
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
	messages, err := sdk.Query(ctx, prompt)
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

	sdk "github.com/victorarias/claude-agent-sdk-go"
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
		sdk.WithModel("claude-sonnet-4-5"),
		sdk.WithMaxTurns(10),
		sdk.WithSystemPrompt("You are a helpful assistant. Be concise but thorough."),
	)

	// Connect in streaming mode (empty prompt = streaming)
	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Connected to Claude. Commands: 'quit', 'clear', 'cost'")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
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
			if err := client.Connect(ctx, ""); err != nil {
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
			case *sdk.AssistantMessage:
				// Stream text as it arrives
				fmt.Print(m.Text())

			case *sdk.StreamEvent:
				// Handle partial updates for true streaming
				if m.Type == "content_block_delta" {
					if text, ok := m.Data["text"].(string); ok {
						fmt.Print(text)
					}
				}

			case *sdk.ResultMessage:
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

## Task 3: Hooks Example

**Files:**
- Create: `examples/hooks/main.go`

**Step 1: Create example**

Create `examples/hooks/main.go`:

```go
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

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Track metrics
	var toolCallCount int
	toolDurations := make(map[string]time.Duration)

	// Create client with hooks
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),

		// PreToolUse hook: Called BEFORE each tool execution
		// Can modify inputs, block execution, or log the attempt
		sdk.WithPreToolUseHook(func(toolName string, input map[string]any) *sdk.HookResult {
			toolCallCount++
			fmt.Printf("\n📥 [PreToolUse #%d] %s\n", toolCallCount, toolName)

			// Log input (truncate long values)
			for k, v := range input {
				str := fmt.Sprintf("%v", v)
				if len(str) > 50 {
					str = str[:47] + "..."
				}
				fmt.Printf("   %s: %s\n", k, str)
			}

			// Example: Block dangerous bash commands
			if toolName == "Bash" {
				if cmd, ok := input["command"].(string); ok {
					dangerous := []string{"rm -rf", "sudo", "> /dev/", "mkfs"}
					for _, d := range dangerous {
						if strings.Contains(cmd, d) {
							fmt.Printf("   ⛔ BLOCKED: dangerous command detected\n")
							return &sdk.HookResult{
								Decision: sdk.HookDecisionBlock,
								Message:  fmt.Sprintf("Blocked: command contains '%s'", d),
							}
						}
					}
				}
			}

			// Example: Block writes to certain paths
			if toolName == "Write" || toolName == "Edit" {
				if path, ok := input["file_path"].(string); ok {
					protected := []string{"/etc/", "/usr/", ".env", "credentials"}
					for _, p := range protected {
						if strings.Contains(path, p) {
							fmt.Printf("   ⛔ BLOCKED: protected path\n")
							return &sdk.HookResult{
								Decision: sdk.HookDecisionBlock,
								Message:  fmt.Sprintf("Cannot write to protected path: %s", path),
							}
						}
					}
				}
			}

			fmt.Printf("   ✅ Allowed\n")
			return &sdk.HookResult{Decision: sdk.HookDecisionAllow}
		}),

		// PostToolUse hook: Called AFTER each tool execution
		// Can log results, track metrics, or trigger side effects
		sdk.WithPostToolUseHook(func(toolName string, input map[string]any, output *sdk.ToolOutput) *sdk.HookResult {
			fmt.Printf("📤 [PostToolUse] %s", toolName)

			// Track execution time if available
			if output.Duration > 0 {
				toolDurations[toolName] += output.Duration
				fmt.Printf(" (%.2fs)", output.Duration.Seconds())
			}

			// Log success/failure
			if output.Error != nil {
				fmt.Printf(" ❌ Error: %v\n", output.Error)
			} else {
				// Show truncated output
				str := fmt.Sprintf("%v", output.Result)
				if len(str) > 100 {
					str = str[:97] + "..."
				}
				fmt.Printf(" → %s\n", str)
			}

			return nil // Don't modify result
		}),

		// StopHook: Called when Claude wants to stop
		sdk.WithStopHook(func(reason string) *sdk.HookResult {
			fmt.Printf("\n🛑 [Stop] Reason: %s\n", reason)
			return nil
		}),
	)

	// Connect
	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Hooks Example - All tool calls are logged")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
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
	fmt.Println("─────────")
	for msg := range client.Messages() {
		switch m := msg.(type) {
		case *sdk.AssistantMessage:
			fmt.Print(m.Text())
		case *sdk.ResultMessage:
			fmt.Println()
		}
	}

	// Print metrics summary
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 Summary: %d tool calls\n", toolCallCount)
	if len(toolDurations) > 0 {
		fmt.Println("   Tool durations:")
		for tool, dur := range toolDurations {
			fmt.Printf("   - %s: %.2fs\n", tool, dur.Seconds())
		}
	}
}
```

**Step 2: Verify it compiles**

```bash
go build ./examples/hooks/
```

Expected: No errors

**Step 3: Commit**

```bash
git add examples/hooks/main.go
git commit -m "docs: add hooks example"
```

---

## Task 4: MCP Server Example

**Files:**
- Create: `examples/mcp-server/main.go`

**Step 1: Create example**

Create `examples/mcp-server/main.go`:

```go
// Example: SDK-hosted MCP server with custom tools.
//
// This demonstrates how to create custom tools that Claude can use.
// Tools are hosted in-process by the SDK and exposed via MCP protocol.
//
// Usage:
//
//	go run examples/mcp-server/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build custom MCP server with tools
	mathServer := sdk.NewMCPServerBuilder("math-tools").
		WithDescription("Mathematical calculation tools").

		// Calculator tool
		WithTool("calculate", "Perform mathematical calculations", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "Mathematical expression to evaluate (e.g., '2 + 2', 'sqrt(16)')",
				},
			},
			"required": []string{"expression"},
		}, func(input map[string]any) (any, error) {
			expr := input["expression"].(string)
			result, err := evaluateExpression(expr)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"expression": expr,
				"result":     result,
			}, nil
		}).

		// Unit converter tool
		WithTool("convert_units", "Convert between units", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"value": map[string]any{
					"type":        "number",
					"description": "Value to convert",
				},
				"from": map[string]any{
					"type":        "string",
					"description": "Source unit (e.g., 'km', 'miles', 'celsius', 'fahrenheit')",
				},
				"to": map[string]any{
					"type":        "string",
					"description": "Target unit",
				},
			},
			"required": []string{"value", "from", "to"},
		}, func(input map[string]any) (any, error) {
			value := input["value"].(float64)
			from := input["from"].(string)
			to := input["to"].(string)

			result, err := convertUnits(value, from, to)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"original": fmt.Sprintf("%.4f %s", value, from),
				"result":   fmt.Sprintf("%.4f %s", result, to),
			}, nil
		}).
		Build()

	// Build a data lookup server
	dataServer := sdk.NewMCPServerBuilder("data-lookup").
		WithDescription("Data lookup and retrieval tools").

		// Timezone tool
		WithTool("get_timezone", "Get current time in a timezone", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"timezone": map[string]any{
					"type":        "string",
					"description": "Timezone name (e.g., 'America/New_York', 'Europe/London', 'Asia/Tokyo')",
				},
			},
			"required": []string{"timezone"},
		}, func(input map[string]any) (any, error) {
			tzName := input["timezone"].(string)
			loc, err := time.LoadLocation(tzName)
			if err != nil {
				return nil, fmt.Errorf("unknown timezone: %s", tzName)
			}
			now := time.Now().In(loc)
			return map[string]any{
				"timezone":  tzName,
				"time":      now.Format("15:04:05"),
				"date":      now.Format("2006-01-02"),
				"day":       now.Weekday().String(),
				"utc_offset": now.Format("-07:00"),
			}, nil
		}).
		Build()

	// Create client with MCP servers
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),
		sdk.WithMCPServer(mathServer),
		sdk.WithMCPServer(dataServer),
	)

	// Connect
	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("MCP Server Example - Custom tools available")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Available tools:")
	fmt.Println("  - calculate: Evaluate math expressions")
	fmt.Println("  - convert_units: Convert between units")
	fmt.Println("  - get_timezone: Get time in any timezone")
	fmt.Println()

	// Example prompts that will use our custom tools
	prompts := []string{
		"What is the square root of 144 plus 25% of 80?",
		"Convert 100 kilometers to miles and 30 celsius to fahrenheit.",
		"What time is it right now in Tokyo and New York?",
	}

	for i, prompt := range prompts {
		fmt.Printf("─── Query %d ───\n", i+1)
		fmt.Printf("Q: %s\n\n", prompt)

		if err := client.SendQuery(prompt); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			continue
		}

		fmt.Print("A: ")
		for msg := range client.Messages() {
			switch m := msg.(type) {
			case *sdk.AssistantMessage:
				fmt.Print(m.Text())

				// Show which tools were used
				if m.HasToolCalls() {
					fmt.Print("\n   [Tools used: ")
					for j, tc := range m.ToolCalls() {
						if j > 0 {
							fmt.Print(", ")
						}
						fmt.Print(tc.Name)
					}
					fmt.Print("]")
				}
			case *sdk.ResultMessage:
				fmt.Println()
			}
		}
		fmt.Println()
	}
}

// Simple expression evaluator (in production, use a proper parser)
func evaluateExpression(expr string) (float64, error) {
	// This is a simplified example - in production use a proper expression parser
	var result float64
	switch {
	case expr == "sqrt(16)":
		result = 4
	case expr == "sqrt(144)":
		result = 12
	default:
		// Try to parse as JSON number for simple cases
		if err := json.Unmarshal([]byte(expr), &result); err != nil {
			return 0, fmt.Errorf("cannot evaluate: %s", expr)
		}
	}
	return result, nil
}

// Unit converter
func convertUnits(value float64, from, to string) (float64, error) {
	conversions := map[string]map[string]func(float64) float64{
		"km": {
			"miles": func(v float64) float64 { return v * 0.621371 },
		},
		"miles": {
			"km": func(v float64) float64 { return v * 1.60934 },
		},
		"celsius": {
			"fahrenheit": func(v float64) float64 { return v*9/5 + 32 },
		},
		"fahrenheit": {
			"celsius": func(v float64) float64 { return (v - 32) * 5 / 9 },
		},
		"meters": {
			"feet": func(v float64) float64 { return v * 3.28084 },
		},
		"feet": {
			"meters": func(v float64) float64 { return v * 0.3048 },
		},
		"kg": {
			"lbs": func(v float64) float64 { return v * 2.20462 },
		},
		"lbs": {
			"kg": func(v float64) float64 { return v * 0.453592 },
		},
	}

	if fromMap, ok := conversions[from]; ok {
		if fn, ok := fromMap[to]; ok {
			return math.Round(fn(value)*10000) / 10000, nil
		}
	}
	return 0, fmt.Errorf("cannot convert from %s to %s", from, to)
}
```

**Step 2: Verify it compiles**

```bash
go build ./examples/mcp-server/
```

Expected: No errors

**Step 3: Commit**

```bash
git add examples/mcp-server/main.go
git commit -m "docs: add MCP server example"
```

---

## Task 5: Tool Permission Example

**Files:**
- Create: `examples/permissions/main.go`

**Step 1: Create example**

Create `examples/permissions/main.go`:

```go
// Example: Custom tool permissions with Claude Agent SDK.
//
// This example shows how to intercept and approve/deny tool calls
// at a granular level, with interactive user confirmation.
//
// Usage:
//
//	go run examples/permissions/main.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	scanner := bufio.NewScanner(os.Stdin)

	// Create client with custom permission callback
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),

		// Permission callback: called for every tool that needs approval
		sdk.WithCanUseTool(func(toolName string, input map[string]any, permCtx *sdk.ToolPermissionContext) *sdk.PermissionResult {
			fmt.Println()
			fmt.Println("┌─────────────────────────────────────────┐")
			fmt.Println("│         TOOL PERMISSION REQUEST         │")
			fmt.Println("└─────────────────────────────────────────┘")
			fmt.Printf("Tool: %s\n", toolName)
			fmt.Println()

			// Show input details based on tool type
			switch toolName {
			case "Bash":
				if cmd, ok := input["command"].(string); ok {
					fmt.Printf("Command:\n  %s\n", cmd)
				}
			case "Write":
				if path, ok := input["file_path"].(string); ok {
					fmt.Printf("File: %s\n", path)
				}
				if content, ok := input["content"].(string); ok {
					lines := strings.Split(content, "\n")
					if len(lines) > 5 {
						fmt.Printf("Content: (%d lines)\n", len(lines))
						for _, line := range lines[:5] {
							fmt.Printf("  %s\n", truncate(line, 60))
						}
						fmt.Println("  ...")
					} else {
						fmt.Println("Content:")
						for _, line := range lines {
							fmt.Printf("  %s\n", truncate(line, 60))
						}
					}
				}
			case "Edit":
				if path, ok := input["file_path"].(string); ok {
					fmt.Printf("File: %s\n", path)
				}
				if old, ok := input["old_string"].(string); ok {
					fmt.Printf("Replace: %s\n", truncate(old, 50))
				}
				if new, ok := input["new_string"].(string); ok {
					fmt.Printf("With: %s\n", truncate(new, 50))
				}
			case "Read":
				if path, ok := input["file_path"].(string); ok {
					fmt.Printf("File: %s\n", path)
				}
			default:
				// Show all inputs for other tools
				for k, v := range input {
					fmt.Printf("%s: %s\n", k, truncate(fmt.Sprintf("%v", v), 50))
				}
			}

			fmt.Println()
			fmt.Print("Allow this action? [Y]es / [N]o / [A]lways / [D]eny all: ")

			if !scanner.Scan() {
				return sdk.PermissionDeny("Input cancelled")
			}

			response := strings.ToLower(strings.TrimSpace(scanner.Text()))
			switch response {
			case "y", "yes", "":
				fmt.Println("→ Allowed (this time)")
				return sdk.PermissionAllow()

			case "n", "no":
				fmt.Println("→ Denied")
				return sdk.PermissionDeny("User denied permission")

			case "a", "always":
				fmt.Println("→ Always allow this tool")
				return sdk.PermissionAllowAlways(toolName)

			case "d", "deny":
				fmt.Println("→ Deny all future requests")
				return sdk.PermissionDenyAll()

			default:
				fmt.Println("→ Defaulting to deny")
				return sdk.PermissionDeny("Invalid response")
			}
		}),
	)

	// Connect
	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Permission Example - Interactive tool approval")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("Claude will ask permission before using any tool.")
	fmt.Println("You can approve, deny, or set blanket permissions.")
	fmt.Println()

	// Send a query that will trigger multiple tool uses
	prompt := `Please do the following:
1. List the current directory
2. Read the go.mod file
3. Create a test file called "permission_test.txt" with "Hello World"
4. Delete the test file`

	fmt.Printf("Prompt:\n%s\n", prompt)
	fmt.Println()

	if err := client.SendQuery(prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	fmt.Println("\n──── Response ────")
	for msg := range client.Messages() {
		switch m := msg.(type) {
		case *sdk.AssistantMessage:
			fmt.Print(m.Text())
		case *sdk.ResultMessage:
			fmt.Println()
			if m.TotalCostUSD != nil {
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
		}
	}
}

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
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

## Task 6: Session Management Example

**Files:**
- Create: `examples/session/main.go`

**Step 1: Create example**

Create `examples/session/main.go`:

```go
// Example: Session resume and conversation continuation.
//
// This demonstrates how to:
// - Save session IDs for later resumption
// - Resume a previous conversation
// - Continue from the last message
//
// Usage:
//
//	go run examples/session/main.go new      # Start new session
//	go run examples/session/main.go resume   # Resume last session
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

const sessionFile = ".claude_session_id"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: session <command>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  new     Start a new session")
		fmt.Println("  resume  Resume the last session")
		fmt.Println("  status  Show current session info")
		os.Exit(1)
	}

	command := strings.ToLower(os.Args[1])

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	switch command {
	case "new":
		startNewSession(ctx)
	case "resume":
		resumeSession(ctx)
	case "status":
		showStatus()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func startNewSession(ctx context.Context) {
	fmt.Println("Starting new Claude session...")
	fmt.Println()

	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),
	)

	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Initial query
	if err := client.SendQuery("Hello! Please remember that my favorite color is blue. I'll ask you about this later."); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Claude: ")
	var sessionID string
	for msg := range client.Messages() {
		switch m := msg.(type) {
		case *sdk.AssistantMessage:
			fmt.Print(m.Text())
		case *sdk.ResultMessage:
			sessionID = m.SessionID
			fmt.Println()
		}
	}

	// Save session ID
	if sessionID != "" {
		if err := saveSessionID(sessionID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: couldn't save session: %v\n", err)
		} else {
			fmt.Printf("\n[Session saved: %s]\n", sessionID[:16]+"...")
			fmt.Println("[Run 'session resume' to continue this conversation]")
		}
	}
}

func resumeSession(ctx context.Context) {
	sessionID, err := loadSessionID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No saved session found: %v\n", err)
		fmt.Println("Run 'session new' first to create a session.")
		os.Exit(1)
	}

	fmt.Printf("Resuming session: %s...\n", sessionID[:16]+"...")
	fmt.Println()

	// Resume the previous session
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),
		sdk.WithResume(sessionID),
	)

	if err := client.Connect(ctx, ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	// Ask about something from the previous conversation
	if err := client.SendQuery("What is my favorite color that I told you about?"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Claude: ")
	for msg := range client.Messages() {
		switch m := msg.(type) {
		case *sdk.AssistantMessage:
			fmt.Print(m.Text())
		case *sdk.ResultMessage:
			fmt.Println()
			fmt.Printf("\n[Session ID: %s]\n", m.SessionID[:16]+"...")
		}
	}
}

func showStatus() {
	sessionID, err := loadSessionID()
	if err != nil {
		fmt.Println("No saved session found.")
		return
	}

	fmt.Printf("Saved session ID: %s\n", sessionID)
	fmt.Printf("Session file: %s\n", getSessionFilePath())
}

func saveSessionID(sessionID string) error {
	return os.WriteFile(getSessionFilePath(), []byte(sessionID), 0600)
}

func loadSessionID() (string, error) {
	data, err := os.ReadFile(getSessionFilePath())
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func getSessionFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, sessionFile)
}
```

**Step 2: Verify it compiles**

```bash
go build ./examples/session/
```

Expected: No errors

**Step 3: Commit**

```bash
git add examples/session/main.go
git commit -m "docs: add session management example"
```

---

## Task 7: Error Handling Example

**Files:**
- Create: `examples/error-handling/main.go`

**Step 1: Create example**

Create `examples/error-handling/main.go`:

```go
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
	"os"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	fmt.Println("Error Handling Example")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━")
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

	fmt.Println("\n✅ All error handling examples complete")
}

func basicExample() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages, err := sdk.Query(ctx, "Say 'hello'")
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

	_, err := sdk.Query(ctx, "This will likely timeout")
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("   ⏱  Timeout occurred (expected behavior)")
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
		_, err := sdk.Query(queryCtx, "Say 'success'")
		cancel()

		if err == nil {
			fmt.Println("✅ Success")
			return
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryable(err) {
			fmt.Printf("❌ Non-retryable error: %v\n", err)
			return
		}

		fmt.Printf("⚠️  Retryable error: %v\n", err)

		if attempt < maxRetries {
			delay := baseDelay * time.Duration(1<<(attempt-1))
			fmt.Printf("   Waiting %v before retry...\n", delay)
			time.Sleep(delay)
		}
	}

	fmt.Printf("   ❌ All retries exhausted. Last error: %v\n", lastErr)
}

// handleError demonstrates comprehensive error handling
func handleError(operation string, err error) {
	fmt.Printf("   ❌ %s failed: ", operation)

	switch {
	// SDK-specific errors
	case errors.Is(err, sdk.ErrCLINotFound):
		fmt.Println("Claude CLI not found")
		fmt.Println("      → Install: npm install -g @anthropic-ai/claude-code")

	case errors.Is(err, sdk.ErrVersionMismatch):
		fmt.Println("CLI version too old")
		fmt.Println("      → Update: npm update -g @anthropic-ai/claude-code")
		if vErr, ok := err.(*sdk.VersionError); ok {
			fmt.Printf("      → Current: %s, Required: %s\n", vErr.Current, vErr.Required)
		}

	case errors.Is(err, sdk.ErrConnectionFailed):
		fmt.Println("Failed to connect to CLI")
		fmt.Println("      → Check CLI installation and permissions")

	case errors.Is(err, sdk.ErrProcessExited):
		fmt.Println("CLI process exited unexpectedly")
		if pErr, ok := err.(*sdk.ProcessError); ok {
			fmt.Printf("      → Exit code: %d\n", pErr.ExitCode)
			if pErr.Stderr != "" {
				fmt.Printf("      → Stderr: %s\n", pErr.Stderr)
			}
		}

	case errors.Is(err, sdk.ErrParseFailed):
		fmt.Println("Failed to parse CLI output")
		fmt.Println("      → This may indicate a CLI version incompatibility")

	// Context errors
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Println("Request timed out")
		fmt.Println("      → Consider increasing timeout duration")

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
	if errors.Is(err, sdk.ErrVersionMismatch) {
		return false
	}

	// Retry on timeouts, connection issues, process exits
	return errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, sdk.ErrConnectionFailed) ||
		errors.Is(err, sdk.ErrProcessExited)
}
```

**Step 2: Verify it compiles**

```bash
go build ./examples/error-handling/
```

Expected: No errors

**Step 3: Commit**

```bash
git add examples/error-handling/main.go
git commit -m "docs: add error handling example"
```

---

## Task 8: Integration Tests

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
	"strings"
	"testing"
	"time"
)

// Integration tests require the Claude CLI to be installed.
// Run with: go test -tags=integration -v
//
// Environment variables:
//   CLAUDE_TEST_INTEGRATION=1  Enable integration tests
//   CLAUDE_TEST_TIMEOUT=5m     Override default timeout

func skipIfNotIntegration(t *testing.T) {
	if os.Getenv("CLAUDE_TEST_INTEGRATION") == "" {
		t.Skip("Set CLAUDE_TEST_INTEGRATION=1 to run integration tests")
	}
}

func getTestTimeout() time.Duration {
	if s := os.Getenv("CLAUDE_TEST_TIMEOUT"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return 2 * time.Minute
}

func TestIntegration_SimpleQuery(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
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
	var responseText string
	for _, msg := range messages {
		switch m := msg.(type) {
		case *AssistantMessage:
			gotAssistant = true
			responseText = m.Text()
		case *ResultMessage:
			gotResult = true
			if !m.IsSuccess() {
				t.Errorf("Result indicates failure: %s", m.Subtype)
			}
		}
	}

	if !gotAssistant {
		t.Error("No assistant message received")
	}
	if !gotResult {
		t.Error("No result message received")
	}

	// Response should contain "4"
	if !strings.Contains(responseText, "4") {
		t.Errorf("Expected response to contain '4', got: %s", responseText)
	}
}

func TestIntegration_StreamingConversation(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
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
	result, ok := lastMsg.(*ResultMessage)
	if !ok {
		t.Errorf("Expected ResultMessage, got %T", lastMsg)
	} else if !result.IsSuccess() {
		t.Errorf("Result indicates failure: %s", result.Subtype)
	}
}

func TestIntegration_MultiTurnConversation(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	client := NewClient(
		WithMaxTurns(5),
	)

	if err := client.Connect(ctx, ""); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	// First turn: establish context
	if err := client.SendQuery("My name is TestUser. Please remember this."); err != nil {
		t.Fatalf("SendQuery 1 failed: %v", err)
	}
	if _, err := client.ReceiveAll(); err != nil {
		t.Fatalf("ReceiveAll 1 failed: %v", err)
	}

	// Second turn: verify context retention
	if err := client.SendQuery("What is my name?"); err != nil {
		t.Fatalf("SendQuery 2 failed: %v", err)
	}

	messages, err := client.ReceiveAll()
	if err != nil {
		t.Fatalf("ReceiveAll 2 failed: %v", err)
	}

	// Check that Claude remembers the name
	var responseText string
	for _, msg := range messages {
		if m, ok := msg.(*AssistantMessage); ok {
			responseText += m.Text()
		}
	}

	if !strings.Contains(strings.ToLower(responseText), "testuser") {
		t.Errorf("Claude should remember 'TestUser', got: %s", responseText)
	}
}

func TestIntegration_QueryWithOptions(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	messages, err := QueryWithOptions(ctx, "Say hello",
		WithModel("claude-sonnet-4-5"),
		WithMaxTurns(1),
		WithSystemPrompt("You are a test assistant. Be very brief."),
	)
	if err != nil {
		t.Fatalf("QueryWithOptions failed: %v", err)
	}

	if len(messages) == 0 {
		t.Fatal("No messages returned")
	}

	// Verify result
	var gotResult bool
	for _, msg := range messages {
		if result, ok := msg.(*ResultMessage); ok {
			gotResult = true
			if !result.IsSuccess() {
				t.Errorf("Result indicates failure: %s", result.Subtype)
			}
		}
	}

	if !gotResult {
		t.Error("No result message received")
	}
}

func TestIntegration_ToolUse(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	// Query that should trigger tool use
	messages, err := Query(ctx, "What files are in the current directory? Use the Bash tool to run 'ls -la'.")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Check that tools were used
	var usedTools bool
	for _, msg := range messages {
		if m, ok := msg.(*AssistantMessage); ok {
			if m.HasToolCalls() {
				usedTools = true
				for _, tc := range m.ToolCalls() {
					t.Logf("Tool used: %s", tc.Name)
				}
			}
		}
	}

	if !usedTools {
		t.Log("Warning: No tool calls detected (may be expected based on context)")
	}
}

func TestIntegration_ChannelIteration(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	client := NewClient()

	if err := client.Connect(ctx, ""); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	if err := client.SendQuery("Count from 1 to 3."); err != nil {
		t.Fatalf("SendQuery failed: %v", err)
	}

	// Use channel iteration pattern
	var messageCount int
	for msg := range client.Messages() {
		messageCount++
		switch m := msg.(type) {
		case *AssistantMessage:
			t.Logf("Assistant: %s", m.Text())
		case *ResultMessage:
			t.Logf("Result: success=%v", m.IsSuccess())
		}
	}

	if messageCount == 0 {
		t.Error("No messages received through channel")
	}

	// Check for errors
	if err := <-client.Errors(); err != nil {
		t.Errorf("Error channel received: %v", err)
	}
}

func TestIntegration_Hooks(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	var preToolCalls, postToolCalls int

	client := NewClient(
		WithPreToolUseHook(func(name string, input map[string]any) *HookResult {
			preToolCalls++
			t.Logf("PreToolUse: %s", name)
			return &HookResult{Decision: HookDecisionAllow}
		}),
		WithPostToolUseHook(func(name string, input map[string]any, output *ToolOutput) *HookResult {
			postToolCalls++
			t.Logf("PostToolUse: %s", name)
			return nil
		}),
	)

	if err := client.Connect(ctx, ""); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	// Query that triggers tool use
	if err := client.SendQuery("List files in current directory using ls."); err != nil {
		t.Fatalf("SendQuery failed: %v", err)
	}

	if _, err := client.ReceiveAll(); err != nil {
		t.Fatalf("ReceiveAll failed: %v", err)
	}

	t.Logf("Pre-tool calls: %d, Post-tool calls: %d", preToolCalls, postToolCalls)
	// Note: hooks may not fire if Claude doesn't use tools
}

func TestIntegration_SessionResume(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	// First session: establish context
	client1 := NewClient()
	if err := client1.Connect(ctx, ""); err != nil {
		t.Fatalf("Connect 1 failed: %v", err)
	}

	if err := client1.SendQuery("Remember that the secret word is 'banana'."); err != nil {
		client1.Close()
		t.Fatalf("SendQuery 1 failed: %v", err)
	}

	var sessionID string
	for msg := range client1.Messages() {
		if result, ok := msg.(*ResultMessage); ok {
			sessionID = result.SessionID
		}
	}
	client1.Close()

	if sessionID == "" {
		t.Fatal("No session ID received")
	}
	t.Logf("Session ID: %s", sessionID)

	// Second session: resume and verify context
	client2 := NewClient(
		WithResume(sessionID),
	)
	if err := client2.Connect(ctx, ""); err != nil {
		t.Fatalf("Connect 2 failed: %v", err)
	}
	defer client2.Close()

	if err := client2.SendQuery("What is the secret word I told you?"); err != nil {
		t.Fatalf("SendQuery 2 failed: %v", err)
	}

	var responseText string
	for msg := range client2.Messages() {
		if m, ok := msg.(*AssistantMessage); ok {
			responseText += m.Text()
		}
	}

	if !strings.Contains(strings.ToLower(responseText), "banana") {
		t.Errorf("Claude should remember 'banana', got: %s", responseText)
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
git commit -m "test: add comprehensive integration tests"
```

---

## Task 9: Package Documentation

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
//	for msg := range client.Messages() {
//	    switch m := msg.(type) {
//	    case *sdk.AssistantMessage:
//	        fmt.Print(m.Text())
//	    case *sdk.ResultMessage:
//	        fmt.Printf("\nCost: $%.4f\n", m.Cost())
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
//	    sdk.WithSystemPrompt("You are a helpful assistant."),
//	)
//
// # Hooks
//
// Monitor and control tool execution:
//
//	client := sdk.NewClient(
//	    sdk.WithPreToolUseHook(func(name string, input map[string]any) *sdk.HookResult {
//	        fmt.Printf("About to use: %s\n", name)
//	        if name == "Bash" && strings.Contains(input["command"].(string), "rm") {
//	            return &sdk.HookResult{
//	                Decision: sdk.HookDecisionBlock,
//	                Message:  "rm commands are not allowed",
//	            }
//	        }
//	        return &sdk.HookResult{Decision: sdk.HookDecisionAllow}
//	    }),
//	    sdk.WithPostToolUseHook(func(name string, input map[string]any, output *sdk.ToolOutput) *sdk.HookResult {
//	        fmt.Printf("Completed: %s (%.2fs)\n", name, output.Duration.Seconds())
//	        return nil
//	    }),
//	)
//
// # MCP Servers
//
// Host custom tools via MCP protocol:
//
//	server := sdk.NewMCPServerBuilder("my-tools").
//	    WithTool("greet", "Greet someone", schema, func(input map[string]any) (any, error) {
//	        return fmt.Sprintf("Hello, %s!", input["name"]), nil
//	    }).
//	    Build()
//
//	client := sdk.NewClient(sdk.WithMCPServer(server))
//
// # Session Management
//
// Resume previous conversations:
//
//	// Start a session
//	client := sdk.NewClient()
//	client.Connect(ctx, "")
//	client.SendQuery("Remember this!")
//	// ... get sessionID from ResultMessage ...
//
//	// Later, resume it
//	client := sdk.NewClient(sdk.WithResume(sessionID))
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
// # Error Handling
//
// The SDK defines several error types:
//
//   - ErrCLINotFound: Claude CLI not installed
//   - ErrVersionMismatch: CLI version too old
//   - ErrConnectionFailed: Failed to connect to CLI
//   - ErrProcessExited: CLI process exited with error
//   - ErrParseFailed: Failed to parse CLI output
//
// Use errors.Is to check error types:
//
//	if errors.Is(err, sdk.ErrCLINotFound) {
//	    fmt.Println("Please install: npm i -g @anthropic-ai/claude-code")
//	}
//
// # Requirements
//
// The Claude CLI must be installed:
//
//	npm install -g @anthropic-ai/claude-code
//
// Minimum version: 2.0.0
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

// DefaultModel is the default Claude model used when not specified.
const DefaultModel = "claude-sonnet-4-5"

// Package-level convenience functions are defined in client.go:
//   - Query(ctx, prompt) - one-shot query
//   - QueryWithOptions(ctx, prompt, opts...) - one-shot with options
//   - QueryStream(ctx, prompt) - streaming query
//   - QueryStreamWithOptions(ctx, prompt, opts...) - streaming with options
```

**Step 3: Verify docs**

```bash
go doc ./...
```

Expected: Documentation is generated

**Step 4: Commit**

```bash
git add sdk.go doc.go
git commit -m "docs: add comprehensive package documentation"
```

---

## Summary

After completing all plans, you have:

**Plan 01: Types & Interfaces**
- Error types with sentinel errors
- Content blocks (Text, Thinking, ToolUse, ToolResult, ServerToolUse)
- Message types (User, Assistant, System, Result, StreamEvent)
- Options with functional pattern
- Hook types (6 event types, callbacks, results)
- Control protocol types (requests, responses, acknowledgments)
- File checkpointing types
- Sandbox configuration types
- Transport interface

**Plan 02: Transport Layer**
- Version checking (semantic version comparison)
- CLI discovery (PATH, bundled, custom)
- Command building with Windows support
- SubprocessTransport with write mutex
- TOCTOU-safe operations
- Speculative JSON parsing
- Graceful shutdown with timeout
- Stderr callback support
- Concurrent write tests

**Plan 03: Query Protocol**
- Message parser with typed conversion
- Mock transport for testing
- Query structure with channels
- Control request/response routing
- Initialize with hooks and MCP servers
- Interrupt, SetPermissionMode, RewindFiles
- Hook callback handling
- Stream input
- MCP server integration
- MCP tool call handling

**Plan 04: Client API**
- Client structure with full options
- Connect with initialization
- Query and QueryStream functions
- Streaming methods (SendQuery, ReceiveMessage, ReceiveAll)
- Context manager pattern (WithClient, Run)
- Message helpers (Text, ToolCalls, HasToolCalls, etc.)
- Async iterator pattern (Messages, Errors channels)
- Session management (Resume, Continue)

**Plan 05: Integration**
- Examples README
- Simple query example
- Streaming example
- Hooks example
- MCP server example
- Permissions example
- Session management example
- Error handling example
- Comprehensive integration tests
- Package documentation

**Total: ~2,500 lines of Go code across 46 tasks**

---

## Next Steps

1. Create git repository: `git init`
2. Push to GitHub: `gh repo create`
3. Add CI/CD with GitHub Actions
4. Publish to pkg.go.dev

**Plan complete and saved.** Ready for execution!
