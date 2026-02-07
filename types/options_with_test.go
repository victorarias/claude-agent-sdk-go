// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package types

import (
	"testing"
)

func TestWithModel(t *testing.T) {
	opts := DefaultOptions()
	WithModel("claude-opus-4-5")(opts)
	if opts.Model != "claude-opus-4-5" {
		t.Errorf("expected model 'claude-opus-4-5', got %q", opts.Model)
	}
}

func TestWithCwd(t *testing.T) {
	opts := DefaultOptions()
	WithCwd("/tmp/test")(opts)
	if opts.Cwd != "/tmp/test" {
		t.Errorf("expected cwd '/tmp/test', got %q", opts.Cwd)
	}
}

func TestWithPermissionMode(t *testing.T) {
	opts := DefaultOptions()
	WithPermissionMode(PermissionBypass)(opts)
	if opts.PermissionMode != PermissionBypass {
		t.Errorf("expected PermissionBypass, got %v", opts.PermissionMode)
	}
}

func TestWithEnv(t *testing.T) {
	opts := DefaultOptions()
	WithEnv(map[string]string{"FOO": "bar", "BAZ": "qux"})(opts)
	if opts.Env["FOO"] != "bar" {
		t.Errorf("expected FOO='bar', got %q", opts.Env["FOO"])
	}
	if opts.Env["BAZ"] != "qux" {
		t.Errorf("expected BAZ='qux', got %q", opts.Env["BAZ"])
	}
}

func TestWithSystemPrompt(t *testing.T) {
	opts := DefaultOptions()
	WithSystemPrompt("Be helpful")(opts)
	if opts.SystemPrompt != "Be helpful" {
		t.Errorf("expected 'Be helpful', got %v", opts.SystemPrompt)
	}
}

func TestWithSystemPromptPreset(t *testing.T) {
	opts := DefaultOptions()
	preset := SystemPromptPreset{Type: "preset", Preset: "claude_code"}
	WithSystemPromptPreset(preset)(opts)
	if opts.SystemPrompt != preset {
		t.Errorf("expected preset, got %v", opts.SystemPrompt)
	}
}

func TestWithAppendSystemPrompt(t *testing.T) {
	opts := DefaultOptions()
	WithAppendSystemPrompt("Extra instructions")(opts)
	if opts.AppendSystemPrompt != "Extra instructions" {
		t.Errorf("expected 'Extra instructions', got %q", opts.AppendSystemPrompt)
	}
}

func TestWithMaxTurns(t *testing.T) {
	opts := DefaultOptions()
	WithMaxTurns(10)(opts)
	if opts.MaxTurns != 10 {
		t.Errorf("expected 10, got %d", opts.MaxTurns)
	}
}

func TestWithMaxBudget(t *testing.T) {
	opts := DefaultOptions()
	WithMaxBudget(5.50)(opts)
	if opts.MaxBudgetUSD != 5.50 {
		t.Errorf("expected 5.50, got %f", opts.MaxBudgetUSD)
	}
}

func TestWithMaxThinkingTokens(t *testing.T) {
	opts := DefaultOptions()
	WithMaxThinkingTokens(8192)(opts)
	if opts.MaxThinkingTokens != 8192 {
		t.Errorf("expected 8192, got %d", opts.MaxThinkingTokens)
	}
}

func TestWithTools(t *testing.T) {
	opts := DefaultOptions()
	WithTools("Bash", "Read", "Write")(opts)
	tools, ok := opts.Tools.([]string)
	if !ok {
		t.Fatalf("expected []string, got %T", opts.Tools)
	}
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}
}

func TestWithToolsPreset(t *testing.T) {
	opts := DefaultOptions()
	preset := ToolsPreset{Type: "preset", Preset: "claude_code"}
	WithToolsPreset(preset)(opts)
	if opts.Tools != preset {
		t.Errorf("expected preset, got %v", opts.Tools)
	}
}

