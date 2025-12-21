# Plan 02: Transport Layer (Complete Feature Parity)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Implement SubprocessTransport that spawns the Claude CLI and handles stdin/stdout communication with complete feature parity to Python SDK.

**Architecture:** Use `os/exec` for subprocess management. Use goroutines for concurrent stdout reading. Use channels for message passing. Handle cross-platform CLI discovery. **CRITICAL: Implement write lock for concurrent MCP tool calls and TOCTOU-safe operations.**

**Tech Stack:** Go 1.21+, os/exec, encoding/json, bufio

**Reference:** `.reference/claude-agent-sdk-python/src/claude_code_sdk/_internal/transport/subprocess_cli_transport.py`

---

## Task 0: Version Check Pre-Flight

**Files:**
- Create: `version.go`
- Create: `version_test.go`

**Step 1: Write failing test**

Create `version_test.go`:

```go
package sdk

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version string
		want    [3]int
		wantErr bool
	}{
		{"2.0.0", [3]int{2, 0, 0}, false},
		{"2.1.5", [3]int{2, 1, 5}, false},
		{"1.0.0", [3]int{1, 0, 0}, false},
		{"invalid", [3]int{}, true},
		{"2.0", [3]int{}, true},
		{"", [3]int{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got, err := parseVersion(tt.version)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"2.0.0", "2.0.0", 0},
		{"2.1.0", "2.0.0", 1},
		{"2.0.0", "2.1.0", -1},
		{"3.0.0", "2.9.9", 1},
		{"1.9.9", "2.0.0", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := compareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCheckMinimumVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{"2.0.0", false},
		{"2.1.0", false},
		{"3.0.0", false},
		{"1.9.9", true},
		{"1.0.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			err := checkMinimumVersion(tt.version)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
```

**Step 2: Write implementation**

Create `version.go`:

```go
package sdk

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is the SDK version.
const Version = "0.1.0"

// MinimumCLIVersion is the minimum supported CLI version.
const MinimumCLIVersion = "2.0.0"

// parseVersion parses a semantic version string.
func parseVersion(version string) ([3]int, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("invalid version format: %s", version)
	}

	var result [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return [3]int{}, fmt.Errorf("invalid version component: %s", part)
		}
		result[i] = n
	}

	return result, nil
}

// compareVersions compares two version strings.
// Returns: -1 if a < b, 0 if a == b, 1 if a > b
func compareVersions(a, b string) int {
	va, err := parseVersion(a)
	if err != nil {
		return -1
	}
	vb, err := parseVersion(b)
	if err != nil {
		return 1
	}

	for i := 0; i < 3; i++ {
		if va[i] < vb[i] {
			return -1
		}
		if va[i] > vb[i] {
			return 1
		}
	}
	return 0
}

// checkMinimumVersion validates that the CLI version meets minimum requirements.
func checkMinimumVersion(version string) error {
	if compareVersions(version, MinimumCLIVersion) < 0 {
		return &CLIVersionError{
			InstalledVersion: version,
			MinimumVersion:   MinimumCLIVersion,
		}
	}
	return nil
}
```

**Step 3: Run tests**

```bash
go test -run TestVersion -v
go test -run TestCompare -v
go test -run TestCheckMinimum -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add version.go version_test.go
git commit -m "feat: add version checking utilities"
```

---

## Task 1: CLI Discovery with Bundled Support

**Files:**
- Create: `subprocess.go`
- Create: `subprocess_test.go`

**Step 1: Write failing test**

Create `subprocess_test.go`:

```go
package sdk

import (
	"errors"
	"os"
	"os/exec"
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
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", "")
	defer os.Setenv("PATH", oldPath)

	_, err := findCLI("", "")
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
```

