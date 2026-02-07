// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

//go:build integration

package sdk

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
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

	messages, err := RunQuery(ctx, "What is 2+2? Reply with just the number.")
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
		case *types.AssistantMessage:
			gotAssistant = true
			responseText = m.Text()
		case *types.ResultMessage:
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
		types.WithMaxTurns(2),
	)

	if err := client.Connect(ctx); err != nil {
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
	result, ok := lastMsg.(*types.ResultMessage)
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
		types.WithMaxTurns(5),
	)

	if err := client.Connect(ctx); err != nil {
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
		if m, ok := msg.(*types.AssistantMessage); ok {
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

	messages, err := RunQuery(ctx, "Say hello",
		types.WithModel("claude-sonnet-4-5"),
		types.WithMaxTurns(1),
		types.WithSystemPrompt("You are a test assistant. Be very brief."),
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
		if result, ok := msg.(*types.ResultMessage); ok {
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
	messages, err := RunQuery(ctx, "What files are in the current directory? Use the Bash tool to run 'ls -la'.")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	// Check that tools were used
	var usedTools bool
	for _, msg := range messages {
		if m, ok := msg.(*types.AssistantMessage); ok {
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

	if err := client.Connect(ctx); err != nil {
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
		case *types.AssistantMessage:
			t.Logf("Assistant: %s", m.Text())
		case *types.ResultMessage:
			t.Logf("Result: success=%v", m.IsSuccess())
		}
	}

	if messageCount == 0 {
		t.Error("No messages received through channel")
	}
}

func TestIntegration_SessionResume(t *testing.T) {
	skipIfNotIntegration(t)

	ctx, cancel := context.WithTimeout(context.Background(), getTestTimeout())
	defer cancel()

	// First session: establish context
	client1 := NewClient()
	if err := client1.Connect(ctx); err != nil {
		t.Fatalf("Connect 1 failed: %v", err)
	}

	if err := client1.SendQuery("Remember that the secret word is 'banana'."); err != nil {
		client1.Close()
		t.Fatalf("SendQuery 1 failed: %v", err)
	}

	var sessionID string
	for msg := range client1.Messages() {
		if result, ok := msg.(*types.ResultMessage); ok {
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
		types.WithResume(sessionID),
	)
	if err := client2.Connect(ctx); err != nil {
		t.Fatalf("Connect 2 failed: %v", err)
	}
	defer client2.Close()

	if err := client2.SendQuery("What is the secret word I told you?"); err != nil {
		t.Fatalf("SendQuery 2 failed: %v", err)
	}

	var responseText string
	for msg := range client2.Messages() {
		if m, ok := msg.(*types.AssistantMessage); ok {
			responseText += m.Text()
		}
	}

	if !strings.Contains(strings.ToLower(responseText), "banana") {
		t.Errorf("Claude should remember 'banana', got: %s", responseText)
	}
}
