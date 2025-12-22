// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestBuildCommand_Agents tests that agents are passed via --agents flag.
func TestBuildCommand_Agents(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Agents = map[string]types.AgentDefinition{
		"researcher": {
			Description: "Research assistant",
			Prompt:      "You are a research assistant.",
			Tools:       []string{"WebSearch", "Read"},
			Model:       types.AgentModelSonnet,
		},
	}

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	// Find --agents flag
	found := false
	for i := 0; i < len(cmd)-1; i++ {
		if cmd[i] == "--agents" {
			found = true
			// Next arg should be JSON
			agentsJSON := cmd[i+1]

			var agents map[string]any
			if err := json.Unmarshal([]byte(agentsJSON), &agents); err != nil {
				t.Fatalf("Failed to parse agents JSON: %v", err)
			}

			researcher, ok := agents["researcher"].(map[string]any)
			if !ok {
				t.Error("researcher should be in agents JSON")
			}

			if researcher["description"] != "Research assistant" {
				t.Errorf("Expected 'Research assistant', got %q", researcher["description"])
			}

			if researcher["model"] != "sonnet" {
				t.Errorf("Expected 'sonnet', got %q", researcher["model"])
			}

			break
		}
	}

	if !found {
		t.Error("--agents flag not found in command")
	}
}

// TestBuildCommand_AgentsJSONFormat tests the exact JSON format.
func TestBuildCommand_AgentsJSONFormat(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Agents = map[string]types.AgentDefinition{
		"test": {
			Description: "Test",
			Prompt:      "Test prompt",
			Tools:       []string{"Tool1"},
			Model:       types.AgentModelOpus,
		},
	}

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	var agentsJSON string
	for i := 0; i < len(cmd)-1; i++ {
		if cmd[i] == "--agents" {
			agentsJSON = cmd[i+1]
			break
		}
	}

	if agentsJSON == "" {
		t.Fatal("--agents flag not found")
	}

	// Verify it's valid JSON
	var parsed map[string]map[string]any
	if err := json.Unmarshal([]byte(agentsJSON), &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	testAgent := parsed["test"]
	if testAgent["description"] != "Test" {
		t.Error("description mismatch")
	}

	if testAgent["prompt"] != "Test prompt" {
		t.Error("prompt mismatch")
	}
}

// TestBuildCommand_NoAgents tests that no --agents flag is added when agents is nil.
func TestBuildCommand_NoAgents(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Agents = nil

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	for _, arg := range cmd {
		if arg == "--agents" {
			t.Error("--agents should not be present when agents is nil")
		}
	}
}

// TestBuildCommand_EmptyAgents tests that no --agents flag is added when agents is empty.
func TestBuildCommand_EmptyAgents(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Agents = map[string]types.AgentDefinition{}

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	for _, arg := range cmd {
		if arg == "--agents" {
			t.Error("--agents should not be present when agents is empty")
		}
	}
}

// TestAgentsTempFileFallback tests that agents use temp file when command is too long.
func TestAgentsTempFileFallback(t *testing.T) {
	// Create a large agents map to exceed command length limit
	agents := make(map[string]types.AgentDefinition)

	// Create enough agents with long descriptions to exceed the limit
	// On Windows, limit is 8000. On Unix, 100000.
	// We'll create a payload that definitely exceeds 8000.
	longDescription := strings.Repeat("A", 1000)
	longPrompt := strings.Repeat("B", 1000)

	for i := 0; i < 10; i++ {
		agentName := strings.Repeat("agent", 100) + string(rune(i))
		agents[agentName] = types.AgentDefinition{
			Description: longDescription,
			Prompt:      longPrompt,
			Tools:       []string{"Tool1", "Tool2", "Tool3"},
			Model:       types.AgentModelSonnet,
		}
	}

	opts := types.DefaultOptions()
	opts.Agents = agents

	transport := NewSubprocessTransport("Hello", opts)

	// Build the command (this should handle the temp file fallback)
	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	cmdLength := getCmdLength(cmd)

	// On Windows, if original command would exceed 8000, it should use temp file
	// On Unix, we won't hit the 100000 limit with this test data
	if runtime.GOOS == "windows" {
		// The command length after temp file optimization should be much shorter
		// Find the --agents flag and check if it starts with @
		for i := 0; i < len(cmd)-1; i++ {
			if cmd[i] == "--agents" {
				agentsArg := cmd[i+1]
				if strings.HasPrefix(agentsArg, "@") {
					// Good - using temp file reference
					t.Logf("Using temp file fallback: %s", agentsArg)
					return
				} else if cmdLength <= 8000 {
					// Command is short enough, no need for temp file
					t.Logf("Command length %d is within limit, no temp file needed", cmdLength)
					return
				} else {
					t.Errorf("Command length %d exceeds limit but not using temp file", cmdLength)
				}
			}
		}
	}

	// Clean up temp files
	transport.Close()
}

// TestAgentsTempFileCleanup tests that temp files are cleaned up on Close.
func TestAgentsTempFileCleanup(t *testing.T) {
	// Create large agents to trigger temp file
	agents := make(map[string]types.AgentDefinition)
	longText := strings.Repeat("X", 2000)

	for i := 0; i < 5; i++ {
		agentName := strings.Repeat("agent", 50) + string(rune(i))
		agents[agentName] = types.AgentDefinition{
			Description: longText,
			Prompt:      longText,
			Model:       types.AgentModelSonnet,
		}
	}

	opts := types.DefaultOptions()
	opts.Agents = agents

	transport := NewSubprocessTransport("Hello", opts)

	// Simulate adding temp files
	transport.AddTempFile("/tmp/fake-agents-1.json")
	transport.AddTempFile("/tmp/fake-agents-2.json")

	// Verify temp files are tracked
	transport.tempMu.Lock()
	tempFileCount := len(transport.tempFiles)
	transport.tempMu.Unlock()

	if tempFileCount != 2 {
		t.Errorf("Expected 2 temp files, got %d", tempFileCount)
	}

	// Close should clean up temp files
	transport.Close()

	// Verify temp files list is cleared
	transport.tempMu.Lock()
	tempFileCount = len(transport.tempFiles)
	transport.tempMu.Unlock()

	if tempFileCount != 0 {
		t.Errorf("Expected 0 temp files after Close, got %d", tempFileCount)
	}
}

// TestAgentCommandLengthLimit tests the command length validation.
func TestAgentCommandLengthLimit(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Command length limit only applies on Windows")
	}

	// Create a command that exceeds Windows limit
	opts := types.DefaultOptions()

	// Create a large system prompt to push us over the limit
	opts.SystemPrompt = strings.Repeat("A", 10000)

	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	err := checkCommandLength(cmd)
	if err == nil {
		t.Error("Expected error for command exceeding length limit")
	}

	if !strings.Contains(err.Error(), "exceeds Windows limit") {
		t.Errorf("Expected Windows limit error, got: %v", err)
	}
}