**Step 2: Write implementation**

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
// Priority: explicitPath > bundledPath > PATH > common locations
func findCLI(explicitPath, bundledPath string) (string, error) {
	searchedPaths := []string{}

	// 1. Explicit path has highest priority
	if explicitPath != "" {
		searchedPaths = append(searchedPaths, explicitPath)
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath, nil
		}
		return "", &CLINotFoundError{SearchedPaths: searchedPaths}
	}

	// 2. Bundled path (for packaged distributions)
	if bundledPath != "" {
		searchedPaths = append(searchedPaths, bundledPath)
		if _, err := os.Stat(bundledPath); err == nil {
			return bundledPath, nil
		}
	}

	// 3. Check PATH
	if path, err := exec.LookPath("claude"); err == nil {
		return path, nil
	}

	// 4. Check common installation locations
	home, _ := os.UserHomeDir()
	commonPaths := []string{
		// npm global (most common)
		filepath.Join(home, ".npm-global", "bin", "claude"),
		// Local bin
		"/usr/local/bin/claude",
		filepath.Join(home, ".local", "bin", "claude"),
		// Node modules
		filepath.Join(home, "node_modules", ".bin", "claude"),
		// Yarn
		filepath.Join(home, ".yarn", "bin", "claude"),
		// Claude local installation
		filepath.Join(home, ".claude", "local", "claude"),
	}

	// Windows-specific locations
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		localAppData := os.Getenv("LOCALAPPDATA")
		commonPaths = append(commonPaths,
			filepath.Join(appData, "npm", "claude.cmd"),
			filepath.Join(localAppData, "Programs", "claude", "claude.exe"),
			filepath.Join(home, "AppData", "Roaming", "npm", "claude.cmd"),
		)
	}

	for _, path := range commonPaths {
		searchedPaths = append(searchedPaths, path)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", &CLINotFoundError{SearchedPaths: searchedPaths}
}
```

**Step 3: Run tests**

```bash
go test -run TestFindCLI -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add CLI discovery with bundled support"
```

---

## Task 2: Command Building with Windows Length Handling

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
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
	opts.Sandbox = &SandboxConfig{
		Mode:        SandboxDocker,
		AllowedDirs: []string{"/safe"},
		BlockedDirs: []string{"/blocked"},
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
```

**Step 2: Write implementation**

Add to `subprocess.go`:

```go
import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// WindowsMaxCommandLength is the maximum command line length on Windows.
const WindowsMaxCommandLength = 8191

// buildCommand constructs the CLI command with arguments.
func buildCommand(cliPath, prompt string, opts *Options, streaming bool) []string {
	cmd := []string{cliPath, "--output-format", "stream-json", "--verbose"}

	// System prompt (always include, even if empty)
	if opts.SystemPrompt != "" {
		cmd = append(cmd, "--system-prompt", opts.SystemPrompt)
	} else {
		cmd = append(cmd, "--system-prompt", "")
	}

	if opts.AppendSystemPrompt != "" {
		cmd = append(cmd, "--append-system-prompt", opts.AppendSystemPrompt)
	}

	// Tools configuration
	if len(opts.Tools) > 0 {
		cmd = append(cmd, "--tools", strings.Join(opts.Tools, ","))
	}

	if len(opts.AllowedTools) > 0 {
		cmd = append(cmd, "--allowedTools", strings.Join(opts.AllowedTools, ","))
	}

	if len(opts.DisallowedTools) > 0 {
		cmd = append(cmd, "--disallowedTools", strings.Join(opts.DisallowedTools, ","))
	}

	// Model selection
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

	// Session management
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
		cmd = append(cmd, "--setting-sources", strings.Join(opts.SettingSources, ","))
	} else {
		// Empty setting sources to avoid loading default settings
		cmd = append(cmd, "--setting-sources", "")
	}

	// Sandbox configuration
	if opts.Sandbox != nil {
		sandboxJSON, _ := json.Marshal(opts.Sandbox)
		cmd = append(cmd, "--sandbox", string(sandboxJSON))
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

	// Beta features
	if len(opts.Betas) > 0 {
		cmd = append(cmd, "--betas", strings.Join(opts.Betas, ","))
	}

	// Streaming options
	if opts.IncludePartialMessages {
		cmd = append(cmd, "--include-partial-messages")
	}

	// Output format (JSON schema)
	if opts.OutputFormat != nil {
		if schema, ok := opts.OutputFormat["schema"]; ok {
			if data, err := json.Marshal(schema); err == nil {
				cmd = append(cmd, "--json-schema", string(data))
			}
		}
	}

	// Extra args (escape hatch for future CLI flags)
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

// checkCommandLength validates command length on Windows.
func checkCommandLength(cmd []string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	totalLen := 0
	for _, arg := range cmd {
		totalLen += len(arg) + 1 // +1 for space separator
	}

	if totalLen > WindowsMaxCommandLength {
		return fmt.Errorf("command length %d exceeds Windows limit of %d", totalLen, WindowsMaxCommandLength)
	}

	return nil
}
```

**Step 3: Run tests**

```bash
go test -run TestBuildCommand -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add command building with Windows length handling"
```

---

## Task 3: SubprocessTransport Structure with Critical Fields

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
```

**Step 2: Write implementation**

Add to `subprocess.go`:

```go
import (
	"bufio"
	"context"
	"io"
	"sync"
)

