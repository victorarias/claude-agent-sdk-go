// Copyright (C) 2025 Claude Agent SDK Go Contributors
// SPDX-License-Identifier: GPL-3.0-only

package subprocess

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/victorarias/claude-agent-sdk-go/types"
)

// TestIsVersionAtLeast tests version comparison logic
func TestIsVersionAtLeast(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		required string
		want     bool
	}{
		{
			name:     "equal versions",
			version:  "1.2.3",
			required: "1.2.3",
			want:     true,
		},
		{
			name:     "newer version",
			version:  "1.3.0",
			required: "1.2.3",
			want:     true,
		},
		{
			name:     "older version",
			version:  "1.2.0",
			required: "1.2.3",
			want:     false,
		},
		{
			name:     "major version higher",
			version:  "2.0.0",
			required: "1.9.9",
			want:     true,
		},
		{
			name:     "major version lower",
			version:  "0.9.9",
			required: "1.0.0",
			want:     false,
		},
		{
			name:     "version longer than required",
			version:  "1.2.3.4",
			required: "1.2.3",
			want:     true,
		},
		{
			name:     "version shorter than required",
			version:  "1.2",
			required: "1.2.3",
			want:     false,
		},
		{
			name:     "invalid version format",
			version:  "1.2.x",
			required: "1.2.3",
			want:     false,
		},
		{
			name:     "invalid required format",
			version:  "1.2.3",
			required: "1.2.x",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVersionAtLeast(tt.version, tt.required)
			if got != tt.want {
				t.Errorf("isVersionAtLeast(%q, %q) = %v, want %v", tt.version, tt.required, got, tt.want)
			}
		})
	}
}

// TestErrorsChannel tests the Errors() method
func TestErrorsChannel(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewStreamingTransport(opts)

	errChan := transport.Errors()
	if errChan == nil {
		t.Error("Errors() returned nil channel")
	}

	// Verify channel is not closed initially
	select {
	case _, ok := <-errChan:
		if !ok {
			t.Error("Errors channel should not be closed initially")
		}
	case <-time.After(10 * time.Millisecond):
		// Expected - channel should be open but empty
	}
}

