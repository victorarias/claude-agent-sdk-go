package sdk

import (
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.PermissionMode != "" {
		t.Errorf("expected empty permission mode, got %q", opts.PermissionMode)
	}
}

func TestOptionsWithModel(t *testing.T) {
	opts := DefaultOptions()
	WithModel("claude-opus-4")(opts)
	if opts.Model != "claude-opus-4" {
		t.Errorf("got %q, want %q", opts.Model, "claude-opus-4")
	}
}

func TestOptionsWithCwd(t *testing.T) {
	opts := DefaultOptions()
	WithCwd("/tmp/test")(opts)
	if opts.Cwd != "/tmp/test" {
		t.Errorf("got %q, want %q", opts.Cwd, "/tmp/test")
	}
}

func TestOptionsWithPermissionMode(t *testing.T) {
	opts := DefaultOptions()
	WithPermissionMode(PermissionBypass)(opts)
	if opts.PermissionMode != PermissionBypass {
		t.Errorf("got %q, want %q", opts.PermissionMode, PermissionBypass)
	}
}

func TestOptionsWithEnv(t *testing.T) {
	opts := DefaultOptions()
	WithEnv(map[string]string{"FOO": "bar"})(opts)
	if opts.Env["FOO"] != "bar" {
		t.Errorf("got %q, want %q", opts.Env["FOO"], "bar")
	}
}

func TestOptionsWithSandbox(t *testing.T) {
	opts := DefaultOptions()
	WithSandbox(SandboxSettings{Enabled: true})(opts)
	if opts.Sandbox == nil || !opts.Sandbox.Enabled {
		t.Error("expected sandbox to be enabled")
	}
}

func TestOptionsWithAgents(t *testing.T) {
	opts := DefaultOptions()
	WithAgents(map[string]AgentDefinition{
		"test": {Description: "Test agent", Prompt: "You are a test"},
	})(opts)
	if opts.Agents["test"].Description != "Test agent" {
		t.Error("expected agent to be set")
	}
}

func TestOptionsWithMaxTurns(t *testing.T) {
	opts := DefaultOptions()
	WithMaxTurns(10)(opts)
	if opts.MaxTurns != 10 {
		t.Errorf("got %d, want 10", opts.MaxTurns)
	}
}

func TestOptionsWithMaxBudget(t *testing.T) {
	opts := DefaultOptions()
	WithMaxBudget(5.0)(opts)
	if opts.MaxBudgetUSD != 5.0 {
		t.Errorf("got %f, want 5.0", opts.MaxBudgetUSD)
	}
}

func TestApplyOptions(t *testing.T) {
	opts := DefaultOptions()
	ApplyOptions(opts,
		WithModel("claude-opus-4"),
		WithCwd("/tmp"),
		WithMaxTurns(5),
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