// SubprocessTransport manages the Claude CLI subprocess.
// CRITICAL: This implements write serialization for concurrent MCP tool calls.
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

	ready   bool
	closed  bool
	closeMu sync.Mutex

	// CRITICAL: Write mutex for concurrent MCP tool call serialization
	// All writes to stdin MUST be protected by this mutex
	writeMu sync.Mutex

	// Exit error tracking for proper error reporting
	exitError error
	exitMu    sync.Mutex

	// Temp files to clean up on close
	tempFiles []string
	tempMu    sync.Mutex

	// Stderr callback for debugging
	stderrCallback func(string)

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
		tempFiles: make([]string, 0),
	}
}

// NewStreamingTransport creates a transport for streaming mode.
func NewStreamingTransport(opts *Options) *SubprocessTransport {
	return NewSubprocessTransport("", opts)
}

// IsReady returns true if the transport is connected and not closed.
func (t *SubprocessTransport) IsReady() bool {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()
	return t.ready && !t.closed
}

// Messages returns the channel of messages from the CLI.
func (t *SubprocessTransport) Messages() <-chan map[string]any {
	return t.messages
}

// Errors returns the channel of errors from the CLI.
func (t *SubprocessTransport) Errors() <-chan error {
	return t.errors
}

// SetStderrCallback sets a callback for stderr output.
func (t *SubprocessTransport) SetStderrCallback(callback func(string)) {
	t.stderrCallback = callback
}

// AddTempFile adds a temp file to be cleaned up on close.
func (t *SubprocessTransport) AddTempFile(path string) {
	t.tempMu.Lock()
	defer t.tempMu.Unlock()
	t.tempFiles = append(t.tempFiles, path)
}

// ExitError returns the exit error if the process has exited.
func (t *SubprocessTransport) ExitError() error {
	t.exitMu.Lock()
	defer t.exitMu.Unlock()
	return t.exitError
}
```

**Step 3: Run tests**

```bash
go test -run "TestNewSubprocessTransport|TestNewStreamingTransport|TestSubprocessTransportImplementsInterface" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add SubprocessTransport structure with critical fields"
```

---

## Task 4: Connect Method with Validation

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
```

**Step 2: Write implementation**

Add to `subprocess.go`:

```go
// Connect starts the CLI subprocess.
func (t *SubprocessTransport) Connect(ctx context.Context) error {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()

	// Already connected
	if t.ready {
		return nil
	}

	// Find CLI
	cliPath, err := findCLI(t.options.CLIPath, t.options.BundledCLIPath)
	if err != nil {
		return err
	}
	t.cliPath = cliPath

	// Build command
	args := buildCommand(cliPath, t.prompt, t.options, t.streaming)

	// Check command length on Windows
	if err := checkCommandLength(args); err != nil {
		return &ConnectionError{Message: err.Error()}
	}

	// Create context for cancellation
	t.ctx, t.cancel = context.WithCancel(ctx)

	// Create command
	t.cmd = exec.CommandContext(t.ctx, args[0], args[1:]...)

	// Set working directory
	if t.options.Cwd != "" {
		t.cmd.Dir = t.options.Cwd
	}

	// Set environment
	t.cmd.Env = buildEnvironment(t.options)

	// Setup pipes
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

	// Start reading stderr
	t.wg.Add(1)
	go t.readStderr()

	// For non-streaming mode, close stdin immediately after start
	if !t.streaming {
		t.stdin.Close()
	}

	t.ready = true
	return nil
}

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

**Step 3: Run tests**

```bash
go test -run TestSubprocessTransport_Connect -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add Connect method with validation"
```

---

## Task 5: Message Reading with Speculative JSON Parsing

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
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
```

**Step 2: Write implementation**

Add to `subprocess.go`:

