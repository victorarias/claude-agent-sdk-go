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

	"github.com/victorarias/claude-agent-sdk-go/sdk"
	"github.com/victorarias/claude-agent-sdk-go/types"
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
		types.WithModel("claude-sonnet-4-5"),
	)

	if err := client.Connect(ctx); err != nil {
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
			sessionID = m.SessionID
			fmt.Println()
			goto done
		}
	}
done:

	// Save session ID
	if sessionID != "" {
		if err := saveSessionID(sessionID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: couldn't save session: %v\n", err)
		} else {
			truncatedID := sessionID
			if len(sessionID) > 16 {
				truncatedID = sessionID[:16] + "..."
			}
			fmt.Printf("\n[Session saved: %s]\n", truncatedID)
			fmt.Println("[Run 'session resume' to continue this conversation]")
		}
	} else {
		fmt.Println("\n[No session ID received]")
	}
}

func resumeSession(ctx context.Context) {
	sessionID, err := loadSessionID()
	if err != nil {
		fmt.Fprintf(os.Stderr, "No saved session found: %v\n", err)
		fmt.Println("Run 'session new' first to create a session.")
		os.Exit(1)
	}

	truncatedID := sessionID
	if len(sessionID) > 16 {
		truncatedID = sessionID[:16] + "..."
	}
	fmt.Printf("Resuming session: %s\n", truncatedID)
	fmt.Println()

	// Resume the previous session
	client := sdk.NewClient(
		types.WithModel("claude-sonnet-4-5"),
		types.WithResume(sessionID),
	)

	if err := client.Connect(ctx); err != nil {
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
			truncatedNewID := m.SessionID
			if len(m.SessionID) > 16 {
				truncatedNewID = m.SessionID[:16] + "..."
			}
			fmt.Printf("\n[Session ID: %s]\n", truncatedNewID)
			return
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
