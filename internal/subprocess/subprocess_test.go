// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

func TestFindCLI_WithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "my-claude")
	if err := os.WriteFile(mockCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := findCLI(mockCLI, "")
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != mockCLI {
		t.Errorf("got %q, want %q", path, mockCLI)
	}
}

func TestFindCLI_WithBundledPath(t *testing.T) {
	tmpDir := t.TempDir()
	bundledCLI := filepath.Join(tmpDir, "bundled-claude")
	if err := os.WriteFile(bundledCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := findCLI("", bundledCLI)
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != bundledCLI {
		t.Errorf("got %q, want %q", path, bundledCLI)
	}
}

func TestFindCLI_WithEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(mockCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	path, err := findCLI("", "")
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != mockCLI {
		t.Errorf("got %q, want %q", path, mockCLI)
	}
}

func TestFindCLI_NotFound(t *testing.T) {
	// Use a nonexistent bundled path and empty explicit path
	// The CLI might be found in PATH or common locations on the test system,
	// so we test the error case by providing an invalid explicit path
	_, err := findCLI("/definitely/nonexistent/path/claude", "")
	if err == nil {
		t.Error("expected error when CLI not found")
	}

	var notFoundErr *types.CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T", err)
	}
}

func TestFindCLI_ExplicitPathNotExists(t *testing.T) {
	_, err := findCLI("/nonexistent/path/to/claude", "")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}

	var notFoundErr *types.CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T", err)
	}
}

func TestBuildCommand_Basic(t *testing.T) {
	opts := types.DefaultOptions()
	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	if cmd[0] != "/usr/bin/claude" {
		t.Errorf("got %q, want %q", cmd[0], "/usr/bin/claude")
	}

	hasOutputFormat := false
	for i, arg := range cmd {
		if arg == "--output-format" && i+1 < len(cmd) && cmd[i+1] == "stream-json" {
			hasOutputFormat = true
			break
		}
	}
	if !hasOutputFormat {
		t.Error("missing --output-format stream-json")
	}
}

func TestBuildCommand_Streaming(t *testing.T) {
	opts := types.DefaultOptions()
	cmd := buildCommand("/usr/bin/claude", "", opts, true)

	hasInputFormat := false
	for i, arg := range cmd {
		if arg == "--input-format" && i+1 < len(cmd) && cmd[i+1] == "stream-json" {
			hasInputFormat = true
			break
		}
	}
	if !hasInputFormat {
		t.Error("missing --input-format stream-json for streaming mode")
	}
}

func TestBuildProcessCommand_NativeEntrypoint(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Executable = "node" // Ignored for native entrypoints (TS parity)
	opts.ExecutableArgs = []string{"--trace-warnings"}

	cmd := buildProcessCommand("/usr/bin/claude", []string{"--output-format", "stream-json"}, opts)
	if len(cmd) != 3 {
		t.Fatalf("unexpected command length: %+v", cmd)
	}
	if cmd[0] != "/usr/bin/claude" {
		t.Fatalf("expected native entrypoint command, got %q", cmd[0])
	}
	if cmd[1] != "--output-format" {
		t.Fatalf("expected native command to ignore executable args, got %+v", cmd)
	}
	if cmd[2] != "stream-json" {
		t.Fatalf("unexpected cli arg position: %+v", cmd)
	}
}

func TestBuildProcessCommand_ScriptEntrypoint(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Executable = "bun"
	opts.ExecutableArgs = []string{"--smol"}

	cmd := buildProcessCommand("/tmp/cli.js", []string{"--output-format", "stream-json"}, opts)
	expected := []string{"bun", "--smol", "/tmp/cli.js", "--output-format", "stream-json"}
	if len(cmd) != len(expected) {
		t.Fatalf("unexpected command length: got %d want %d (%+v)", len(cmd), len(expected), cmd)
	}
	for i := range expected {
		if cmd[i] != expected[i] {
			t.Fatalf("unexpected command at %d: got %q want %q (%+v)", i, cmd[i], expected[i], cmd)
		}
	}
}

func TestBuildProcessCommand_ScriptEntrypoint_DefaultRuntime(t *testing.T) {
	opts := types.DefaultOptions()
	cmd := buildProcessCommand("/tmp/cli.mjs", []string{"--output-format", "stream-json"}, opts)
	if len(cmd) < 2 {
		t.Fatalf("unexpected command length: %+v", cmd)
	}
	if cmd[0] != "node" {
		t.Fatalf("expected default runtime to be node, got %q", cmd[0])
	}
	if cmd[1] != "/tmp/cli.mjs" {
		t.Fatalf("expected script path in argv, got %+v", cmd)
	}
}