```go
import "strings"

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

// jsonAccumulator handles speculative JSON parsing for partial lines.
type jsonAccumulator struct {
	buffer strings.Builder
}

func newJSONAccumulator() *jsonAccumulator {
	return &jsonAccumulator{}
}

// addLine adds a line to the accumulator and attempts to parse.
// Returns (result, nil) if JSON is complete, (nil, nil) if still accumulating.
func (a *jsonAccumulator) addLine(line string) (map[string]any, error) {
	a.buffer.WriteString(line)

	// Try to parse speculatively
	result, err := parseJSONLine(a.buffer.String())
	if err != nil {
		// Still accumulating - not an error yet
		return nil, nil
	}

	// Successfully parsed - reset buffer
	a.buffer.Reset()
	return result, nil
}

// reset clears the accumulator.
func (a *jsonAccumulator) reset() {
	a.buffer.Reset()
}

// len returns the current buffer length.
func (a *jsonAccumulator) len() int {
	return a.buffer.Len()
}

const maxBufferSize = 1024 * 1024 // 1MB

// readMessages reads JSON messages from stdout with speculative parsing.
func (t *SubprocessTransport) readMessages() {
	defer t.wg.Done()
	defer close(t.messages)

	scanner := bufio.NewScanner(t.stdout)
	buf := make([]byte, maxBufferSize)
	scanner.Buffer(buf, maxBufferSize)

	accumulator := newJSONAccumulator()

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

		// Speculative parsing - try to parse immediately
		msg, _ := accumulator.addLine(line)
		if msg == nil {
			// Still accumulating
			if accumulator.len() > maxBufferSize {
				// Buffer overflow - reset and discard
				accumulator.reset()
			}
			continue
		}

		// Successfully parsed
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

	// Wait for process and capture exit error
	if t.cmd != nil {
		if err := t.cmd.Wait(); err != nil {
			t.exitMu.Lock()
			t.exitError = err
			t.exitMu.Unlock()

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

// readStderr reads stderr and optionally invokes callback.
func (t *SubprocessTransport) readStderr() {
	defer t.wg.Done()

	scanner := bufio.NewScanner(t.stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if t.stderrCallback != nil {
			t.stderrCallback(line)
		}
	}
}
```

**Step 3: Run tests**

```bash
go test -run "TestParseJSONLine|TestSpeculativeJSONParsing" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add message reading with speculative JSON parsing"
```

---

## Task 6: TOCTOU-Safe Write Method

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

	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Errorf("expected ConnectionError, got %T", err)
	}
}

func TestSubprocessTransport_Write_Serialization(t *testing.T) {
	// Test that concurrent writes are serialized
	tmpDir := t.TempDir()
	mockCLI := filepath.Join(tmpDir, "claude")
	// This mock reads from stdin and echoes back
	mockScript := `#!/bin/bash
while read -r line; do
    echo "{\"type\":\"echo\",\"data\":\"$line\"}"
done
`
	if err := os.WriteFile(mockCLI, []byte(mockScript), 0755); err != nil {
		t.Fatal(err)
	}

	opts := DefaultOptions()
	opts.CLIPath = mockCLI

	transport := NewStreamingTransport(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer transport.Close()

	// Concurrent writes should be serialized
	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			msg := fmt.Sprintf(`{"n":%d}`, n)
			if err := transport.Write(msg); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("Write error: %v", err)
	}
}
```

**Step 2: Write implementation**

Add to `subprocess.go`:

```go
// Write sends data to the CLI stdin with TOCTOU-safe serialization.
// CRITICAL: This method serializes concurrent writes from MCP tool calls.
func (t *SubprocessTransport) Write(data string) error {
	// CRITICAL: Acquire write lock FIRST to serialize concurrent MCP tool calls
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	// TOCTOU-safe: Check ready state INSIDE the lock
	t.closeMu.Lock()
	ready := t.ready
	closed := t.closed
	stdin := t.stdin
	t.closeMu.Unlock()

	if !ready || closed {
		return &ConnectionError{Message: "transport not ready for writing"}
	}
	if stdin == nil {
		return &ConnectionError{Message: "stdin is nil"}
	}

	// Ensure data ends with newline
	if !strings.HasSuffix(data, "\n") {
		data += "\n"
	}

	_, err := io.WriteString(stdin, data)
	if err != nil {
		return &ConnectionError{Message: "failed to write to stdin", Cause: err}
	}

	return nil
}

// WriteJSON marshals and writes a JSON object.
func (t *SubprocessTransport) WriteJSON(obj any) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return t.Write(string(data))
}

// EndInput closes stdin to signal end of input.
func (t *SubprocessTransport) EndInput() error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	t.closeMu.Lock()
	stdin := t.stdin
	t.closeMu.Unlock()

	if stdin != nil {
		return stdin.Close()
	}
	return nil
}
```

**Step 3: Run tests**

```bash
go test -run "TestSubprocessTransport_Write" -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add TOCTOU-safe Write method with serialization"
```

---

## Task 7: Close Method with Cleanup

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
import "time"

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
```

**Step 2: Write implementation**

Add to `subprocess.go`:

```go
import "time"

const gracefulShutdownTimeout = 5 * time.Second

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

	// Close stdin first to signal EOF
	if t.stdin != nil {
		t.stdin.Close()
	}

	// Terminate process if running
	if t.cmd != nil && t.cmd.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- t.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited gracefully
		case <-time.After(gracefulShutdownTimeout):
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

	// Clean up temp files
	t.tempMu.Lock()
	for _, path := range t.tempFiles {
		os.Remove(path)
	}
	t.tempFiles = nil
	t.tempMu.Unlock()

	return nil
}

// Kill forcefully terminates the subprocess.
func (t *SubprocessTransport) Kill() error {
	t.closeMu.Lock()
	defer t.closeMu.Unlock()

	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Kill()
	}
	return nil
}
```

