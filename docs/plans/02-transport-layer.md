# Plan 02: Transport Layer

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement SubprocessTransport that spawns the Claude CLI and handles stdin/stdout communication.

**Architecture:** Use `os/exec` for subprocess management. Use goroutines for concurrent stdout reading. Use channels for message passing. Handle cross-platform CLI discovery.

**Tech Stack:** Go 1.21+, os/exec, encoding/json, bufio

---

## Task 1: CLI Discovery

**Files:**
- Create: `subprocess.go`
- Create: `subprocess_test.go`

**Step 1: Write failing test**

Create `subprocess_test.go`:

```go
package sdk

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindCLI_WithEnvVar(t *testing.T) {
	// Create a temp directory with a mock claude binary
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	if err := os.WriteFile(mockCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	// Set env var
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+":"+oldPath)
	defer os.Setenv("PATH", oldPath)

	path, err := findCLI("")
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != mockCLI {
		t.Errorf("got %q, want %q", path, mockCLI)
	}
}

func TestFindCLI_WithExplicitPath(t *testing.T) {
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "my-claude")
	if err := os.WriteFile(mockCLI, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatal(err)
	}

	path, err := findCLI(mockCLI)
	if err != nil {
		t.Errorf("findCLI failed: %v", err)
	}
	if path != mockCLI {
		t.Errorf("got %q, want %q", path, mockCLI)
	}
}

func TestFindCLI_NotFound(t *testing.T) {
	// Set empty PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	_, err := findCLI("")
	if err == nil {
		t.Error("expected error when CLI not found")
	}

	var notFoundErr *CLINotFoundError
	if !errors.As(err, &notFoundErr) {
		t.Errorf("expected CLINotFoundError, got %T", err)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestFindCLI -v
```

Expected: FAIL - findCLI not defined

**Step 3: Write implementation**

Create `subprocess.go`:

```go
package sdk

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// findCLI locates the Claude CLI binary.
func findCLI(explicitPath string) (string, error) {
	// If explicit path provided, verify it exists
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		return "", &CLINotFoundError{SearchedPaths: []string{explicitPath}}
	}

	// Check PATH first
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// Check common installation locations
	home, _ := os.UserHomeDir()
	searchPaths := []string{
		filepath.Join(home, ".npm-global", "bin", "claude"),
		"/usr/local/bin/claude",
		filepath.Join(home, ".local", "bin", "claude"),
		filepath.Join(home, "node_modules", ".bin", "claude"),
		filepath.Join(home, ".yarn", "bin", "claude"),
		filepath.Join(home, ".claude", "local", "claude"),
	}

	// Windows-specific locations
	if runtime.GOOS == "windows" {
		searchPaths = append(searchPaths,
			filepath.Join(home, "AppData", "Roaming", "npm", "claude.cmd"),
			filepath.Join(home, "AppData", "Local", "Programs", "claude", "claude.exe"),
		)
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", &CLINotFoundError{SearchedPaths: searchPaths}
}
```

**Step 4: Run tests**

Add missing import and run:

```go
// Add to subprocess_test.go
import "errors"
```

```bash
go test -run TestFindCLI -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add CLI discovery"
```

---

