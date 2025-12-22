// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package sdk

import (
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestDefaultOptions(t *testing.T) {
	opts := types.DefaultOptions()
	if opts.PermissionMode != "" {
		t.Errorf("expected empty permission mode, got %q", opts.PermissionMode)
	}
}

func TestOptionsWithModel(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithModel("claude-opus-4")(opts)
	if opts.Model != "claude-opus-4" {
		t.Errorf("got %q, want %q", opts.Model, "claude-opus-4")
	}
}

func TestOptionsWithCwd(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithCwd("/tmp/test")(opts)
	if opts.Cwd != "/tmp/test" {
		t.Errorf("got %q, want %q", opts.Cwd, "/tmp/test")
	}
}

func TestOptionsWithPermissionMode(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithPermissionMode(types.PermissionBypass)(opts)
	if opts.PermissionMode != types.PermissionBypass {
		t.Errorf("got %q, want %q", opts.PermissionMode, types.PermissionBypass)
	}
}

func TestOptionsWithEnv(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithEnv(map[string]string{"FOO": "bar"})(opts)
	if opts.Env["FOO"] != "bar" {
		t.Errorf("got %q, want %q", opts.Env["FOO"], "bar")
	}
}

func TestOptionsWithSandbox(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithSandbox(types.SandboxSettings{Enabled: true})(opts)
	if opts.Sandbox == nil || !opts.Sandbox.Enabled {
		t.Error("expected sandbox to be enabled")
	}
}

func TestOptionsWithAgents(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithAgents(map[string]types.AgentDefinition{
		"test": {Description: "Test agent", Prompt: "You are a test"},
	})(opts)
	if opts.Agents["test"].Description != "Test agent" {
		t.Error("expected agent to be set")
	}
}

func TestOptionsWithMaxTurns(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithMaxTurns(10)(opts)
	if opts.MaxTurns != 10 {
		t.Errorf("got %d, want 10", opts.MaxTurns)
	}
}

func TestOptionsWithMaxBudget(t *testing.T) {
	opts := types.DefaultOptions()
	types.WithMaxBudget(5.0)(opts)
	if opts.MaxBudgetUSD != 5.0 {
		t.Errorf("got %f, want 5.0", opts.MaxBudgetUSD)
	}
}

func TestApplyOptions(t *testing.T) {
	opts := types.DefaultOptions()
	types.ApplyOptions(opts,
		types.WithModel("claude-opus-4"),
		types.WithCwd("/tmp"),
		types.WithMaxTurns(5),
	)
	if opts.Model != "claude-opus-4" {
		t.Errorf("model: got %q, want %q", opts.Model, "claude-opus-4")
	}
	if opts.Cwd != "/tmp" {
		t.Errorf("cwd: got %q, want %q", opts.Cwd, "/tmp")
	}
	if opts.MaxTurns != 5 {
		t.Errorf("maxTurns: got %d, want 5", opts.MaxTurns)
	}
}
