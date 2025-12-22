// Example: Configuring settings sources.
//
// Settings sources control where Claude CLI loads configuration from. You can:
// - Specify which setting sources to use (user, project, local)
// - Provide inline settings as JSON
// - Combine multiple settings sources with priorities
//
// Usage:
//
//	go run examples/settings-sources/main.go
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

	fmt.Println("Settings Sources Examples")
	fmt.Println("========================")
	fmt.Println()

	// Example 1: Using specific setting sources
	// Control which settings files the CLI loads
	fmt.Println("1. Specific Setting Sources")
	fmt.Println("   Using WithSettingSources() to specify where settings come from")
	settingSourcesExample(ctx)

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Example 2: Inline settings
	// Provide settings directly as JSON
	fmt.Println("2. Inline Settings")
	fmt.Println("   Using WithSettings() to provide settings as JSON")
	inlineSettingsExample(ctx)

	fmt.Println()
	fmt.Println("---")
	fmt.Println()

	// Example 3: Combining settings sources
	// Use both file-based and inline settings
	fmt.Println("3. Combined Settings")
	fmt.Println("   Using both WithSettingSources() and WithSettings()")
	combinedSettingsExample(ctx)
}

// settingSourcesExample demonstrates specifying which settings files to load
func settingSourcesExample(ctx context.Context) {
	// Setting sources determine which settings files the CLI loads:
	// - SettingSourceUser: ~/.config/claude/settings.json (user-level)
	// - SettingSourceProject: .claude/settings.json (project-level)
	// - SettingSourceLocal: .claude/local_settings.json (local overrides)
	//
	// By default, all sources are loaded with local > project > user priority.
	// You can override this to use only specific sources.

	// Only use user and project settings (skip local)
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSettingSources(
			types.SettingSourceUser,
			types.SettingSourceProject,
		),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("   Settings loaded from user and project sources only")
	fmt.Println("   (local settings are ignored)")

	// Send a simple query
	if err := client.SendQuery("What settings sources are you using?"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Println("   Response:", m.Text())
		case *types.ResultMessage:
			return
		}
	}
}

// inlineSettingsExample demonstrates providing settings as JSON
func inlineSettingsExample(ctx context.Context) {
	// Instead of loading from files, you can provide settings directly as JSON.
	// This is useful for:
	// - Testing with specific configurations
	// - Programmatic configuration
	// - Overriding file-based settings

	// Define settings as JSON string
	settingsJSON := `{
		"models": {
			"default": "claude-sonnet-4-5"
		},
		"tools": {
			"allowed": ["Bash", "Read", "Write"]
		}
	}`

	// Create client with inline settings
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithSettings(settingsJSON),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("   Settings provided as inline JSON")
	fmt.Println("   (no settings files are loaded)")

	// Send a simple query
	if err := client.SendQuery("List the available tools"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Println("   Response:", m.Text())
		case *types.ResultMessage:
			return
		}
	}
}

// combinedSettingsExample demonstrates using both settings sources and inline settings
func combinedSettingsExample(ctx context.Context) {
	// You can combine settings sources with inline settings.
	// Inline settings typically have highest priority and override file-based settings.

	// Define inline settings
	settingsJSON := `{
		"maxTurns": 10,
		"sandbox": {
			"enabled": true
		}
	}`

	// Create client with both settings sources and inline settings
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		// Load from user and project files
		types.WithSettingSources(
			types.SettingSourceUser,
			types.SettingSourceProject,
		),
		// Also apply inline settings (highest priority)
		types.WithSettings(settingsJSON),
		types.WithPermissionMode(types.PermissionBypass),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("   Settings loaded from:")
	fmt.Println("   1. User settings file")
	fmt.Println("   2. Project settings file")
	fmt.Println("   3. Inline JSON (overrides file settings)")

	// Send a simple query
	if err := client.SendQuery("What's your configuration?"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Receive response
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			break
		}

		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Println("   Response:", m.Text())
		case *types.ResultMessage:
			return
		}
	}
}