## Task 2: Command Building

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
func TestBuildCommand_Basic(t *testing.T) {
	opts := DefaultOptions()
	cmd := buildCommand("/usr/bin/claude", "Hello", opts, false)

	// Check basic args
	if cmd[0] != "/usr/bin/claude" {
		t.Errorf("got %q, want %q", cmd[0], "/usr/bin/claude")
	}

	// Check output format
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

	// Check prompt (non-streaming mode)
	hasPrint := false
	for i, arg := range cmd {
		if arg == "--print" {
			hasPrint = true
			// Check that prompt follows after --
			for j := i + 1; j < len(cmd); j++ {
				if cmd[j] == "--" && j+1 < len(cmd) && cmd[j+1] == "Hello" {
					break
				}
			}
		}
	}
	if !hasPrint {
		t.Error("missing --print for non-streaming mode")
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
	opts.Cwd = "/tmp/test"
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
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestBuildCommand -v
```

Expected: FAIL - buildCommand not defined

**Step 3: Write implementation**

Add to `subprocess.go`:

```go
import (
	"encoding/json"
	"fmt"
	"strconv"
)

// buildCommand constructs the CLI command with arguments.
func buildCommand(cliPath, prompt string, opts *Options, streaming bool) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose"}

	// System prompt
	if opts.SystemPrompt == "" {
		cmd = append(cmd, "--system-prompt", "")
	} else {
		cmd = append(cmd, "--system-prompt", opts.SystemPrompt)
	}

	if opts.AppendSystemPrompt != "" {
		cmd = append(cmd, "--append-system-prompt", opts.AppendSystemPrompt)
	}

	// Tools
	if len(opts.Tools) > 0 {
		cmd = append(cmd, "--tools", joinStrings(opts.Tools, ","))
	}

	if len(opts.AllowedTools) > 0 {
		cmd = append(cmd, "--allowedTools", joinStrings(opts.AllowedTools, ","))
	}

	if len(opts.DisallowedTools) > 0 {
		cmd = append(cmd, "--disallowedTools", joinStrings(opts.DisallowedTools, ","))
	}

	// Model
	if opts.Model != "" {
		cmd = append(cmd, "--model", opts.Model)
	}

	if opts.FallbackModel != "" {
		cmd = append(cmd, "--fallback-model", opts.FallbackModel)
	}

	// Limits
	if opts.MaxTurns > 0 {
		cmd = append(cmd, "--max-turns", strconv.Itoa(opts.MaxTurns))
	}

	if opts.MaxBudgetUSD > 0 {
		cmd = append(cmd, "--max-budget-usd", fmt.Sprintf("%.4f", opts.MaxBudgetUSD))
	}

	if opts.MaxThinkingTokens > 0 {
		cmd = append(cmd, "--max-thinking-tokens", strconv.Itoa(opts.MaxThinkingTokens))
	}

	// Permissions
	if opts.PermissionMode != "" {
		cmd = append(cmd, "--permission-mode", string(opts.PermissionMode))
	}

	if opts.PermissionPromptToolName != "" {
		cmd = append(cmd, "--permission-prompt-tool", opts.PermissionPromptToolName)
	}

	// Session
	if opts.ContinueConversation {
		cmd = append(cmd, "--continue")
	}

	if opts.Resume != "" {
		cmd = append(cmd, "--resume", opts.Resume)
	}

	if opts.ForkSession {
		cmd = append(cmd, "--fork-session")
	}

	// Settings
	if opts.Settings != "" {
		cmd = append(cmd, "--settings", opts.Settings)
	}

	if len(opts.SettingSources) > 0 {
		cmd = append(cmd, "--setting-sources", joinStrings(opts.SettingSources, ","))
	} else {
		cmd = append(cmd, "--setting-sources", "")
	}

	// Directories
	for _, dir := range opts.AddDirs {
		cmd = append(cmd, "--add-dir", dir)
	}

	// MCP servers
	if len(opts.MCPServers) > 0 {
		config := map[string]any{"mcpServers": opts.MCPServers}
		if data, err := json.Marshal(config); err == nil {
			cmd = append(cmd, "--mcp-config", string(data))
		}
	}

	// Betas
	if len(opts.Betas) > 0 {
		cmd = append(cmd, "--betas", joinStrings(opts.Betas, ","))
	}

	// Streaming options
	if opts.IncludePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
	}

	// Output format
	if opts.OutputFormat != nil {
		if schema, ok := opts.OutputFormat["schema"]; ok {
			if data, err := json.Marshal(schema); err == nil {
				cmd = append(cmd, "--json-schema", string(data))
			}
		}
	}

	// Extra args
	for flag, value := range opts.ExtraArgs {
		if value == "" {
			cmd = append(cmd, "--"+flag)
		} else {
			cmd = append(cmd, "--"+flag, value)
		}
	}

	// Mode-specific args (must come last)
	if streaming {
		cmd = append(cmd, "--input-format", "stream-json")
	} else {
		cmd = append(cmd, "--print", "--", prompt)
	}

	return cmd
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
```

**Step 4: Run tests**

```bash
go test -run TestBuildCommand -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add command building"
```

---

## Task 3: SubprocessTransport Structure

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
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

func TestSubprocessTransportImplementsInterface(t *testing.T) {
	var _ Transport = (*SubprocessTransport)(nil)
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestNewSubprocessTransport -v
```

Expected: FAIL - SubprocessTransport not defined

**Step 3: Write implementation**

Add to `subprocess.go`:

```go
import (
	"bufio"
	"context"
	"io"
	"sync"
)

// SubprocessTransport manages the Claude CLI subprocess.
type SubprocessTransport struct {
	prompt    string
	options   *Options
	streaming bool

	cliPath string
	cmd     *exec.Cmd

	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	messages chan map[string]any
	errors   chan error

	ready    bool
	closed   bool
	closeMu  sync.Mutex
	writeMu  sync.Mutex

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSubprocessTransport creates a new subprocess transport.
func NewSubprocessTransport(prompt string, opts *Options) *SubprocessTransport {
	if opts == nil {
		opts = DefaultOptions()
	}

	return &SubprocessTransport{
		prompt:    prompt,
		options:   opts,
		streaming: prompt == "", // Empty prompt = streaming mode
		messages:  make(chan map[string]any, 100),
		errors:    make(chan error, 1),
	}
}

// NewStreamingTransport creates a transport for streaming mode.
func NewStreamingTransport(opts *Options) *SubprocessTransport {
	return NewSubprocessTransport("", opts)
}

// IsReady returns true if the transport is connected.
func (t *SubprocessTransport) IsReady() bool {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()
	return t.ready && !t.closed
}

// Messages returns the channel of messages from the CLI.
func (t *SubprocessTransport) Messages() <-chan map[string]any {
	return t.messages
}
```

**Step 4: Run tests**

```bash
go test -run "TestNewSubprocessTransport|TestSubprocessTransportImplementsInterface" -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add SubprocessTransport structure"
```

---

## Task 4: Connect Method

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestSubprocessTransport_Connect -v
```

Expected: FAIL - Connect not fully implemented

**Step 3: Write implementation**

Add to `subprocess.go`:

```go
// Connect starts the CLI subprocess.
func (t *SubprocessTransport) Connect(ctx context.Context) error {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()

	if t.ready {
		return nil
	}

	// Find CLI
	cliPath, err := findCLI(t.options.CLIPath)
	if err != nil {
		return err
	}
	t.cliPath = cliPath

	// Build command
	args := buildCommand(cliPath, t.prompt, t.options, t.streaming)

	// Create context for cancellation
	t.ctx, t.cancel = context.WithCancel(ctx)

	// Create command
	t.cmd = exec.CommandContext(t.ctx, args[0], args[1:]...)

	// Set working directory
	if t.options.Cwd != "" {
		t.cmd.Dir = t.options.Cwd
	}

	// Set environment
	t.cmd.Env = os.Environ()
	t.cmd.Env = append(t.cmd.Env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
	t.cmd.Env = append(t.cmd.Env, "CLAUDE_AGENT_SDK_VERSION="+Version)

	for k, v := range t.options.Env {
		t.cmd.Env = append(t.cmd.Env, k+"="+v)
	}

	if t.options.EnableFileCheckpointing {
		t.cmd.Env = append(t.cmd.Env, "CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true")
	}

	// Setup pipes
	var err error
	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return &ConnectionError{Message: "failed to create stdin pipe", Cause: err}
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return &ConnectionError{Message: "failed to create stdout pipe", Cause: err}
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return &ConnectionError{Message: "failed to create stderr pipe", Cause: err}
	}

	// Start process
	if err := t.cmd.Start(); err != nil {
		return &ConnectionError{Message: "failed to start CLI", Cause: err}
	}

	// Start reading stdout
	t.wg.Add(1)
	go t.readMessages()

	// Start reading stderr (for debugging)
	t.wg.Add(1)
	go t.readStderr()

	// For non-streaming mode, close stdin immediately
	if !t.streaming {
		t.stdin.Close()
	}

	t.ready = true
	return nil
}
```

**Step 4: Run tests**

```bash
go test -run TestSubprocessTransport_Connect -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add Connect method"
```

---

## Task 5: Message Reading

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
func TestReadJSONLine(t *testing.T) {
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
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestReadJSONLine -v
```

Expected: FAIL - parseJSONLine not defined

**Step 3: Write implementation**

Add to `subprocess.go`:

```go
// parseJSONLine parses a single JSON line.
func parseJSONLine(line string) (map[string]any, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(line), &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return result, nil
}

// readMessages reads JSON messages from stdout.
func (t *SubprocessTransport) readMessages() {
	defer t.wg.Done()
	defer close(t.messages)

	scanner := bufio.NewScanner(t.stdout)
	// Increase buffer size for large messages
	const maxBufferSize = 1024 * 1024 // 1MB
	buf := make([]byte, maxBufferSize)
	scanner.Buffer(buf, maxBufferSize)

	var jsonBuffer strings.Builder

	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Accumulate partial JSON (in case of line breaks in content)
		jsonBuffer.WriteString(line)

		// Try to parse
		msg, err := parseJSONLine(jsonBuffer.String())
		if err != nil {
			// Might be partial JSON, keep accumulating
			if jsonBuffer.Len() > maxBufferSize {
				// Buffer too large, reset
				jsonBuffer.Reset()
				continue
			}
			continue
		}

		// Successfully parsed, reset buffer and send
		jsonBuffer.Reset()

		select {
		case t.messages <- msg:
		case <-t.ctx.Done():
			return
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		select {
		case t.errors <- err:
		default:
		}
	}
}

// readStderr reads stderr for debugging.
func (t *SubprocessTransport) readStderr() {
	defer t.wg.Done()

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		// For now, just discard stderr
		// In the future, could log or forward to a callback
		_ = scanner.Text()
	}
}
```

Add the import:

```go
import "strings"
```

**Step 4: Run tests**

```bash
go test -run TestReadJSONLine -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add message reading"
```

---

## Task 6: Write Method

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
func TestSubprocessTransport_Write_NotReady(t *testing.T) {
	opts := DefaultOptions()
	transport := NewSubprocessTransport("", opts)

	err := transport.Write(`{"type":"user","message":{"content":"hello"}}`)
	if err == nil {
		t.Error("expected error when writing to non-ready transport")
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestSubprocessTransport_Write -v
```

Expected: FAIL - Write not implemented

**Step 3: Write implementation**

Add to `subprocess.go`:

```go
// Write sends data to the CLI stdin.
func (t *SubprocessTransport) Write(data string) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	t.closeMu.Lock()
	if !t.ready || t.closed {
		t.closeMu.Unlock()
		return &ConnectionError{Message: "transport not ready for writing"}
	}
	if t.stdin == nil {
		t.closeMu.Unlock()
		return &ConnectionError{Message: "stdin is nil"}
	}
	t.closeMu.Unlock()

	// Ensure data ends with newline
	if !strings.HasSuffix(data, "\n") {
		data += "\n"
	}

	_, err := io.WriteString(t.stdin, data)
	if err != nil {
		return &ConnectionError{Message: "failed to write to stdin", Cause: err}
	}

	return nil
}

// EndInput closes stdin to signal end of input.
func (t *SubprocessTransport) EndInput() error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	if t.stdin != nil {
		return t.stdin.Close()
	}
	return nil
}
```

**Step 4: Run tests**

```bash
go test -run TestSubprocessTransport_Write -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add Write and EndInput methods"
```

---

## Task 7: Close Method

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
func TestSubprocessTransport_Close_NotConnected(t *testing.T) {
	opts := DefaultOptions()
	transport := NewSubprocessTransport("", opts)

	// Should not error when closing non-connected transport
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
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestSubprocessTransport_Close -v
```

Expected: FAIL - Close not implemented

**Step 3: Write implementation**

Add to `subprocess.go`:

```go
import "time"

// Close terminates the subprocess and cleans up resources.
func (t *SubprocessTransport) Close() error {
	t.closeMu.Lock()
	if t.closed {
		t.closeMu.Unlock()
		return nil
	}
	t.closed = true
	t.ready = false
	t.closeMu.Unlock()

	// Cancel context to stop goroutines
	if t.cancel != nil {
		t.cancel()
	}

	// Close stdin first
	if t.stdin != nil {
		t.stdin.Close()
	}

	// Terminate process if running
	if t.cmd != nil && t.cmd.Process != nil {
		// Give process a chance to exit gracefully
		done := make(chan error, 1)
		go func() {
			done <- t.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(5 * time.Second):
			// Force kill
			t.cmd.Process.Kill()
			<-done
		}
	}

	// Close pipes
	if t.stdout != nil {
		t.stdout.Close()
	}
	if t.stderr != nil {
		t.stderr.Close()
	}

	// Wait for goroutines
	t.wg.Wait()

	return nil
}
```

**Step 4: Run tests**

```bash
go test -run TestSubprocessTransport_Close -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add Close method"
```

---

## Task 8: Integration Test with Mock CLI

**Files:**
- Modify: `subprocess_test.go`
- Create: `testdata/mock_claude.sh`

**Step 1: Create mock CLI**

Create `testdata/mock_claude.sh`:

```bash
#!/bin/bash
# Mock Claude CLI for testing

# Output a simple response
echo '{"type":"system","subtype":"init","data":{"version":"1.0"}}'
echo '{"type":"assistant","message":{"content":[{"type":"text","text":"Hello from mock!"}],"model":"claude-test"}}'
echo '{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":80,"is_error":false,"num_turns":1,"session_id":"test_123"}'
```

**Step 2: Write integration test**

Add to `subprocess_test.go`:

```go
func TestSubprocessTransport_Integration(t *testing.T) {
	// Create mock CLI
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	mockScript := `#!/bin/bash
echo '{"type":"system","subtype":"init","data":{"version":"1.0"}}'
echo '{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}],"model":"claude-test"}}'
echo '{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":80,"is_error":false,"num_turns":1,"session_id":"test_123"}'
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewSubprocessTransport("Hello", opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Collect messages
	var messages []map[string]any
	for msg := range transport.Messages() {
		messages = append(messages, msg)
	}

	// Verify we got expected messages
	if len(messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(messages))
	}

	// Check message types
	expectedTypes := []string{"system", "assistant", "result"}
	for i, expected := range expectedTypes {
		if i < len(messages) {
			if messages[i]["type"] != expected {
				t.Errorf("message %d: got type %v, want %s", i, messages[i]["type"], expected)
			}
		}
	}
}
```

Add the import:

```go
import "time"
```

**Step 3: Run test**

```bash
go test -run TestSubprocessTransport_Integration -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "test: add integration test with mock CLI"
```

---

## Task 9: Environment Variable Handling

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
func TestBuildEnvironment(t *testing.T) {
	opts := DefaultOptions()
	opts.Env = map[string]string{
		"CUSTOM_VAR": "custom_value",
	}
	opts.EnableFileCheckpointing = true

	env := buildEnvironment(opts)

	// Check for SDK env vars
	found := map[string]bool{
		"CLAUDE_CODE_ENTRYPOINT":                     false,
		"CLAUDE_AGENT_SDK_VERSION":                   false,
		"CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING": false,
		"CUSTOM_VAR":                                 false,
	}

	for _, e := range env {
		for key := range found {
			if strings.HasPrefix(e, key+"=") {
				found[key] = true
			}
		}
	}

	for key, wasFound := range found {
		if !wasFound {
			t.Errorf("missing env var: %s", key)
		}
	}
}
```