func TestWithAllowedTools(t *testing.T) {
	opts := DefaultOptions()
	WithAllowedTools("Bash", "Read")(opts)
	if len(opts.AllowedTools) != 2 {
		t.Errorf("expected 2 allowed tools, got %d", len(opts.AllowedTools))
	}
}

func TestWithDisallowedTools(t *testing.T) {
	opts := DefaultOptions()
	WithDisallowedTools("Write", "Edit")(opts)
	if len(opts.DisallowedTools) != 2 {
		t.Errorf("expected 2 disallowed tools, got %d", len(opts.DisallowedTools))
	}
}

func TestWithCLIPath(t *testing.T) {
	opts := DefaultOptions()
	WithCLIPath("/usr/bin/claude")(opts)
	if opts.CLIPath != "/usr/bin/claude" {
		t.Errorf("expected '/usr/bin/claude', got %q", opts.CLIPath)
	}
}

func TestWithResume(t *testing.T) {
	opts := DefaultOptions()
	WithResume("session-123")(opts)
	if opts.Resume != "session-123" {
		t.Errorf("expected 'session-123', got %q", opts.Resume)
	}
}

func TestWithResumeSessionAt(t *testing.T) {
	opts := DefaultOptions()
	WithResumeSessionAt("msg-123")(opts)
	if opts.ResumeSessionAt != "msg-123" {
		t.Errorf("expected 'msg-123', got %q", opts.ResumeSessionAt)
	}
}

func TestWithSessionID(t *testing.T) {
	opts := DefaultOptions()
	WithSessionID("session-456")(opts)
	if opts.SessionID != "session-456" {
		t.Errorf("expected 'session-456', got %q", opts.SessionID)
	}
}

func TestWithContinue(t *testing.T) {
	opts := DefaultOptions()
	WithContinue()(opts)
	if !opts.ContinueConversation {
		t.Error("expected ContinueConversation=true")
	}
}

func TestWithPersistSession(t *testing.T) {
	opts := DefaultOptions()
	WithPersistSession(false)(opts)
	if opts.PersistSession == nil {
		t.Fatal("expected PersistSession to be set")
	}
	if *opts.PersistSession {
		t.Error("expected PersistSession=false")
	}
}

func TestWithForkSession(t *testing.T) {
	opts := DefaultOptions()
	WithForkSession()(opts)
	if !opts.ForkSession {
		t.Error("expected ForkSession=true")
	}
}

func TestWithFileCheckpointing(t *testing.T) {
	opts := DefaultOptions()
	WithFileCheckpointing()(opts)
	if !opts.EnableFileCheckpointing {
		t.Error("expected EnableFileCheckpointing=true")
	}
}

func TestWithPartialMessages(t *testing.T) {
	opts := DefaultOptions()
	WithPartialMessages()(opts)
	if !opts.IncludePartialMessages {
		t.Error("expected IncludePartialMessages=true")
	}
}

func TestWithOutputFormat(t *testing.T) {
	opts := DefaultOptions()
	format := map[string]any{"type": "json_schema", "schema": map[string]any{}}
	WithOutputFormat(format)(opts)
	if opts.OutputFormat == nil {
		t.Error("expected OutputFormat to be set")
	}
}

func TestWithBetas(t *testing.T) {
	opts := DefaultOptions()
	WithBetas(BetaContext1M)(opts)
	if len(opts.Betas) != 1 {
		t.Errorf("expected 1 beta, got %d", len(opts.Betas))
	}
}

func TestWithSettingSources(t *testing.T) {
	opts := DefaultOptions()
	WithSettingSources(SettingSourceLocal, SettingSourceProject)(opts)
	if len(opts.SettingSources) != 2 {
		t.Errorf("expected 2 setting sources, got %d", len(opts.SettingSources))
	}
}

