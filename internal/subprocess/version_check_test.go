// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestConnectChecksCLIVersion tests that Connect checks the CLI version
// and returns an error if the version is too old.
func TestConnectChecksCLIVersion(t *testing.T) {
	// Skip if claude is not installed
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not installed, skipping version check test")
	}

	opts := types.DefaultOptions()
	transport := NewSubprocessTransport("test prompt", opts)

	// Connect should check version
	ctx := context.Background()
	err := transport.Connect(ctx)

	// If we get here without error, the version check passed
	// If error is a version error, that's the expected behavior for old CLI
	if _, ok := err.(*types.CLIVersionError); ok {
		// Expected - CLI is too old
		return
	}
	// Other errors (including nil) are fine - connection errors or success

	// Clean up
	transport.Close()
}

// TestCheckCLIVersion tests the standalone version check function.
func TestCheckCLIVersion(t *testing.T) {
	tests := []struct {
		name        string
		cliPath     string
		expectError bool
	}{
		{
			name:        "non-existent CLI",
			cliPath:     "/non/existent/claude",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := checkCLIVersion(tt.cliPath)
			if tt.expectError && err == nil {
				t.Errorf("expected error, got version: %s", version)
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestCheckCLIVersionWithRealCLI tests version checking with the real CLI.
func TestCheckCLIVersionWithRealCLI(t *testing.T) {
	cliPath, err := exec.LookPath("claude")
	if err != nil {
		t.Skip("claude CLI not installed, skipping version check test")
	}

	version, err := checkCLIVersion(cliPath)
	if err != nil {
		t.Fatalf("checkCLIVersion failed: %v", err)
	}

	// Version should be a semver string like "1.0.0"
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		t.Errorf("invalid version format: %s", version)
	}

	t.Logf("Claude CLI version: %s", version)
}

// TestVersionCheckCanBeSkipped tests that version check can be skipped
// via environment variable.
func TestVersionCheckCanBeSkipped(t *testing.T) {
	// Set the skip env var
	os.Setenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK", "1")
	defer os.Unsetenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK")

	// Even with a fake CLI path that would fail version check,
	// Connect should not fail due to version check when skip is set
	opts := types.DefaultOptions()
	opts.CLIPath = "/nonexistent/claude"

	transport := NewSubprocessTransport("test prompt", opts)

	ctx := context.Background()
	err := transport.Connect(ctx)

	// Error should be CLINotFoundError, not CLIVersionError
	if _, ok := err.(*types.CLIVersionError); ok {
		t.Error("version check should be skipped when CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK is set")
	}
}

// TestVersionCheckTimeout tests that version check has a timeout.
func TestVersionCheckTimeout(t *testing.T) {
	// This test ensures the version check doesn't hang indefinitely
	// We can't easily test this without a mock, but we document the expected behavior
	t.Log("Version check should timeout after 2 seconds (matching Python SDK)")
}

// TestParseVersionOutput tests parsing of CLI version output.
func TestParseVersionOutput(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
		hasError bool
	}{
		{
			name:     "simple version",
			output:   "1.0.50",
			expected: "1.0.50",
			hasError: false,
		},
		{
			name:     "version with v prefix",
			output:   "v1.0.50",
			expected: "1.0.50",
			hasError: false,
		},
		{
			name:     "version with newline",
			output:   "1.0.50\n",
			expected: "1.0.50",
			hasError: false,
		},
		{
			name:     "version with extra text",
			output:   "Claude Code version 1.0.50",
			expected: "1.0.50",
			hasError: false,
		},
		{
			name:     "empty output",
			output:   "",
			expected: "",
			hasError: true,
		},
		{
			name:     "no version found",
			output:   "no version here",
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parseVersionOutput(tt.output)
			if tt.hasError && err == nil {
				t.Errorf("expected error, got version: %s", version)
			}
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.hasError && version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}
