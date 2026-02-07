// Example: Unstable v2 session APIs.
//
// This demonstrates how to:
// - Create an unstable v2 session
// - Send and stream messages
// - Run one-shot prompt helper
//
// Usage:
//
//	go run examples/unstable-session/main.go
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

	session, err := sdk.UnstableV2CreateSession(
		ctx,
		types.WithModel("claude-sonnet-4-5"),
		types.WithMaxTurns(3),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create session: %v\n", err)
		os.Exit(1)
	}
	defer session.Close()

	if err := session.Send("Say hello in one short sentence."); err != nil {
		fmt.Fprintf(os.Stderr, "failed to send message: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Claude: ")
	for msg := range session.Stream() {
		switch m := msg.(type) {
		case *types.AssistantMessage:
			fmt.Print(m.Text())
		case *types.ResultMessage:
			fmt.Println()
			goto oneshot
		}
	}

oneshot:
	result, err := sdk.UnstableV2Prompt(ctx, "Reply with the single word: done", types.WithModel("claude-sonnet-4-5"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "unstable one-shot failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[one-shot result subtype: %s]\n", result.Subtype)
}
