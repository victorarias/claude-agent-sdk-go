package sdk

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"
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

func TestNewSubprocessTransport(t *testing.T) {
	opts := DefaultOptions()
	transport := NewSubprocessTransport("Hello", opts)

	if transport == nil {
		t.Fatal("NewSubprocessTransport returned nil")
	}

	if transport.IsReady() {
		t.Error("should not be ready before Connect")
	}
}

func TestNewStreamingTransport(t *testing.T) {
	opts := DefaultOptions()
	transport := NewStreamingTransport(opts)

	if transport == nil {
		t.Fatal("NewStreamingTransport returned nil")
	}

	if !transport.streaming {
		t.Error("should be in streaming mode")
	}
}

func TestSubprocessTransportImplementsInterface(t *testing.T) {
	var _ Transport = (*SubprocessTransport)(nil)
}

func TestSubprocessTransport_Connect_NotFound(t *testing.T) {
	opts := DefaultOptions()
	opts.CLIPath = "/nonexistent/path/to/claude"

	transport := NewSubprocessTransport("Hello", opts)
	err := transport.Connect(context.Background())

	if err == nil {
		t.Error("expected error for nonexistent CLI")
	}

	var notFoundErr *CLINotFoundError
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

	opts := DefaultOptions()
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
	opts := DefaultOptions()
	transport := NewSubprocessTransport("", opts)

	err := transport.Write(`{"type":"user","message":{"content":"hello"}}`)
	if err == nil {
		t.Error("expected error when writing to non-ready transport")
	}

	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestSubprocessTransport_Close_NotConnected(t *testing.T) {
	opts := DefaultOptions()
	transport := NewSubprocessTransport("", opts)

	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSubprocessTransport_Close_Idempotent(t *testing.T) {
	opts := DefaultOptions()
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

	opts := DefaultOptions()
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

	opts := DefaultOptions()
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

	opts := DefaultOptions()
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

	opts := DefaultOptions()
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