func TestBuildCommand_WithOptions(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Model = "claude-opus-4"
	opts.MaxTurns = 5
	opts.PermissionMode = types.PermissionBypass
	opts.SystemPrompt = "You are helpful"

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	checks := map[string]string{
		"--model":           "claude-opus-4",
		"--max-turns":       "5",
		"--permission-mode": "bypassPermissions",
		"--system-prompt":   "You are helpful",
	}

	for flag, value := range checks {
		found := false
		for i, arg := range cmd {
			if arg == flag && i+1 < len(cmd) && cmd[i+1] == value {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing %s %s in command", flag, value)
		}
	}
}

func TestBuildCommand_ToolsEmptyArrayDisablesTools(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Tools = []string{}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	found := false
	for i, arg := range cmd {
		if arg == "--tools" && i+1 < len(cmd) && cmd[i+1] == "" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected --tools with empty value, got %+v", cmd)
	}
}

func TestBuildCommand_ToolsPresetUsesDefault(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Tools = types.ToolsPreset{Type: "preset", Preset: "claude_code"}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	found := false
	for i, arg := range cmd {
		if arg == "--tools" && i+1 < len(cmd) && cmd[i+1] == "default" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected --tools default, got %+v", cmd)
	}
}

func TestBuildCommand_ToolsPresetPassThrough(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Tools = types.ToolsPreset{Type: "preset", Preset: "experimental_toolset"}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	found := false
	for i, arg := range cmd {
		if arg == "--tools" && i+1 < len(cmd) && cmd[i+1] == "experimental_toolset" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected --tools experimental_toolset, got %+v", cmd)
	}
}

func TestBuildCommand_SystemPromptPresetWithoutAppend(t *testing.T) {
	opts := types.DefaultOptions()
	opts.SystemPrompt = types.SystemPromptPreset{Type: "preset", Preset: "claude_code"}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	for _, arg := range cmd {
		if arg == "--system-prompt" {
			t.Fatalf("did not expect --system-prompt for preset without append, got %+v", cmd)
		}
		if arg == "--append-system-prompt" {
			t.Fatalf("did not expect --append-system-prompt for preset without append, got %+v", cmd)
		}
	}
}

func TestBuildCommand_SystemPromptPresetWithAppend(t *testing.T) {
	opts := types.DefaultOptions()
	appendText := "Always include test coverage notes."
	opts.SystemPrompt = types.SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: &appendText,
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	hasAppend := false
	hasSystemPrompt := false
	for i, arg := range cmd {
		if arg == "--append-system-prompt" && i+1 < len(cmd) && cmd[i+1] == appendText {
			hasAppend = true
		}
		if arg == "--system-prompt" {
			hasSystemPrompt = true
		}
	}
	if !hasAppend {
		t.Fatalf("expected --append-system-prompt for preset append, got %+v", cmd)
	}
	if hasSystemPrompt {
		t.Fatalf("did not expect --system-prompt for preset append flow, got %+v", cmd)
	}
}

func TestBuildCommand_SystemPromptPresetPassThrough(t *testing.T) {
	opts := types.DefaultOptions()
	opts.SystemPrompt = types.SystemPromptPreset{Type: "preset", Preset: "custom_preset"}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	var rawPreset string
	for i, arg := range cmd {
		if arg == "--system-prompt" && i+1 < len(cmd) {
			rawPreset = cmd[i+1]
			break
		}
	}
	if rawPreset == "" {
		t.Fatalf("expected --system-prompt for custom preset, got %+v", cmd)
	}
	var preset map[string]any
	if err := json.Unmarshal([]byte(rawPreset), &preset); err != nil {
		t.Fatalf("expected JSON preset payload, got %q (%v)", rawPreset, err)
	}
	if preset["type"] != "preset" || preset["preset"] != "custom_preset" {
		t.Fatalf("unexpected preset payload: %+v", preset)
	}
}

func TestIsScriptEntryPoint_ExtensionlessShebang(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "claude-wrapper")
	if err := os.WriteFile(scriptPath, []byte("#!/usr/bin/env node\nconsole.log('hi');\n"), 0755); err != nil {
		t.Fatalf("failed to write test script: %v", err)
	}
	if !isScriptEntryPoint(scriptPath) {
		t.Fatalf("expected extensionless shebang script to be treated as script: %s", scriptPath)
	}
}

