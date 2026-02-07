// Example: Runtime control APIs for model, permissions, sessions, and MCP.
//
// This demonstrates how to:
// - Read initialization metadata (models, commands, account)
// - Change model/permission mode during a live session
// - Set and clear max thinking tokens
// - Inspect and manage MCP servers
// - Use rewind-files dry-run with a captured user message UUID
//
// Usage:
//
//	go run examples/runtime-controls/main.go
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

	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithPermissionMode(types.PermissionDefault),
	)

	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	initMeta, err := client.InitializationResult()
	if err == nil {
		fmt.Printf("Output style: %s\n", initMeta.OutputStyle)
	}

	models, err := client.SupportedModels()
	if err == nil && len(models) > 0 {
		fmt.Printf("Models available: %d (first: %s)\n", len(models), models[0].Value)
	}

	commands, err := client.SupportedCommands()
	if err == nil {
		fmt.Printf("Slash commands available: %d\n", len(commands))
	}

	account, err := client.AccountInfo()
	if err == nil && account != nil && account.Email != "" {
		fmt.Printf("Authenticated as: %s\n", account.Email)
	}

	if err := client.SetPermissionMode(types.PermissionDefault); err != nil {
		fmt.Fprintf(os.Stderr, "set permission mode failed: %v\n", err)
	}

	if err := client.SetModel("claude-sonnet-4-5"); err != nil {
		fmt.Fprintf(os.Stderr, "set model failed: %v\n", err)
	}

	maxThinking := 1024
	if err := client.SetMaxThinkingTokens(&maxThinking); err != nil {
		fmt.Fprintf(os.Stderr, "set max thinking tokens failed: %v\n", err)
	}
	if err := client.SetMaxThinkingTokens(nil); err != nil {
		fmt.Fprintf(os.Stderr, "clear max thinking tokens failed: %v\n", err)
	}

	statuses, err := client.MCPServerStatus()
	if err == nil {
		fmt.Printf("MCP servers visible: %d\n", len(statuses))
		if len(statuses) > 0 {
			name := statuses[0].Name
			if err := client.ToggleMCPServer(name, true); err != nil {
				fmt.Fprintf(os.Stderr, "toggle MCP server failed: %v\n", err)
			}
			if err := client.ReconnectMCPServer(name); err != nil {
				fmt.Fprintf(os.Stderr, "reconnect MCP server failed: %v\n", err)
			}
		}
	}

	if _, err := client.SetMCPServers(map[string]any{}); err != nil {
		fmt.Fprintf(os.Stderr, "set MCP servers failed: %v\n", err)
	}

	if err := client.SendQuery("Say hello in one short sentence."); err != nil {
		fmt.Fprintf(os.Stderr, "send query failed: %v\n", err)
		os.Exit(1)
	}

	var firstUserMessageID string
	for {
		msg, err := client.ReceiveMessage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "receive message failed: %v\n", err)
			os.Exit(1)
		}

		switch m := msg.(type) {
		case *types.UserMessage:
			if firstUserMessageID == "" {
				firstUserMessageID = m.UUID
			}
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			fmt.Println()
			goto rewind
		}
	}

rewind:
	if firstUserMessageID == "" {
		fmt.Println("No user message UUID available for rewind demo.")
		return
	}
	dryRun := true
	rewindResult, err := client.RewindFilesWithOptions(firstUserMessageID, &dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "rewind files dry-run failed: %v\n", err)
		return
	}
	fmt.Printf("Rewind dry-run: canRewind=%t filesChanged=%d\n", rewindResult.CanRewind, len(rewindResult.FilesChanged))
}