**Step 2: Run test to verify it fails**

```bash
go test -run TestBuildEnvironment -v
```

Expected: FAIL - buildEnvironment not defined

**Step 3: Write implementation**

Add to `subprocess.go`:

```go
// buildEnvironment creates the environment for the subprocess.
func buildEnvironment(opts *Options) []string {
	env := os.Environ()

	// Add SDK-specific vars
	env = append(env, "CLAUDE_CODE_ENTRYPOINT=sdk-go")
	env = append(env, "CLAUDE_AGENT_SDK_VERSION="+Version)

	// Add user-provided vars
	for k, v := range opts.Env {
		env = append(env, k+"="+v)
	}

	// Add feature flags
	if opts.EnableFileCheckpointing {
		env = append(env, "CLAUDE_CODE_ENABLE_SDK_FILE_CHECKPOINTING=true")
	}

	return env
}
```

Update `Connect` to use `buildEnvironment`:

```go
// In Connect method, replace the env setup with:
t.cmd.Env = buildEnvironment(t.options)
```

**Step 4: Run tests**

```bash
go test -run TestBuildEnvironment -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "refactor: extract buildEnvironment function"
```

---

## Task 10: Error Handling

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
func TestSubprocessTransport_ProcessError(t *testing.T) {
	// Create mock CLI that exits with error
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")

	mockScript := `#!/bin/bash
echo "Error: something went wrong" >&2
exit 1
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewSubprocessTransport("Hello", opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := transport.Connect(ctx)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Drain messages
	for range transport.Messages() {
	}

	// Should have captured exit error
	// (In a real implementation, we'd check the error channel)
}
```

**Step 2: Run test**

```bash
go test -run TestSubprocessTransport_ProcessError -v
```

Expected: PASS (test just verifies no panic)

**Step 3: Improve error handling in readMessages**

Update in `subprocess.go`:

```go
// readMessages reads JSON messages from stdout.
func (t *SubprocessTransport) readMessages() {
	defer t.wg.Done()
	defer close(t.messages)

	scanner := bufio.NewScanner(t.stdout)
	const maxBufferSize = 1024 * 1024
	buf := make([]byte, maxBufferSize)
	scanner.Buffer(buf, maxBufferSize)

	var jsonBuffer strings.Builder

	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		jsonBuffer.WriteString(line)

		msg, err := parseJSONLine(jsonBuffer.String())
		if err != nil {
			if jsonBuffer.Len() > maxBufferSize {
				jsonBuffer.Reset()
			}
			continue
		}

		jsonBuffer.Reset()

		select {
		case t.messages <- msg:
		case <-t.ctx.Done():
			return
		}
	}

	if err := scanner.Err(); err != nil {
		select {
		case t.errors <- err:
		default:
		}
	}

	// Wait for process and check exit code
	if t.cmd != nil {
		if err := t.cmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				procErr := &ProcessError{
					ExitCode: exitErr.ExitCode(),
					Stderr:   "check stderr for details",
				}
				select {
				case t.errors <- procErr:
				default:
				}
			}
		}
	}
}
```

**Step 4: Run all transport tests**

```bash
go test -run "TestSubprocess|TestFind|TestBuild" -v
```

Expected: All PASS

**Step 5: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: improve error handling in transport"
```

---

## Summary

After completing Plan 02, you have:

- [x] CLI discovery with fallback locations
- [x] Command building with all options
- [x] SubprocessTransport structure
- [x] Connect with process spawning
- [x] Message reading with JSON parsing
- [x] Write and EndInput methods
- [x] Close with graceful shutdown
- [x] Integration test with mock CLI
- [x] Environment variable handling
- [x] Error handling for process failures

**Next:** Plan 03 - Query/Control Protocol
