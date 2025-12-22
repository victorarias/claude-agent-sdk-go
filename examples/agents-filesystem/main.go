// Example: Filesystem-based agents with Claude Agent SDK.
//
// This demonstrates how to use filesystem-based agents loaded from .claude/agents/:
// - Load agents from .claude/agents/ directory using setting_sources
// - Configure directory access constraints
// - Show multi-agent workflow patterns
// - Combine filesystem agents with inline agent definitions
//
// Filesystem agents are defined in .claude/agents/ as markdown files with frontmatter.
// This approach allows sharing agent configurations across sessions and teams.
//
// Usage:
//
//	go run examples/agents-filesystem/main.go
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("Filesystem-Based Agents Example")
	fmt.Println("================================")
	fmt.Println()

	// Get the project root directory (where .claude/agents/ is located)
	projectRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get working directory: %v\n", err)
		os.Exit(1)
	}

	// Define the directories we want to allow the agent to access
	// In this example, we'll restrict access to examples and types directories
	examplesDir := filepath.Join(projectRoot, "examples")
	typesDir := filepath.Join(projectRoot, "types")

	fmt.Println("Configuration:")
	fmt.Printf("  Project root: %s\n", projectRoot)
	fmt.Printf("  Allowed directories:\n")
	fmt.Printf("    - %s\n", examplesDir)
	fmt.Printf("    - %s\n", typesDir)
	fmt.Println()

	// Define inline agents to complement filesystem-based ones
	inlineAgents := map[string]types.AgentDefinition{
		// Security auditor with read-only access
		"security-auditor": {
			Description: "Security analysis agent with read-only filesystem access",
			Prompt: `You are a security auditor. Review code for:
- Security vulnerabilities and potential exploits
- Unsafe patterns and anti-patterns
- Hardcoded secrets or sensitive data
- Input validation issues
Be thorough but constructive in your findings.`,
			Tools: []string{"Read", "Grep", "Glob"},
			Model: types.AgentModelSonnet,
		},

		// Code organizer with write access
		"code-organizer": {
			Description: "Code organization agent that can modify files",
			Prompt: `You are a code organization expert. Help with:
- Refactoring and restructuring code
- Improving file and package organization
- Applying consistent coding patterns
- Documenting code structure
Always explain your changes clearly.`,
			Tools: []string{"Read", "Write", "Edit", "Glob", "Grep"},
			Model: types.AgentModelSonnet,
		},
	}

	// Create client with both filesystem-based and inline agents
	// The "project" setting source loads agents from .claude/agents/
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithCwd(projectRoot),
		types.WithSettingSources(types.SettingSourceProject), // Load from .claude/agents/
		types.WithAgents(inlineAgents),                       // Add inline agents
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Available agents:")
	fmt.Println("  Filesystem-based (from .claude/agents/):")
	fmt.Println("    - file-analyzer: File system analysis with restricted access")
	fmt.Println()
	fmt.Println("  Inline agents:")
	for name, agent := range inlineAgents {
		fmt.Printf("    - %s: %s\n", name, agent.Description)
	}
	fmt.Println()

	// Example 1: Use filesystem-based agent to analyze directory structure
	fmt.Println("=== Example 1: Analyzing Directory Structure ===")
	fmt.Println()
	prompt1 := "Use the file-analyzer agent to list all Go files in the examples directory and describe the examples you find."
	fmt.Printf("Q: %s\n\n", prompt1)

	if err := client.SendQuery(prompt1); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("A: ")
	var totalCost float64
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			fmt.Println()
			if m.TotalCostUSD != nil {
				totalCost = *m.TotalCostUSD
				fmt.Printf("\n[Cost: $%.4f]\n", totalCost)
			}
			goto example2
		}
	}

example2:
	fmt.Println()
	fmt.Println("=== Example 2: Multi-Agent Workflow ===")
	fmt.Println()
	prompt2 := "First use the security-auditor agent to check the types/options.go file for any security concerns, then use the file-analyzer to provide a summary of the file structure."
	fmt.Printf("Q: %s\n\n", prompt2)

	if err := client.SendQuery(prompt2); err != nil {
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
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			fmt.Println()
			if m.TotalCostUSD != nil {
				totalCost += *m.TotalCostUSD
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
			goto done
		}
	}

done:
	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("Total cost: $%.4f\n", totalCost)
	fmt.Println()
	fmt.Println("This example demonstrated:")
	fmt.Println("  1. Loading agents from .claude/agents/ directory")
	fmt.Println("  2. Combining filesystem-based and inline agent definitions")
	fmt.Println("  3. Using multiple agents in a workflow")
	fmt.Println("  4. Restricting agent access to specific directories")
}
