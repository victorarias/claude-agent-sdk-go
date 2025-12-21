// Example: Custom agent definitions with Claude Agent SDK.
//
// This demonstrates how to define specialized agents:
// - Define agents with specific tools and prompts
// - Configure agents for different use cases
// - Use agent presets for common patterns
//
// Usage:
//
//	go run examples/agents/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	sdk "github.com/victorarias/claude-agent-sdk-go"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Custom Agents Example")
	fmt.Println("=====================")
	fmt.Println()

	// Define custom agents
	agents := map[string]sdk.AgentDefinition{
		// Code reviewer agent - focused on reviewing code quality
		"code-reviewer": {
			Description: "Code review agent focused on quality and best practices",
			Prompt: `You are a code reviewer. Focus on:
- Code quality and best practices
- Potential bugs and security issues
- Performance concerns
- Readability and maintainability
Be constructive and specific in your feedback.`,
			Tools: []string{"Read", "Glob", "Grep"},
		},

		// Data analyst agent - focused on data analysis
		"data-analyst": {
			Description: "Data analysis agent for exploring and analyzing data",
			Prompt: `You are a data analyst. You help users:
- Analyze data files (CSV, JSON)
- Generate insights and summaries
- Create simple visualizations (ASCII charts)
- Explain statistical concepts
Be precise with numbers and clear about methodology.`,
			Tools: []string{"Read", "Bash"},
		},

		// Documentation agent - focused on writing docs
		"doc-writer": {
			Description: "Technical documentation writer",
			Prompt: `You are a technical documentation writer. You:
- Write clear, concise documentation
- Follow standard documentation formats
- Include examples where helpful
- Use proper markdown formatting
Focus on accuracy and clarity.`,
			Tools: []string{"Read", "Write", "Glob"},
		},
	}

	// Create client with custom agents
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),
		sdk.WithAgents(agents),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Custom agents defined:")
	for name, agent := range agents {
		fmt.Printf("  - %s: %s\n", name, agent.Description)
	}
	fmt.Println()

	// Example query that could use the code reviewer agent
	fmt.Println("--- Query ---")
	prompt := "Review the go.mod file for any issues or improvements."
	fmt.Printf("Q: %s\n\n", prompt)

	if err := client.SendQuery(prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("A: ")
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
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
			return
		}
	}
}
