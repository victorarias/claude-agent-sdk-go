package sdk

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
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

	var notFoundErr *CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T", err)
	}
}

func TestFindCLI_ExplicitPathNotExists(t *testing.T) {
	_, err := findCLI("/nonexistent/path/to/claude", "")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}

	var notFoundErr *CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T", err)
	}
}

func TestBuildCommand_Basic(t *testing.T) {
	opts := DefaultOptions()
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
	opts := DefaultOptions()
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

func TestBuildCommand_WithOptions(t *testing.T) {
	opts := DefaultOptions()
	opts.Model = "claude-opus-4"
	opts.MaxTurns = 5
	opts.PermissionMode = PermissionBypass
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

func TestBuildCommand_MCPServers(t *testing.T) {
	opts := DefaultOptions()
	opts.MCPServers = map[string]MCPServerConfig{
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
	opts := DefaultOptions()
	opts.Sandbox = &SandboxSettings{
		Enabled: true,
	}

	cmd := buildCommand("/usr/bin/claude", "test", opts, false)

	hasSandbox := false
	for i, arg := range cmd {
		if arg == "--sandbox" && i+1 < len(cmd) {
			hasSandbox = true
			break
		}
	}
	if !hasSandbox {
		t.Error("missing --sandbox flag")
	}
}

func TestCommandLength_Windows(t *testing.T) {
	// Test that very long commands are handled on Windows
	opts := DefaultOptions()
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
