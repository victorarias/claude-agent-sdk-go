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

	// Track which tools have been permanently allowed
	alwaysAllowed := make(map[string]bool)
	denyAll := false

	// Create client with custom permission callback
	client := sdk.NewClient(
		sdk.WithModel("claude-sonnet-4-5"),

		// Permission callback: called for every tool that needs approval
		sdk.WithCanUseTool(func(toolName string, input map[string]any, permCtx *sdk.ToolPermissionContext) (sdk.PermissionResult, error) {
			// If user has denied all, reject immediately
			if denyAll {
				return &sdk.PermissionResultDeny{
					Behavior: "deny",
					Message:  "All tool requests denied by user",
				}, nil
			}

			// If tool is always allowed, permit immediately
			if alwaysAllowed[toolName] {
				fmt.Printf("[Auto-allowed: %s]\n", toolName)
				return &sdk.PermissionResultAllow{Behavior: "allow"}, nil
			}

			fmt.Println()
			fmt.Println("+-------------------------------------------+")
			fmt.Println("|         TOOL PERMISSION REQUEST           |")
			fmt.Println("+-------------------------------------------+")
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
				if newStr, ok := input["new_string"].(string); ok {
					fmt.Printf("With: %s\n", truncate(newStr, 50))
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
				return &sdk.PermissionResultDeny{
					Behavior: "deny",
					Message:  "Input cancelled",
				}, nil
			}

			response := strings.ToLower(strings.TrimSpace(scanner.Text()))
			switch response {
			case "y", "yes", "":
				fmt.Println("-> Allowed (this time)")
				return &sdk.PermissionResultAllow{Behavior: "allow"}, nil

			case "n", "no":
				fmt.Println("-> Denied")
				return &sdk.PermissionResultDeny{
					Behavior: "deny",
					Message:  "User denied permission",
				}, nil

			case "a", "always":
				fmt.Printf("-> Always allow %s\n", toolName)
				alwaysAllowed[toolName] = true
				return &sdk.PermissionResultAllow{Behavior: "allow"}, nil

			case "d", "deny":
				fmt.Println("-> Deny all future requests")
				denyAll = true
				return &sdk.PermissionResultDeny{
					Behavior: "deny",
					Message:  "User denied all permissions",
				}, nil

			default:
				fmt.Println("-> Defaulting to deny")
				return &sdk.PermissionResultDeny{
					Behavior: "deny",
					Message:  "Invalid response",
				}, nil
			}
		}),
	)

	// Connect
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("Permission Example - Interactive tool approval")
	fmt.Println("-----------------------------------------------")
	fmt.Println()
	fmt.Println("Claude will ask permission before using any tool.")
	fmt.Println("You can approve, deny, or set blanket permissions.")
	fmt.Println()

	// Send a query that will trigger multiple tool uses
	prompt := `Please do the following:
1. List the current directory
2. Read the go.mod file
3. Tell me what this project is about`

	fmt.Printf("Prompt:\n%s\n", prompt)
	fmt.Println()

	if err := client.SendQuery(prompt); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	fmt.Println("\n---- Response ----")
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

func truncate(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}
