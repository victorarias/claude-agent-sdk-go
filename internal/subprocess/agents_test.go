// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"strings"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// Agents are now sent via initialize control request over stdin.
// buildCommand should not include --agents CLI args.
func TestBuildCommand_AgentsNotPassedAsCLIArg(t *testing.T) {
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
	for _, arg := range cmd {
		if arg == "--agents" || strings.HasPrefix(arg, "@") {
			t.Fatalf("agents should not be passed as CLI args: %v", cmd)
		}
	}
}

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

func TestAgentsTempFileCleanup(t *testing.T) {
	transport := NewSubprocessTransport("Hello", types.DefaultOptions())

	// Simulate adding temp files
	transport.AddTempFile("/tmp/fake-agents-1.json")
	transport.AddTempFile("/tmp/fake-agents-2.json")

	transport.tempMu.Lock()
	tempFileCount := len(transport.tempFiles)
	transport.tempMu.Unlock()

	if tempFileCount != 2 {
		t.Errorf("Expected 2 temp files, got %d", tempFileCount)
	}

	transport.Close()

	transport.tempMu.Lock()
	tempFileCount = len(transport.tempFiles)
	transport.tempMu.Unlock()

	if tempFileCount != 0 {
		t.Errorf("Expected 0 temp files after Close, got %d", tempFileCount)
	}
}
