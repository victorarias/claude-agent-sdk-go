// Example: Loading Claude Code plugins with the SDK.
//
// This example demonstrates how to load local Claude Code plugins programmatically.
// Plugins allow you to extend Claude Code with custom commands, agents, skills,
// and hooks. The SDK supports loading local plugins by specifying their directory path.
//
// Note: This example requires a plugin directory structure. For demonstration purposes,
// you can create a minimal plugin or reference the Python SDK's demo plugin structure.
//
// Plugin Structure:
//
//	demo-plugin/
//	├── .claude-plugin/
//	│   └── manifest.json
//	└── commands/
//	    └── greet.js (or other command files)
//
// Usage:
//
//	go run examples/plugin/main.go
//	go run examples/plugin/main.go /path/to/plugin
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
)

func main() {
	// Determine plugin path
	// Default to a demo plugin if one exists, or allow user to specify
	pluginPath := getPluginPath()

	fmt.Println("=== Plugin Example ===\n")
	fmt.Printf("Loading plugin from: %s\n\n", pluginPath)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Configure options with plugin
	options := &types.Options{
		// Load a local plugin by path
		Plugins: []types.Plugin{
			{
				Type: "local",
				Path: pluginPath,
			},
		},
		// Limit to one turn for quick demo
		MaxTurns: intPtr(1),
	}

	fmt.Println("Checking for loaded plugins in system initialization...")

	// Create a client to check if the plugin loaded successfully
	client, err := sdk.NewClient(ctx, options)
	if err != nil {
		handleError(err)
		os.Exit(1)
	}
	defer client.Close()

	// Track if we found plugin information
	foundPluginInfo := false

	// Send a simple query and check the system message for plugin info
	messages, err := client.RunQuery(ctx, "Hello!")
	if err != nil {
		handleError(err)
		os.Exit(1)
	}

	// Look for plugin information in messages
	for _, msg := range messages {
		switch m := msg.(type) {
		case *types.SystemMessage:
			if m.Subtype == "init" {
				fmt.Println("System initialized!")

				// Check for plugins in the system message data
				if pluginsData, ok := m.Data["plugins"]; ok {
					if plugins, ok := pluginsData.([]any); ok && len(plugins) > 0 {
						fmt.Println("\nPlugins loaded:")
						for _, p := range plugins {
							if plugin, ok := p.(map[string]any); ok {
								name := plugin["name"]
								path := plugin["path"]
								fmt.Printf("  - %v (path: %v)\n", name, path)
							}
						}
						foundPluginInfo = true
					}
				}

				if !foundPluginInfo {
					// Plugin might be loaded but not reported in system message
					fmt.Println("\nNote: Plugin was configured via options.")
					fmt.Printf("Plugin path: %s\n", pluginPath)
					fmt.Println("The plugin may be loaded but not visible in system messages.")
					fmt.Println("Check Claude Code documentation for plugin visibility details.")
				}
			}

		case *types.AssistantMessage:
			// Print the response
			if text := m.Text(); text != "" {
				fmt.Printf("\nClaude's response:\n%s\n", text)
			}

		case *types.ResultMessage:
			if m.TotalCostUSD != nil {
				fmt.Printf("\n[Cost: $%.4f]\n", *m.TotalCostUSD)
			}
		}
	}

	if foundPluginInfo {
		fmt.Println("\nPlugin successfully loaded!")
	}

	fmt.Println("\nExample completed successfully!")
}

// getPluginPath determines the plugin path to use
func getPluginPath() string {
	// If user provided a path as argument, use it
	if len(os.Args) > 1 {
		return os.Args[1]
	}

	// Try to find the demo plugin from the Python reference
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not determine executable path")
		return "./demo-plugin" // Fallback
	}

	// Navigate from examples/plugin to reference/examples/plugins/demo-plugin
	repoRoot := filepath.Join(filepath.Dir(exePath), "..", "..")
	demoPlugin := filepath.Join(repoRoot, "reference", "examples", "plugins", "demo-plugin")

	// Check if it exists
	if info, err := os.Stat(demoPlugin); err == nil && info.IsDir() {
		return demoPlugin
	}

	// Fallback to current directory
	fmt.Println("Note: Demo plugin not found in reference/examples/plugins/demo-plugin")
	fmt.Println("Using fallback path. You can specify a plugin path as an argument.")
	return "./demo-plugin"
}

// handleError demonstrates proper error handling patterns
func handleError(err error) {
	switch {
	case errors.Is(err, types.ErrCLINotFound):
		fmt.Fprintln(os.Stderr, "Error: Claude CLI not found")
		fmt.Fprintln(os.Stderr, "Install with: npm install -g @anthropic-ai/claude-code")

	case errors.Is(err, types.ErrCLIVersion):
		fmt.Fprintln(os.Stderr, "Error: Claude CLI version too old")
		fmt.Fprintln(os.Stderr, "Update with: npm update -g @anthropic-ai/claude-code")

	case errors.Is(err, context.DeadlineExceeded):
		fmt.Fprintln(os.Stderr, "Error: Request timed out")

	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

// intPtr is a helper to create an int pointer
func intPtr(i int) *int {
	return &i
}