func TestIsScriptEntryPoint_ExtensionlessShellScriptIsNative(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "claude-shell-wrapper")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho hi\n"), 0755); err != nil {
		t.Fatalf("failed to write shell test script: %v", err)
	}
	if isScriptEntryPoint(scriptPath) {
		t.Fatalf("expected extensionless shell script to be treated as native: %s", scriptPath)
	}
}

func TestSubprocessTransport_PathToClaudeCodeExecutablePrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	cliPath := filepath.Join(tmpDir, "cli-primary")
	overridePath := filepath.Join(tmpDir, "cli-override")
	script := `#!/bin/sh
if [ "$1" = "-v" ]; then
  echo "2.0.0"
  exit 0
fi
while IFS= read -r _line; do
  :
done
`
	if err := os.WriteFile(cliPath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write primary cli script: %v", err)
	}
	if err := os.WriteFile(overridePath, []byte(script), 0755); err != nil {
		t.Fatalf("failed to write override cli script: %v", err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = cliPath
	opts.PathToClaudeCodeExecutable = overridePath

	transport := NewStreamingTransport(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	if transport.cliPath != overridePath {
		t.Fatalf("expected override path %q, got %q", overridePath, transport.cliPath)
	}
}

func TestBuildCommand_NewOptionFlags(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Agent = "code-reviewer"
	opts.SessionID = "11111111-1111-1111-1111-111111111111"
	opts.ResumeSessionAt = "22222222-2222-2222-2222-222222222222"
	opts.Debug = true
	opts.StrictMCPConfig = true
	opts.AllowDangerouslySkipPermissions = true
	persist := false
	opts.PersistSession = &persist

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	checks := map[string]string{
		"--agent":             "code-reviewer",
		"--session-id":        "11111111-1111-1111-1111-111111111111",
		"--resume-session-at": "22222222-2222-2222-2222-222222222222",
	}

	for flag, value := range checks {
		found := false
		for i, arg := range cmd {
			if arg == flag && i+1 < len(cmd) && cmd[i+1] == value {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing %s %s in command", flag, value)
		}
	}

	boolFlags := []string{
		"--debug",
		"--strict-mcp-config",
		"--allow-dangerously-skip-permissions",
		"--no-session-persistence",
	}
	for _, flag := range boolFlags {
		found := false
		for _, arg := range cmd {
			if arg == flag {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing %s in command", flag)
		}
	}
}

func TestBuildCommand_DebugFileTakesPrecedenceOverDebug(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Debug = true
	opts.DebugFile = "/tmp/claude-debug.log"

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	hasDebug := false
	hasDebugFile := false
	for i, arg := range cmd {
		if arg == "--debug" {
			hasDebug = true
		}
		if arg == "--debug-file" && i+1 < len(cmd) && cmd[i+1] == "/tmp/claude-debug.log" {
			hasDebugFile = true
		}
	}

	if hasDebug {
		t.Error("--debug should not be present when --debug-file is set")
	}
	if !hasDebugFile {
		t.Error("missing --debug-file flag")
	}
}

func TestBuildCommand_PersistSessionTriState(t *testing.T) {
	tests := []struct {
		name            string
		persist         *bool
		expectNoPersist bool
	}{
		{
			name:            "nil means default behavior",
			persist:         nil,
			expectNoPersist: false,
		},
		{
			name: "true does not pass no-session flag",
			persist: func() *bool {
				v := true
				return &v
			}(),
			expectNoPersist: false,
		},
		{
			name: "false passes no-session flag",
			persist: func() *bool {
				v := false
				return &v
			}(),
			expectNoPersist: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := types.DefaultOptions()
			opts.PersistSession = tt.persist

			cmd := buildCommand("/usr/bin/claude", "test", opts, false)

			hasNoPersist := false
			for _, arg := range cmd {
				if arg == "--no-session-persistence" {
					hasNoPersist = true
					break
				}
			}

			if hasNoPersist != tt.expectNoPersist {
				t.Fatalf("expected --no-session-persistence presence=%v, got %v", tt.expectNoPersist, hasNoPersist)
			}
		})
	}
}

func TestBuildCommand_MCPServers(t *testing.T) {
	opts := types.DefaultOptions()
	opts.MCPServers = map[string]types.MCPServerConfig{
		"test-server": {
			Command: "node",
			Args:    []string{"server.js"},
		},
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	hasMCPConfig := false
	for i, arg := range cmd {
		if arg == "--mcp-config" && i+1 < len(cmd) {
			hasMCPConfig = true
			break
		}
	}
	if !hasMCPConfig {
		t.Error("missing --mcp-config for MCP servers")
	}
}

func TestBuildCommand_SandboxConfig(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Sandbox = &types.SandboxSettings{
		Enabled: true,
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// After merging change, sandbox should be in --settings, not --sandbox
	hasSettings := false
	for i, arg := range cmd {
		if arg == "--settings" && i+1 < len(cmd) {
			hasSettings = true
			// Verify it contains sandbox
			settingsValue := cmd[i+1]
			var settingsObj map[string]any
			if err := json.Unmarshal([]byte(settingsValue), &settingsObj); err == nil {
				if _, ok := settingsObj["sandbox"]; !ok {
					t.Error("settings does not contain sandbox key")
				}
			} else {
				t.Errorf("failed to parse settings JSON: %v", err)
			}
			break
		}
	}
	if !hasSettings {
		t.Error("missing --settings flag (sandbox should be merged into settings)")
	}

	// Verify --sandbox flag is NOT present
	for _, arg := range cmd {
		if arg == "--sandbox" {
			t.Error("--sandbox flag should not be present (should be merged into settings)")
		}
	}
}

func TestCommandLength_Windows(t *testing.T) {
	// Test that very long commands are handled on Windows
	opts := types.DefaultOptions()
	opts.SystemPrompt = string(make([]byte, 10000)) // Very long prompt

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	// On Windows, total command length should be checked
	totalLen := 0
	for _, arg := range cmd {
		totalLen += len(arg) + 1 // +1 for space
	}

	// Windows limit is 8191 characters
	if runtime.GOOS == "windows" && totalLen > 8191 {
		t.Log("Warning: command exceeds Windows limit, should be handled")
	}
}

func TestNewSubprocessTransport(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewSubprocessTransport("Hello", opts)

	if transport == nil {
		t.Fatal("NewSubprocessTransport returned nil")
	}

	if transport.IsReady() {
		t.Error("should not be ready before Connect")
	}
}

func TestNewStreamingTransport(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewStreamingTransport(opts)

	if transport == nil {
		t.Fatal("NewStreamingTransport returned nil")
	}

	if !transport.streaming {
		t.Error("should be in streaming mode")
	}
}

func TestSubprocessTransportImplementsInterface(t *testing.T) {
	var _ types.Transport = (*SubprocessTransport)(nil)
}

func TestSubprocessTransport_Connect_NotFound(t *testing.T) {
	opts := types.DefaultOptions()
	opts.CLIPath = "/nonexistent/path/to/claude"

	transport := NewSubprocessTransport("Hello", opts)
	err := transport.Connect(context.Background())

	if err == nil {
		t.Error("expected error for nonexistent CLI")
	}

	var notFoundErr *types.CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T: %v", err, err)
	}
}

func TestSubprocessTransport_Connect_AlreadyConnected(t *testing.T) {
	// Create mock CLI
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	mockScript := `#!/bin/bash
echo '{"type":"system","subtype":"init"}'
sleep 0.1
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewSubprocessTransport("Hello", opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First connect should succeed
	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("first Connect failed: %v", err)
	}
	defer transport.Close()

	// Second connect should return nil (already connected)
	if err := transport.Connect(ctx); err != nil {
		t.Errorf("second Connect should return nil: %v", err)
	}
}

func TestParseJSONLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   `{"type":"assistant","message":{"content":"hello"}}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{not valid json`,
			wantErr: true,
		},
		{
			name:    "empty line",
			input:   "",
			wantErr: true,
		},
		{
			name:    "partial json",
			input:   `{"type":"assistant"`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseJSONLine(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected result, got nil")
				}
			}
		})
	}
}

func TestSpeculativeJSONParsing(t *testing.T) {
	// Test that multi-line JSON is accumulated correctly
	lines := []string{
		`{"type":"assistant",`,
		`"message":{"content":"hello"}}`,
	}

	parser := newJSONAccumulator()
	var result map[string]any
	var err error

	for _, line := range lines {
		result, err = parser.addLine(line)
		if result != nil {
			break
		}
	}

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected result after accumulating lines")
	}
	if result["type"] != "assistant" {
		t.Errorf("got type %v, want assistant", result["type"])
	}
}

func TestSubprocessTransport_Write_NotReady(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewSubprocessTransport("", opts)

	err := transport.Write(`{"type":"user","message":{"content":"hello"}}`)
	if err == nil {
		t.Error("expected error when writing to non-ready transport")
	}

	var connErr *types.ConnectionError
	if !errors.As(err, &connErr) {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestSubprocessTransport_Close_NotConnected(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewSubprocessTransport("", opts)

	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubprocessTransport_Close_Idempotent(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewSubprocessTransport("", opts)

	// Multiple closes should not panic
	_ = transport.Close()
	_ = transport.Close()
	_ = transport.Close()
}

func TestSubprocessTransport_Close_CleansUpTempFiles(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "temp.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	transport := NewSubprocessTransport("", opts)
	transport.AddTempFile(tmpFile)

	if err := transport.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Temp file should be deleted
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("temp file should have been deleted")
	}
}

func TestSubprocessTransport_Close_GracefulShutdown(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	// Process that responds to SIGTERM
	mockScript := `#!/bin/bash
trap 'exit 0' TERM
while true; do sleep 0.1; done
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewStreamingTransport(opts)
	ctx := context.Background()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Close should complete within timeout
	done := make(chan struct{})
	go func() {
		transport.Close()
		close(done)
	}()

	select {
	case <-done:
		// Good - closed within expected time
	case <-time.After(10 * time.Second):
		t.Error("Close took too long")
	}
}

func TestSubprocessTransport_StderrCallback(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	mockScript := `#!/bin/bash
echo "stderr line 1" >&2
echo "stderr line 2" >&2
echo '{"type":"result"}'
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewSubprocessTransport("Hello", opts)

	stderrLines := make([]string, 0)
	var mu sync.Mutex
	transport.SetStderrCallback(func(line string) {
		mu.Lock()
		stderrLines = append(stderrLines, line)
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	// Drain messages
	for range transport.Messages() {
	}
	transport.Close()

	mu.Lock()
	defer mu.Unlock()

	if len(stderrLines) != 2 {
		t.Errorf("expected 2 stderr lines, got %d: %v", len(stderrLines), stderrLines)
	}
}

func TestSubprocessTransport_ConcurrentWrites_Race(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	mockScript := `#!/bin/bash
while read -r line; do
    echo '{"type":"ack"}'
done
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewStreamingTransport(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Hammer with concurrent writes to trigger race conditions
	const numWriters = 50
	const writesPerWriter = 100

	var wg sync.WaitGroup
	errors := make(chan error, numWriters*writesPerWriter)

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < writesPerWriter; j++ {
				msg := `{"writer":` + string(rune('0'+writerID%10)) + `,"msg":` + string(rune('0'+j%10)) + `}`
				if err := transport.Write(msg); err != nil {
					errors <- err
					return
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	var errCount int
	for err := range errors {
		t.Errorf("Write error: %v", err)
		errCount++
	}

	if errCount > 0 {
		t.Errorf("Total errors: %d", errCount)
	}
}

func TestSubprocessTransport_Integration_MockCLI(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	mockScript := `#!/bin/bash
echo '{"type":"system","subtype":"init","data":{"version":"2.0.0"}}'
echo '{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}],"model":"claude-test"}}'
echo '{"type":"result","subtype":"success","duration_ms":100,"is_error":false}'
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewSubprocessTransport("Hello", opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Collect messages
	var messages []map[string]any
	for msg := range transport.Messages() {
		messages = append(messages, msg)
	}

	// Verify messages
	if len(messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(messages))
	}

	expectedTypes := []string{"system", "assistant", "result"}
	for i, expected := range expectedTypes {
		if messages[i]["type"] != expected {
			t.Errorf("message %d: got type %v, want %s", i, messages[i]["type"], expected)
		}
	}
}

func TestSubprocessTransport_Integration_StreamingMode(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	mockScript := `#!/bin/bash
echo '{"type":"system","subtype":"init"}'
while read -r line; do
    if [[ -n "$line" ]]; then
        echo '{"type":"assistant","message":{"content":[{"type":"text","text":"Echo"}]}}'
        echo '{"type":"result","subtype":"success"}'
        exit 0
    fi
done
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewStreamingTransport(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Send a message
	if err := transport.Write(`{"type":"user","message":{"content":"test"}}`); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Collect messages
	var messages []map[string]any
	for msg := range transport.Messages() {
		messages = append(messages, msg)
	}

	if len(messages) < 2 {
		t.Errorf("expected at least 2 messages, got %d", len(messages))
	}
}