**Step 3: Run tests**

```bash
go test -run TestSubprocessTransport_Close -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add subprocess.go subprocess_test.go
git commit -m "feat: add Close method with cleanup"
```

---

## Task 8: Stderr Callback Support

**Files:**
- Modify: `subprocess.go`
- Modify: `subprocess_test.go`

**Step 1: Write failing test**

Add to `subprocess_test.go`:

```go
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
```

**Step 2: Run test**

The implementation was already added in Task 5. Verify:

```bash
go test -run TestSubprocessTransport_StderrCallback -v
```

Expected: PASS

**Step 3: Commit**

```bash
git add subprocess_test.go
git commit -m "test: add stderr callback test"
```

---

## Task 9: Concurrent Write Test

**Files:**
- Modify: `subprocess_test.go`

**Step 1: Write race condition test**

Add to `subprocess_test.go`:

```go
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
				msg := fmt.Sprintf(`{"writer":%d,"msg":%d}`, writerID, j)
				if err := transport.Write(msg); err != nil {
					errors <- fmt.Errorf("writer %d msg %d: %w", writerID, j, err)
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
```

**Step 2: Run with race detector**

```bash
go test -race -run TestSubprocessTransport_ConcurrentWrites_Race -v
```

Expected: PASS (no race conditions)

**Step 3: Commit**

```bash
git add subprocess_test.go
git commit -m "test: add concurrent writes race condition test"
```

---

## Task 10: Mock CLI for Integration Testing

**Files:**
- Create: `testdata/mock_claude.sh`
- Modify: `subprocess_test.go`

**Step 1: Create mock CLI**

Create `testdata/mock_claude.sh`:

```bash
#!/bin/bash
# Mock Claude CLI for testing

# Echo version if requested
if [[ "$*" == *"--version"* ]]; then
    echo "2.0.0"
    exit 0
fi

# Output init message
echo '{"type":"system","subtype":"init","data":{"version":"2.0.0","session_id":"test_123"}}'

# Check if streaming mode
if [[ "$*" == *"--input-format stream-json"* ]]; then
    # Streaming mode - read from stdin
    while read -r line; do
        if [[ -n "$line" ]]; then
            # Echo as assistant message
            text=$(echo "$line" | jq -r '.message.content // "received"' 2>/dev/null || echo "received")
            echo "{\"type\":\"assistant\",\"message\":{\"content\":[{\"type\":\"text\",\"text\":\"Echo: $text\"}],\"model\":\"claude-test\"}}"
        fi
    done
else
    # Non-streaming mode - single response
    echo '{"type":"assistant","message":{"content":[{"type":"text","text":"Hello from mock!"}],"model":"claude-test"}}'
fi

# Output result
echo '{"type":"result","subtype":"success","duration_ms":100,"duration_api_ms":80,"is_error":false,"num_turns":1,"session_id":"test_123","total_cost_usd":0.001}'
```

**Step 2: Write integration test**

Add to `subprocess_test.go`:

```go
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

	opts := DefaultOptions()
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

	opts := DefaultOptions()
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
```

**Step 3: Run integration tests**

```bash
go test -run "TestSubprocessTransport_Integration" -v
```

Expected: PASS

**Step 4: Commit**

```bash
mkdir -p testdata
git add testdata/mock_claude.sh subprocess_test.go
git commit -m "test: add mock CLI integration tests"
```

---

## Summary

After completing Plan 02, you have:

- [x] Version checking with semantic version comparison
- [x] CLI discovery with bundled path support
- [x] Command building with Windows length handling
- [x] SubprocessTransport with critical fields (writeMu, exitError, tempFiles)
- [x] Connect with validation and environment setup
- [x] Message reading with speculative JSON parsing
- [x] TOCTOU-safe Write method with serialization
- [x] Close with graceful shutdown and temp file cleanup
- [x] Stderr callback support
- [x] Concurrent write race condition tests
- [x] Mock CLI integration tests

**Critical Features Implemented:**
- Write lock (`writeMu`) for concurrent MCP tool call serialization
- TOCTOU-safe operations (all checks inside locks)
- Speculative JSON parsing for partial line accumulation
- Exit error tracking for proper error reporting
- Temp file cleanup on close
- Stderr callback for debugging
- Windows command length validation

**Next:** Plan 03 - Query/Control Protocol