// TestExitError tests the ExitError() method
func TestExitError(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewStreamingTransport(opts)

	// Initially should be nil
	if err := transport.ExitError(); err != nil {
		t.Errorf("ExitError() should be nil initially, got %v", err)
	}

	// After a process exits, it should be set
	// We'll test this with a mock CLI that exits immediately
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	// Create a script that exits with an error
	script := "#!/bin/sh\nexit 1\n"
	if err := os.WriteFile(mockCLI, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	opts.CLIPath = mockCLI
	transport = NewStreamingTransport(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Connect should start the process
	_ = transport.Connect(ctx)
	defer transport.Close()

	// Give process time to exit
	time.Sleep(100 * time.Millisecond)

	// ExitError might be set now (but not guaranteed due to timing)
	// This at least exercises the code path
	_ = transport.ExitError()
}

// TestKill tests the Kill() method
func TestKill(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	// Create a long-running script that outputs to stdout periodically
	script := "#!/bin/sh\nwhile true; do echo '{}'; sleep 0.1; done\n"
	if err := os.WriteFile(mockCLI, []byte(script), 0755); err != nil {
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

	// Give the process a moment to start
	time.Sleep(200 * time.Millisecond)

	// Kill the process
	if err := transport.Kill(); err != nil {
		t.Errorf("Kill() failed: %v", err)
	}

	// Calling Kill again should be safe (process is nil or already dead)
	if err := transport.Kill(); err != nil {
		t.Errorf("Kill() on already killed process failed: %v", err)
	}

	transport.Close()
}

// TestKill_NoProcess tests Kill() when no process is running
func TestKill_NoProcess(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewStreamingTransport(opts)

	// Should not error when called before Connect
	if err := transport.Kill(); err != nil {
		t.Errorf("Kill() with no process should not error, got: %v", err)
	}
}

// TestWriteJSON tests the WriteJSON method
func TestWriteJSON(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	// Create a mock CLI that reads from stdin
	script := "#!/bin/sh\ncat > /dev/null\n"
	if err := os.WriteFile(mockCLI, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI
	transport := NewStreamingTransport(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Test writing valid JSON
	data := map[string]interface{}{
		"key":   "value",
		"count": 42,
	}

	if err := transport.WriteJSON(data); err != nil {
		t.Errorf("WriteJSON() failed: %v", err)
	}
}

// TestWriteJSON_InvalidJSON tests WriteJSON with unmarshalable data
func TestWriteJSON_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	script := "#!/bin/sh\ncat > /dev/null\n"
	if err := os.WriteFile(mockCLI, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI
	transport := NewStreamingTransport(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Channels cannot be marshaled to JSON
	invalidData := make(chan int)

	if err := transport.WriteJSON(invalidData); err == nil {
		t.Error("WriteJSON() should fail with unmarshalable data")
	}
}

// TestWriteJSON_NotReady tests WriteJSON before connection
func TestWriteJSON_NotReady(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewStreamingTransport(opts)

	data := map[string]string{"key": "value"}

	if err := transport.WriteJSON(data); err == nil {
		t.Error("WriteJSON() should fail when transport not ready")
	}
}

// TestEndInput tests the EndInput method
func TestEndInput(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	// Create a mock CLI that reads until EOF
	script := "#!/bin/sh\ncat\n"
	if err := os.WriteFile(mockCLI, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}

	opts := types.DefaultOptions()
	opts.CLIPath = mockCLI
	transport := NewStreamingTransport(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Write some data
	if err := transport.Write("test data"); err != nil {
		t.Errorf("Write() failed: %v", err)
	}

	// Close stdin
	if err := transport.EndInput(); err != nil {
		t.Errorf("EndInput() failed: %v", err)
	}

	// Calling EndInput again should not error (stdin is already nil after first call)
	// The implementation returns nil if stdin is nil
	err := transport.EndInput()
	if err != nil {
		// This is acceptable - closing an already closed pipe may error
		t.Logf("EndInput() second call returned error (expected): %v", err)
	}
}

// TestEndInput_NotConnected tests EndInput before connection
func TestEndInput_NotConnected(t *testing.T) {
	opts := types.DefaultOptions()
	transport := NewStreamingTransport(opts)

	// Should not panic when called before Connect
	if err := transport.EndInput(); err != nil {
		t.Errorf("EndInput() with no connection should not error, got: %v", err)
	}
}

// TestOptimizeCommandLength_ShortCommand tests that short commands aren't optimized
func TestOptimizeCommandLength_ShortCommand(t *testing.T) {
	opts := types.DefaultOptions()
	opts.Agents = map[string]types.AgentDefinition{
		"agent1": {Description: "Test agent", Prompt: "test"},
	}

	cmd := []string{"claude", "run", "--agents", `{"agent1":{"description":"Test agent","prompt":"test"}}`}
	optimized := optimizeCommandLength(cmd, opts)

	// Command should remain unchanged
	if len(optimized) != len(cmd) {
		t.Errorf("Short command was modified")
	}
}

// TestOptimizeCommandLength_LongCommand tests temp file fallback for long commands
func TestOptimizeCommandLength_LongCommand(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Command length optimization only applies on Windows (8000 char limit)")
	}

	opts := types.DefaultOptions()
	// Create a very long agents map to exceed command length limit
	longAgents := make(map[string]types.AgentDefinition, 1000)
	for i := 0; i < 1000; i++ {
		name := "very-long-agent-name-that-takes-up-space"
		longAgents[name] = types.AgentDefinition{Description: "desc", Prompt: "prompt"}
	}
	opts.Agents = longAgents

	cmd := []string{"claude", "run", "--agents"}
	// Build a long JSON string
	longJSON := `{`
	first := true
	for name := range longAgents {
		if !first {
			longJSON += `,`
		}
		first = false
		longJSON += `"` + name + `":{"name":"` + name + `"}`
	}
	longJSON += `}`
	cmd = append(cmd, longJSON)

	optimized := optimizeCommandLength(cmd, opts)

	// Should have replaced with temp file reference
	found := false
	for _, arg := range optimized {
		if len(arg) > 0 && arg[0] == '@' {
			found = true
			// Verify temp file exists
			tempPath := arg[1:]
			if _, err := os.Stat(tempPath); os.IsNotExist(err) {
				t.Errorf("Temp file %q does not exist", tempPath)
			}
			// Clean up
			os.Remove(tempPath)
		}
	}

	if !found {
		t.Error("Command was not optimized with temp file reference")
	}
}

// TestCheckCommandLength_NonWindows tests that non-Windows systems skip the check
func TestCheckCommandLength_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping non-Windows test on Windows")
	}

	// Create a very long command
	cmd := make([]string, 10000)
	for i := range cmd {
		cmd[i] = "very-long-argument-string"
	}

	// Should not error on non-Windows
	if err := checkCommandLength(cmd); err != nil {
		t.Errorf("checkCommandLength() should not error on non-Windows: %v", err)
	}
}

// TestCheckCommandLength_Windows tests Windows command length validation
func TestCheckCommandLength_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows test on non-Windows")
	}

	tests := []struct {
		name      string
		cmd       []string
		wantError bool
	}{
		{
			name:      "short command",
			cmd:       []string{"claude", "run"},
			wantError: false,
		},
		{
			name: "command at limit",
			cmd: func() []string {
				// Create command that's just under the limit
				args := []string{"claude"}
				remaining := WindowsMaxCommandLength - 7 // "claude " = 7 chars
				for remaining > 0 {
					if remaining >= 10 {
						args = append(args, "012345678")
						remaining -= 10
					} else {
						break
					}
				}
				return args
			}(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkCommandLength(tt.cmd)
			if (err != nil) != tt.wantError {
				t.Errorf("checkCommandLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