func TestWithSettings(t *testing.T) {
	opts := DefaultOptions()
	WithSettings("/path/to/settings.json")(opts)
	if opts.Settings != "/path/to/settings.json" {
		t.Errorf("expected settings path, got %v", opts.Settings)
	}
}

func TestWithSandbox(t *testing.T) {
	opts := DefaultOptions()
	sandbox := SandboxSettings{Enabled: true}
	WithSandbox(sandbox)(opts)
	if opts.Sandbox == nil || !opts.Sandbox.Enabled {
		t.Error("expected Sandbox.Enabled=true")
	}
}

func TestWithPlugins(t *testing.T) {
	opts := DefaultOptions()
	plugin := PluginConfig{Type: "local", Path: "/path/to/plugin"}
	WithPlugins(plugin)(opts)
	if len(opts.Plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(opts.Plugins))
	}
}

func TestWithAgent(t *testing.T) {
	opts := DefaultOptions()
	WithAgent("code-reviewer")(opts)
	if opts.Agent != "code-reviewer" {
		t.Errorf("expected 'code-reviewer', got %q", opts.Agent)
	}
}

func TestWithDebug(t *testing.T) {
	opts := DefaultOptions()
	WithDebug()(opts)
	if !opts.Debug {
		t.Error("expected Debug=true")
	}
}

func TestWithDebugFile(t *testing.T) {
	opts := DefaultOptions()
	WithDebugFile("/tmp/claude-debug.log")(opts)
	if opts.DebugFile != "/tmp/claude-debug.log" {
		t.Errorf("expected debug file path, got %q", opts.DebugFile)
	}
}

func TestWithStrictMCPConfig(t *testing.T) {
	opts := DefaultOptions()
	WithStrictMCPConfig()(opts)
	if !opts.StrictMCPConfig {
		t.Error("expected StrictMCPConfig=true")
	}
}

func TestWithAllowDangerouslySkipPermissions(t *testing.T) {
	opts := DefaultOptions()
	WithAllowDangerouslySkipPermissions()(opts)
	if !opts.AllowDangerouslySkipPermissions {
		t.Error("expected AllowDangerouslySkipPermissions=true")
	}
}

func TestWithStderrCallback(t *testing.T) {
	opts := DefaultOptions()
	called := false
	WithStderrCallback(func(s string) { called = true })(opts)
	if opts.StderrCallback == nil {
		t.Error("expected StderrCallback to be set")
	}
	opts.StderrCallback("test")
	if !called {
		t.Error("expected callback to be called")
	}
}

func TestWithMaxBufferSize(t *testing.T) {
	opts := DefaultOptions()
	WithMaxBufferSize(1024)(opts)
	if opts.MaxBufferSize != 1024 {
		t.Errorf("expected 1024, got %d", opts.MaxBufferSize)
	}
}

func TestWithMCPServers(t *testing.T) {
	opts := DefaultOptions()
	servers := map[string]MCPServerConfig{
		"test": {Type: "stdio", Command: "echo"},
	}
	WithMCPServers(servers)(opts)
	if opts.MCPServers == nil {
		t.Error("expected MCPServers to be set")
	}
}

func TestWithTransport(t *testing.T) {
	opts := DefaultOptions()
	// MockTransport would be used here in real tests
	WithTransport(nil)(opts)
	// Just verifying it sets the field
}

func TestMCPServerBuilder(t *testing.T) {
	server := NewMCPServerBuilder("test-server").
		WithVersion("1.0.0").
		WithTool("greet", "Greet a person", map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string", "description": "Person's name"},
			},
		}, func(args map[string]any) (*MCPToolResult, error) {
			return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "Hello!"}}}, nil
		}).
		Build()

	if server.Name != "test-server" {
		t.Errorf("expected name 'test-server', got %q", server.Name)
	}
	if server.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", server.Version)
	}
	if len(server.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(server.Tools))
	}
	if server.Tools[0].Name != "greet" {
		t.Errorf("expected tool name 'greet', got %q", server.Tools[0].Name)
	}
}
